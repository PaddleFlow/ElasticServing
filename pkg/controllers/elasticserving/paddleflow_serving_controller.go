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

	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
	"ElasticServing/pkg/controllers/elasticserving/reconcilers/istio"
	"ElasticServing/pkg/controllers/elasticserving/reconcilers/knative"

	"github.com/go-logr/logr"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// PaddleServiceReconciler reconciles a PaddleService object
type PaddleServiceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=elasticserving.paddlepaddle.org,resources=paddlesvcs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elasticserving.paddlepaddle.org,resources=paddlesvcs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *PaddleServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("paddlesvc", req.NamespacedName)

	// your logic here
	log.Info("reconciling paddlesvc")

	// Load the PaddleService by name
	var paddlesvc elasticservingv1.PaddleService
	if err := r.Get(ctx, req.NamespacedName, &paddlesvc); err != nil {
		log.Error(err, "unable to fetch PaddleService")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("Successfully fetching paddlesvc")

	if err := r.cleanupOwnedResources(ctx, log, &paddlesvc); err != nil {
		log.Error(err, "failed to clean up old Deployment resources for this PaddleService")
		return ctrl.Result{}, err
	}

	log.Info("Successfully cleaning up resources")

	log.Info("Creating Deployment")
	log = log.WithValues("deployment_name", paddlesvc.Spec.DeploymentName)
	log.Info("Checking if an existing Deployment exists for this resource")
	deployment := apps.Deployment{}
	service := core.Service{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: paddlesvc.Namespace, Name: paddlesvc.Spec.DeploymentName}, &deployment)
	if apierrors.IsNotFound(err) {
		log.Info("Could not find existing Deployment for PaddleService, creating one...")

		deployment = *buildDeployment(paddlesvc)
		log.Info("Successfully building the deployment")

		if err := r.Client.Create(ctx, &deployment); err != nil {
			log.Error(err, "failed to create Deployment resource")
			return ctrl.Result{}, err
		}

		log.Info("Deployment created successfully")

		r.Recorder.Eventf(&paddlesvc, core.EventTypeNormal, "Created", "Created deployment %q", deployment.Name)
		log.Info("Created Deployment resource for PaddleService")

		service = *buildService(paddlesvc)
		log.Info("Successfully building the service")

		if err := r.Client.Create(ctx, &service); err != nil {
			log.Error(err, "failed to create Service resource")
			return ctrl.Result{}, err
		}

		log.Info("Service created successfully")

		r.Recorder.Eventf(&paddlesvc, core.EventTypeNormal, "Created", "Created service %q", deployment.Name)
		log.Info("Created service resource for PaddleService")

		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "failed to get Deployment or Service for PaddleService resource")
		return ctrl.Result{}, err
	}

	log.Info("Existing deployment resource already exists for PaddleService, checking replica count")

	expectedReplicas := int32(1)
	if paddlesvc.Spec.Replicas != nil {
		expectedReplicas = *paddlesvc.Spec.Replicas
	}
	if *deployment.Spec.Replicas != expectedReplicas {
		log.Info("updating replica count", "old_count", *deployment.Spec.Replicas, "new_count", expectedReplicas)

		deployment.Spec.Replicas = &expectedReplicas
		if err := r.Client.Update(ctx, &deployment); err != nil {
			log.Error(err, "failed to Deployment update replica count")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&paddlesvc, core.EventTypeNormal, "Scaled", "Scaled deployment %q to %d replicas", deployment.Name, expectedReplicas)

		return ctrl.Result{}, nil
	}

	log.Info("replica count up to date", "replica_count", *deployment.Spec.Replicas)

	log.Info("updating PaddleService resource status")
	paddlesvc.Status.Replicas = deployment.Status.ReadyReplicas
	if r.Client.Status().Update(ctx, &paddlesvc); err != nil {
		log.Error(err, "failed to update PaddleService status")
		return ctrl.Result{}, err
	}

	istioReconciler := istio.NewVirtualServiceReconciler(r.Client, r.Scheme)

	if err := istioReconciler.Reconcile(&paddlesvc); err != nil {
		r.Log.Error(err, "Failed to finish istio reconcile")
		r.Recorder.Eventf(&paddlesvc, core.EventTypeWarning, "InternalError", err.Error())
		return reconcile.Result{}, err
	}

	serviceReconciler := knative.NewServiceReconciler(r.Client, r.Scheme, &paddlesvc)

	if err := serviceReconciler.Reconcile(&paddlesvc); err != nil {
		r.Log.Error(err, "Failed to finish knative reconcile")
		r.Recorder.Eventf(&paddlesvc, core.EventTypeWarning, "InternalError", err.Error())
		return reconcile.Result{}, err
	}

	log.Info("resource status synced")

	return ctrl.Result{}, nil
}

func (r *PaddleServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&elasticservingv1.PaddleService{}).
		Owns(&apps.Deployment{}).
		Owns(&v1alpha3.VirtualService{}).
		Owns(&knservingv1.Service{}).
		Complete(r)
}
