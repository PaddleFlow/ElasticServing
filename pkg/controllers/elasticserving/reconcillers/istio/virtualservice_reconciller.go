package istio

import (
	"ElasticServing/pkg/controllers/paddleserice/resources/istio"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
)

var log = logf.Log.WithName("VirtualServiceReconciler")

type VirtualServiceReconciler struct {
	client         client.Client
	scheme         *runtime.Scheme
	serviceBuilder *istio.VirtualServiceBuilder
}

func NewVirtualServiceReconciler(client client.Client, scheme *runtime.Scheme, config *corev1.ConfigMap) *VirtualServiceReconciler {
	return &VirtualServiceReconciler{
		client:         client,
		scheme:         scheme,
		serviceBuilder: istio.NewVirtualServiceBuilder(config),
	}
}

func (r *VirtualServiceReconciler) Reconcile(paddlesvc *elasticservingv1.PaddleService) error {

	return nil
}
