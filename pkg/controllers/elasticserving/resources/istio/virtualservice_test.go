package istio

import (
	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
	"testing"

	"github.com/google/go-cmp/cmp"

	istiov1alpha3 "istio.io/api/networking/v1alpha3"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateVirtualService(t *testing.T) {
	serviceName := "testresource"
	namespace := "test"
	prefix := "/paddlepaddle/test/testresource/"
	rewrite := "/paddlepaddle/test/testresource/"
	service := "testresource.test.svc.cluster.local"
	host := "*"
	istioGateway := "paddleflow/paddleflow-gateway"

	cases := []struct {
		name       string
		expectedVs *v1alpha3.VirtualService
	}{{
		name: "test case 1",
		expectedVs: &v1alpha3.VirtualService{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha3.SchemeGroupVersion.String(),
				Kind:       "VirtualService",
			},
			ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: namespace},
			Spec: istiov1alpha3.VirtualService{
				Hosts:    []string{host},
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
						Rewrite: &istiov1alpha3.HTTPRewrite{
							Uri: rewrite,
						},
					},
				},
			},
		},
	}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			paddlesvc := &elasticservingv1.PaddleService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testresource",
					Namespace: "test",
				},
				Spec: elasticservingv1.PaddleServiceSpec{
					DeploymentName: "deployment-name",
					RuntimeVersion: "paddlesvc",
					StorageURI:     "hub.baidubce.com/paddlepaddle/serving:latest",
					Port:           9292,
				},
			}

			serviceBuilder := VirtualServiceBuilder{
				ingressConfig: &IngressConfig{
					IngressGateway:     "paddleflow/paddleflow-gateway",
					IngressServiceName: "*",
				},
			}

			createdVs := serviceBuilder.CreateVirtualService(paddlesvc)

			if diff := cmp.Diff(tc.expectedVs, createdVs); diff != "" {
				t.Errorf("Test %q unexpected service (-want +got): %v", tc.name, diff)
			}
		})
	}
}
