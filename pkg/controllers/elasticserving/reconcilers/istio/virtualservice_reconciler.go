package istio

import (
	"ElasticServing/pkg/controllers/elasticserving/resources/istio"
	"context"
	"reflect"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
)

var log = logf.Log.WithName("VirtualServiceReconciler")

type VirtualServiceReconciler struct {
	client         client.Client
	scheme         *runtime.Scheme
	serviceBuilder *istio.VirtualServiceBuilder
}

func NewVirtualServiceReconciler(client client.Client, scheme *runtime.Scheme, configMap *core.ConfigMap) *VirtualServiceReconciler {
	return &VirtualServiceReconciler{
		client:         client,
		scheme:         scheme,
		serviceBuilder: istio.NewVirtualServiceBuilder(configMap),
	}
}

// Reconcile reads that state of the cluster for a Profile object and makes changes based on the state read
// and what is in the Profile.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccount,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=list;create;update;delete
// +kubebuilder:rbac:groups=elasticserving.paddlepaddle.org,resources=paddlesvcs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=elasticserving.paddlepaddle.org,resources=paddlesvcs/status,verbs=get;update;patch

func (r *VirtualServiceReconciler) Reconcile(paddlesvc *elasticservingv1.PaddleService) error {

	desiredVs := r.serviceBuilder.CreateVirtualService(paddlesvc)
	if err := ctrl.SetControllerReference(paddlesvc, desiredVs, r.scheme); err != nil {
		return err
	}
	log.Info("Desired Virtual Service created successfully")

	existingVs := &v1alpha3.VirtualService{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: desiredVs.Name, Namespace: desiredVs.Namespace}, existingVs)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Virtual Service", "namespace", desiredVs.Namespace, "name", desiredVs.Name)
			err = r.client.Create(context.TODO(), desiredVs)
		}
		return err
	}
	log.Info("Existing Virtual Service created successfully")

	if err = r.CompAndCopyVs(desiredVs, existingVs); err != nil {
		return err
	}

	return nil
}

func (r *VirtualServiceReconciler) CompAndCopyVs(desiredVs *v1alpha3.VirtualService, existingVs *v1alpha3.VirtualService) error {
	if !reflect.DeepEqual(desiredVs.Spec, existingVs.Spec) {
		log.Info("Reconciling virtual service")
		log.Info("Updating virtual service", "namespace", existingVs.Namespace, "name", existingVs.Name)
		existingVs.Spec = desiredVs.Spec
		existingVs.ObjectMeta.Annotations = desiredVs.ObjectMeta.Annotations
		existingVs.ObjectMeta.Labels = desiredVs.ObjectMeta.Labels
		return r.client.Update(context.TODO(), existingVs)
	}
	return nil
}
