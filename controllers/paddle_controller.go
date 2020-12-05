/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	elasticservingv1 "ElasticServing/api/v1"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PaddleReconciler reconciles a Paddle object
type PaddleReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=elasticserving.paddlepaddle.org,resources=paddles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elasticserving.paddlepaddle.org,resources=paddles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=elasticserving.paddlepaddle.org,resources=paddles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *PaddleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("paddle", req.NamespacedName)

	// your logic here
	log.Info("reconciling paddle")

	// Load the Paddle by name
	var paddle elasticservingv1.Paddle
	if err := r.Get(ctx, req.NamespacedName, &paddle); err != nil {
		log.Error(err, "unable to fetch Paddle")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Successfully fetching paddle")

	if err := r.cleanupOwnedResources(ctx, log, &paddle); err != nil {
		log.Error(err, "failed to clean up old Deployment resources for this Paddle")
		return ctrl.Result{}, err
	}

	log.Info("Successfully cleaning up resources")

	log.Info("Creating Deployment")
	log = log.WithValues("deployment_name", paddle.Spec.DeploymentName)
	log.Info("Checking if an existing Deployment exists for this resource")
	deployment := apps.Deployment{}
	service := core.Service{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: paddle.Namespace, Name: paddle.Spec.DeploymentName}, &deployment)
	if apierrors.IsNotFound(err) {
		log.Info("Could not find existing Deployment for Paddle, creating one...")

		deployment = *buildDeployment(paddle)
		log.Info("Successfully building the deployment")

		if err := r.Client.Create(ctx, &deployment); err != nil {
			log.Error(err, "failed to create Deployment resource")
			return ctrl.Result{}, err
		}

		log.Info("Deployment created successfully")

		r.Recorder.Eventf(&paddle, core.EventTypeNormal, "Created", "Created deployment %q", deployment.Name)
		log.Info("Created Deployment resource for Paddle")

		service = *buildService(paddle)
		log.Info("Successfully building the service")

		if err := r.Client.Create(ctx, &service); err != nil {
			log.Error(err, "failed to create Service resource")
			return ctrl.Result{}, err
		}

		log.Info("Service created successfully")

		r.Recorder.Eventf(&paddle, core.EventTypeNormal, "Created", "Created service %q", deployment.Name)
		log.Info("Created service resource for Paddle")

		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "failed to get Deployment or Service for Paddle resource")
		return ctrl.Result{}, err
	}

	log.Info("Existing deployment resource already exists for Paddle, checking replica count")

	expectedReplicas := int32(1)
	if paddle.Spec.Replicas != nil {
		expectedReplicas = *paddle.Spec.Replicas
	}
	if *deployment.Spec.Replicas != expectedReplicas {
		log.Info("updating replica count", "old_count", *deployment.Spec.Replicas, "new_count", expectedReplicas)

		deployment.Spec.Replicas = &expectedReplicas
		if err := r.Client.Update(ctx, &deployment); err != nil {
			log.Error(err, "failed to Deployment update replica count")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&paddle, core.EventTypeNormal, "Scaled", "Scaled deployment %q to %d replicas", deployment.Name, expectedReplicas)

		return ctrl.Result{}, nil
	}

	log.Info("replica count up to date", "replica_count", *deployment.Spec.Replicas)

	log.Info("updating Paddle resource status")
	paddle.Status.Replicas = deployment.Status.ReadyReplicas
	if r.Client.Status().Update(ctx, &paddle); err != nil {
		log.Error(err, "failed to update Paddle status")
		return ctrl.Result{}, err
	}

	log.Info("resource status synced")

	return ctrl.Result{}, nil
}

func (r *PaddleReconciler) cleanupOwnedResources(ctx context.Context, log logr.Logger, paddle *elasticservingv1.Paddle) error {
	log.Info("finding existing Deployments for paddle resource")

	// List all deployment resources owned by this paddle
	var deployments apps.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(paddle.Namespace), client.MatchingField(deploymentOwnerKey, paddle.Name)); err != nil {
		return err
	}

	deleted := 0
	for _, depl := range deployments.Items {
		if depl.Name == paddle.Spec.DeploymentName {
			// If this deployment's name matches the one on the paddle resource
			// then do not delete it.
			continue
		}

		if err := r.Client.Delete(ctx, &depl); err != nil {
			log.Error(err, "failed to delete Deployment resource")
			return err
		}

		r.Recorder.Eventf(paddle, core.EventTypeNormal, "Deleted", "Deleted deployment %q", depl.Name)
		deleted++
	}

	log.Info("finished cleaning up old Deployment resources", "number_deleted", deleted)

	log.Info("finding existing Deployments for paddle resource")

	// List all deployment resources owned by this paddle
	// var services core.ServiceList
	// if err := r.List(ctx, &services, client.InNamespace(paddle.Namespace), client.MatchingField(deploymentOwnerKey, paddle.Name)); err != nil {
	// 	return err
	// }

	// deleted = 0
	// for _, svc := range services.Items {
	// 	if svc.Name == paddle.Spec.DeploymentName {
	// 		// If this deployment's name matches the one on the paddle resource
	// 		// then do not delete it.
	// 		continue
	// 	}

	// 	if err := r.Client.Delete(ctx, &svc); err != nil {
	// 		log.Error(err, "failed to delete Deployment resource")
	// 		return err
	// 	}

	// 	r.Recorder.Eventf(paddle, core.EventTypeNormal, "Deleted", "Deleted deployment %q", svc.Name)
	// 	deleted++
	// }

	// log.Info("finished cleaning up old Deployment resources", "number_deleted", deleted)
	return nil
}

func buildService(paddle elasticservingv1.Paddle) *core.Service {
	name := paddle.Spec.DeploymentName
	service := core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: paddle.Namespace,
		},
		Spec: core.ServiceSpec{
			Selector: map[string]string{
				"elastic-serving.paddlepaddle.org/deployment-name": paddle.Spec.DeploymentName,
			},
			Type: "LoadBalancer",
			Ports: []core.ServicePort{
				{
					Port: 80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 80,
					},
				},
			},
		},
	}
	return &service
}

func buildDeployment(paddle elasticservingv1.Paddle) *apps.Deployment {
	deployment := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            paddle.Spec.DeploymentName,
			Namespace:       paddle.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&paddle, elasticservingv1.GroupVersion.WithKind("paddle"))},
		},
		Spec: apps.DeploymentSpec{
			Replicas: paddle.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"elastic-serving.paddlepaddle.org/deployment-name": paddle.Spec.DeploymentName,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"elastic-serving.paddlepaddle.org/deployment-name": paddle.Spec.DeploymentName,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  paddle.Spec.RuntimeVersion,
							Image: paddle.Spec.StorageURI,
						},
					},
				},
			},
		},
	}
	return &deployment
}

var (
	deploymentOwnerKey = ".metadata.controller"
)

// func (r *PaddleReconciler) SetupWithManager(mgr ctrl.Manager) error {
// 	return ctrl.NewControllerManagedBy(mgr).
// 		For(&elasticservingv1.Paddle{}).
// 		Complete(r)
// }

func (r *PaddleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&apps.Deployment{}, deploymentOwnerKey, func(rawObj runtime.Object) []string {
		// grab the Deployment object, extract the owner...
		depl := rawObj.(*apps.Deployment)
		owner := metav1.GetControllerOf(depl)
		if owner == nil {
			return nil
		}
		// ...make sure it's a MyKind...
		if owner.APIVersion != elasticservingv1.GroupVersion.String() || owner.Kind != "paddle" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticservingv1.Paddle{}).
		Owns(&apps.Deployment{}).
		Complete(r)
}
