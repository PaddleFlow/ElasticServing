package istio

import (
	"ElasticServing/pkg/controllers/elasticserving/resources/istio"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("VirtualServiceReconciler")

type VirtualServiceReconciler struct {
	client         client.Client
	scheme         *runtime.Scheme
	serviceBuilder *istio.VirtualServiceBuilder
}

func (r *VirtualServiceReconciler) Reconcile(paddlesvc *elasticservingv1.PaddleService) error {

	return nil
}
