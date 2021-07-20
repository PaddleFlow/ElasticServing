package knative

import (
	"ElasticServing/pkg/constants"
	"ElasticServing/pkg/controllers/elasticserving/resources/knative"
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
)

var log = logf.Log.WithName("ServiceReconciler")

type ServiceReconciler struct {
	client         client.Client
	scheme         *runtime.Scheme
	serviceBuilder *knative.ServiceBuilder
}

func NewServiceReconciler(client client.Client, scheme *runtime.Scheme, paddlesvc *elasticservingv1.PaddleService) *ServiceReconciler {
	return &ServiceReconciler{
		client:         client,
		scheme:         scheme,
		serviceBuilder: knative.NewServiceBuilder(paddlesvc),
	}
}

// +kubebuilder:rbac:groups=serving.knative.dev,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serving.knative.dev,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=serving.knative.dev,resources=revisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serving.knative.dev,resources=revisions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;create;
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;create;
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;create;
// +kubebuilder:rbac:groups="",resources=services,verbs=*
// +kubebuilder:rbac:groups="",resources=pods,verbs=*
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=elasticserving.paddlepaddle.org,resources=paddleservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elasticserving.paddlepaddle.org,resources=paddleservices/status,verbs=get;update;patch

func (r *ServiceReconciler) Reconcile(paddlesvc *elasticservingv1.PaddleService) error {
	var service *knservingv1.Service
	var serviceWithCanary *knservingv1.Service
	var err error
	serviceName := paddlesvc.Name
	service, err = r.serviceBuilder.CreateService(serviceName, paddlesvc, false)
	if err != nil {
		return err
	}

	if service == nil {
		if err = r.finalizeService(serviceName, paddlesvc.Namespace); err != nil {
			return err
		}
		// TODO: Modify status
		// paddlesvc.Status.PropagateStatus(nil)
		return nil
	}

	if _, err := r.reconcileDefaultService(paddlesvc, service); err != nil {
		return err
	} else {
		// TODO: Modify status
		// paddlesvc.Status.PropagateStatus(status)
	}

	serviceWithCanary, err = r.serviceBuilder.CreateService(serviceName, paddlesvc, true)
	if err != nil {
		return err
	}
	if serviceWithCanary == nil {
		return nil
	}

	if _, err := r.reconcileCanaryService(paddlesvc, serviceWithCanary); err != nil {
		return err
	} else {
		// TODO: Modify status
		// paddlesvc.Status.PropagateStatus(status)
	}

	return nil
}

func (r *ServiceReconciler) finalizeService(serviceName, namespace string) error {
	existing := &knservingv1.Service{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: namespace}, existing); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		log.Info("Deleting Knative Service", "namespace", namespace, "name", serviceName)
		if err := r.client.Delete(context.TODO(), existing, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
		}
	}
	return nil
}

// func (r *ServiceReconciler) finalizeCanaryEndpoints(serviceName, namespace string) error {
// 	existing := &knservingv1.Service{}
// 	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: namespace}, existing); err != nil {
// 		if !errors.IsNotFound(err) {
// 			return err
// 		}
// 	} else {
// 		log.Info("Deleting canary endpoint", "namespace", namespace, "name", constants.CanaryServiceName(serviceName))
// 	}
// 	return nil
// }

func (r *ServiceReconciler) reconcileDefaultService(paddlesvc *elasticservingv1.PaddleService, desired *knservingv1.Service) (*knservingv1.ServiceStatus, error) {
	// Set Paddlesvc as owner of desired service
	if err := controllerutil.SetControllerReference(paddlesvc, desired, r.scheme); err != nil {
		return nil, err
	}

	// Create service if does not exist
	existing := &knservingv1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Knative Service", "namespace", desired.Namespace, "name", desired.Name)
			return &desired.Status, r.client.Create(context.TODO(), desired)
		}
		return nil, err
	}

	// existingRevision := &knservingv1.Revision{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: constants.DefaultServiceName(desired.Name), Namespace: desired.Namespace}, existingRevision)
	// if err != nil {
	// 	return nil, err
	// }
	// desiredRevision, err := r.serviceBuilder.CreateRevision(constants.DefaultServiceName(desired.Name), paddlesvc, false)
	// if err != nil {
	// 	return nil, err
	// }

	// if knativeRevisionSemanticEquals(desiredRevision, existingRevision) {
	// 	log.Info("No differences on revision found")
	// 	return &existing.Status, nil
	// }

	// // if !equality.Semantic.DeepDerivative(desiredRevision.ObjectMeta.Annotations, existingRevision.ObjectMeta.Annotations) {
	// // 	existing.Spec.ConfigurationSpec.Template.Annotations = desiredRevision.ObjectMeta.Annotations
	// // 	log.Info("Existing Spec", "Existing Spec", existing.Spec)
	// // 	if err := r.client.Update(context.TODO(), existing); err != nil {
	// // 		return &existing.Status, err
	// // 	}
	// // }

	// existingRevision.ObjectMeta.Annotations = desiredRevision.ObjectMeta.Annotations
	// existingRevision.Spec = desiredRevision.Spec
	// if err := r.client.Update(context.TODO(), existingRevision); err != nil {
	// 	return &existing.Status, err
	// }
	return &existing.Status, nil
}

func (r *ServiceReconciler) reconcileCanaryService(paddlesvc *elasticservingv1.PaddleService, desired *knservingv1.Service) (*knservingv1.ServiceStatus, error) {
	log.Info("IN", "IN", desired)
	// Set Paddlesvc as owner of desired service
	if err := controllerutil.SetControllerReference(paddlesvc, desired, r.scheme); err != nil {
		return nil, err
	}

	existingRevision := &knservingv1.Revision{}
	existing := &knservingv1.Service{}

	//Create canary revision if does not exist
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: constants.CanaryServiceName(desired.Name), Namespace: desired.Namespace}, existingRevision)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Canary Revision", "namespace", desired.Namespace, "name", desired.Name)
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, existing)
			if err != nil {
				return &desired.Status, err
			}

			log.Info("LALALALALALAL", "Desired.Spec", desired.Spec)
			existing.Spec = desired.Spec

			err = r.client.Update(context.TODO(), existing)
			log.Info("Existing after updated", "existing", existing)
			if err != nil {
				return nil, err
			}

			log.Info("At canary stage", "existing", existing)
			return &existing.Status, nil
		}
		return nil, err
	}

	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: constants.CanaryServiceName(desired.Name), Namespace: desired.Namespace}, existingRevision)
	// if err != nil {
	// 	return nil, err
	// }
	// desiredRevision, err := r.serviceBuilder.CreateRevision(constants.CanaryServiceName(desired.Name), paddlesvc, true)
	// if err != nil {
	// 	return nil, err
	// }

	// if knativeRevisionSemanticEquals(desiredRevision, existingRevision) {
	// 	log.Info("No differences on revision found")
	// 	return nil, nil
	// }

	return nil, nil
}

func knativeServiceSemanticEquals(desired, service *knservingv1.Service) bool {
	return equality.Semantic.DeepEqual(desired.Spec.ConfigurationSpec, service.Spec.ConfigurationSpec) &&
		equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, service.ObjectMeta.Labels)
}

func knativeRevisionSemanticEquals(desired, existing *knservingv1.Revision) bool {
	return equality.Semantic.DeepDerivative(desired.ObjectMeta.Annotations, existing.ObjectMeta.Annotations) &&
		equality.Semantic.DeepEqual(desired.Spec, existing.Spec)
}
