package knative

import (
	"ElasticServing/pkg/constants"
	"ElasticServing/pkg/controllers/elasticserving/resources/knative"
	"context"
	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/kmp"
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

func NewServiceReconciler(client client.Client, scheme *runtime.Scheme, configMap *core.ConfigMap) *ServiceReconciler {
	return &ServiceReconciler{
		client:         client,
		scheme:         scheme,
		serviceBuilder: knative.NewServiceBuilder(configMap),
	}
}

// +kubebuilder:rbac:groups=serving.knative.dev,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=serving.knative.dev,resources=services/status,verbs=get;update;patch
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
	var err error
	serviceName := constants.DefaultServiceName(paddlesvc.Name)
	service, err = r.serviceBuilder.CreateService(serviceName, paddlesvc)
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

	if _, err := r.reconcileService(paddlesvc, service); err != nil {
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

func (r *ServiceReconciler) reconcileService(paddlesvc *elasticservingv1.PaddleService, desired *knservingv1.Service) (*knservingv1.ServiceStatus, error) {
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

	// Return if no differences to reconcile.
	if knativeServiceSemanticEquals(desired, existing) {
		log.Info("No differences found")
		return &existing.Status, nil
	}

	// Reconcile differences and update
	diff, err := kmp.SafeDiff(desired.Spec.ConfigurationSpec, existing.Spec.ConfigurationSpec)
	if err != nil {
		return &existing.Status, fmt.Errorf("Failed to diff Knative Service: %v", err)
	}

	log.Info("Reconciling Knative Service diff (-desired, +observed):", "diff", diff)
	log.Info("Updating Knative Service", "namespace", desired.Namespace, "name", desired.Name)

	existing.Spec.ConfigurationSpec = desired.Spec.ConfigurationSpec
	existing.ObjectMeta.Labels = desired.ObjectMeta.Labels
	if err := r.client.Update(context.TODO(), existing); err != nil {
		return &existing.Status, err
	}

	return &existing.Status, nil
}

func knativeServiceSemanticEquals(desired, service *knservingv1.Service) bool {
	return equality.Semantic.DeepEqual(desired.Spec.ConfigurationSpec, service.Spec.ConfigurationSpec) &&
		equality.Semantic.DeepEqual(desired.ObjectMeta.Labels, service.ObjectMeta.Labels)
}
