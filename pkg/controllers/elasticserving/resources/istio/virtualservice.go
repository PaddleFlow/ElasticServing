package istio

import (
	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
	"ElasticServing/pkg/constants"
	"fmt"

	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressConfig struct {
	IngressGateway     string `json:"ingressGateway,omitempty"`
	IngressServiceName string `json:"ingressService,omitempty"`
}

type VirtualServiceBuilder struct {
	ingressConfig *IngressConfig
}

func NewVirtualServiceBuilder(configMap *core.ConfigMap) *VirtualServiceBuilder {
	ingressConfig := &IngressConfig{}

	istioIngressConfig, err := elasticservingv1.NewIngressConfig(configMap)
	if err != nil {
		fmt.Printf("Failed to get paddle service config %s", err)
		panic("Failed to get paddle service config")
	}

	ingressConfig.IngressGateway = istioIngressConfig.IngressGateway
	ingressConfig.IngressServiceName = istioIngressConfig.IngressServiceName

	if ingressConfig.IngressGateway == "" || ingressConfig.IngressServiceName == "" {
		panic(fmt.Errorf("invalid ingress config, ingressGateway and ingressService are required"))
	}

	return &VirtualServiceBuilder{ingressConfig: ingressConfig}
}

func (r *VirtualServiceBuilder) CreateVirtualService(paddlesvc *elasticservingv1.PaddleService) *v1alpha3.VirtualService {

	service := constants.DefaultServiceName(paddlesvc.Name)

	istioGateway := r.ingressConfig.IngressGateway
	host := r.ingressConfig.IngressServiceName

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
		Spec: istiov1alpha3.VirtualService{
			Hosts:    []string{host},
			Gateways: []string{istioGateway},
			Http: []*istiov1alpha3.HTTPRoute{
				{
					Route: []*istiov1alpha3.HTTPRouteDestination{
						{
							Destination: &istiov1alpha3.Destination{
								Host: service,
							},
						},
					},
				},
			},
		},
	}

	return &vs
}
