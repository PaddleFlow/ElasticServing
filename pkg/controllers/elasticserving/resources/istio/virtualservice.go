package istio

import (
	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressConfig struct {
	IngressGateway     string `json:"ingressGateway,omitempty"`
	IngressServiceName string `json:"ingressService,omitempty"`
}

type VirtualServiceBuilder struct {
	ingressConfig *IngressConfig
}

func (r *VirtualServiceBuilder) CreateVirtualService(paddlesvc *elasticservingv1.PaddleService) v1alpha3.VirtualService {
	vs := v1alpha3.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha3.SchemeGroupVersion.String(),
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        paddlesvc.Name,
			Namespace:   paddlesvc.Namespace,
			Labels:      paddlesvc.Labels,
			Annotations: paddlesvc.Annotations,
		},
	}

	return vs
}
