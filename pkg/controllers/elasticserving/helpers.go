package controllers

import (
	"context"

	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *PaddleServiceReconciler) cleanupOwnedResources(ctx context.Context, log logr.Logger, paddlesvc *elasticservingv1.PaddleService) error {
	log.Info("finding existing Deployments for paddlesvc resource")

	// List all deployment resources owned by this paddlesvc
	var deployments apps.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(paddlesvc.Namespace)); err != nil {
		return err
	}

	deleted := 0
	for _, depl := range deployments.Items {
		if depl.Name == paddlesvc.Spec.DeploymentName {
			// If this deployment's name matches the one on the paddlesvc resource
			// then do not delete it.
			continue
		}

		if err := r.Client.Delete(ctx, &depl); err != nil {
			log.Error(err, "failed to delete Deployment resource")
			return err
		}

		r.Recorder.Eventf(paddlesvc, core.EventTypeNormal, "Deleted", "Deleted deployment %q", depl.Name)
		deleted++
	}

	log.Info("finished cleaning up old Deployment resources", "number_deleted", deleted)

	var services core.ServiceList

	deleted = 0
	for _, svc := range services.Items {
		if svc.Name == paddlesvc.Spec.DeploymentName {
			continue
		}

		if err := r.Client.Delete(ctx, &svc); err != nil {
			log.Error(err, "dailed to deleted Service resource")
			return err
		}

		r.Recorder.Eventf(paddlesvc, core.EventTypeNormal, "Deleted", "Deleted service %q", svc.Name)
		deleted++
	}

	log.Info("finished cleaning up old Service resources", "number_deleted", deleted)

	return nil
}

func buildService(paddlesvc elasticservingv1.PaddleService) *core.Service {
	name := paddlesvc.Spec.DeploymentName
	service := core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: paddlesvc.Namespace,
		},
		Spec: core.ServiceSpec{
			Selector: map[string]string{
				"elastic-serving.paddlepaddle.org/deployment-name": paddlesvc.Spec.DeploymentName,
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

func buildDeployment(paddlesvc elasticservingv1.PaddleService) *apps.Deployment {
	deployment := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            paddlesvc.Spec.DeploymentName,
			Namespace:       paddlesvc.ObjectMeta.Namespace,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&paddlesvc, elasticservingv1.GroupVersion.WithKind("paddlesvc"))},
		},
		Spec: apps.DeploymentSpec{
			Replicas: paddlesvc.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"elastic-serving.paddlepaddle.org/deployment-name": paddlesvc.Spec.DeploymentName,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"elastic-serving.paddlepaddle.org/deployment-name": paddlesvc.Spec.DeploymentName,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  paddlesvc.Spec.RuntimeVersion,
							Image: paddlesvc.Spec.StorageURI,
							Ports: []core.ContainerPort{
								{ContainerPort: paddlesvc.Spec.Port, Name: "http", Protocol: "TCP"},
							},
						},
					},
				},
			},
		},
	}
	return &deployment
}
