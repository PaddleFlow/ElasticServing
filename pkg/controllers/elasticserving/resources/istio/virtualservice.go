package istio

type IngressConfig struct {
	IngressGateway     string `json:"ingressGateway,omitempty"`
	IngressServiceName string `json:"ingressService,omitempty"`
}

type VirtualServiceBuilder struct {
	ingressConfig *IngressConfig
}
