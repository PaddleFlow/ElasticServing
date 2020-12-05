package controllers

import (
	"context"

	elasticservingv1 "ElasticServing/api/v1"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
