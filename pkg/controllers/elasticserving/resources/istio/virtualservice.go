package istio

import (
	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
	"fmt"
	"os"

	istiov1alpha3 "istio.io/api/networking/v1alpha3"
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
	clusterDomain := "cluster.local"
	prefix := fmt.Sprintf("/paddlepaddle/%s/%s/", paddlesvc.Namespace, paddlesvc.Name)
	// rewrite := fmt.Sprintf("/paddlepaddle/%s/%s/", paddlesvc.Namespace, paddlesvc.Name)

	if clusterDomainFromEnv, ok := os.LookupEnv("CLUSTER_DOMAIN"); ok {
		clusterDomain = clusterDomainFromEnv
	}
	service := fmt.Sprintf("%s.%s.svc.%s", paddlesvc.Name, paddlesvc.Namespace, clusterDomain)

	istioGateway := os.Getenv("ISTIO_GATEWAY")
	if len(istioGateway) == 0 {
		istioGateway = "paddleflow/paddleflow-gateway"
	}

	vs := v1alpha3.VirtualService{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha3.SchemeGroupVersion.String(),
			Kind:       "VirtualService",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      paddlesvc.Name,
			Namespace: paddlesvc.Namespace,
			// Labels:      paddlesvc.Labels,
			// Annotations: paddlesvc.Annotations,
		},
		Spec: istiov1alpha3.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{istioGateway},
			Http: []*istiov1alpha3.HTTPRoute{
				{
					Match: []*istiov1alpha3.HTTPMatchRequest{
						{
							Uri: &istiov1alpha3.StringMatch{
								MatchType: &istiov1alpha3.StringMatch_Prefix{
									Prefix: prefix,
								},
							},
						},
					},
					Route: []*istiov1alpha3.HTTPRouteDestination{
						&istiov1alpha3.HTTPRouteDestination{
							Destination: &istiov1alpha3.Destination{
								Host: service,
								Port: &istiov1alpha3.PortSelector{
									Number: uint32(80),
								},
							},
						},
					},
				},
			},
		},
	}

	return vs
}
