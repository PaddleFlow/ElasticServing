package istio

import (
	corev1 "k8s.io/api/core/v1"
)

type IngressConfig struct {
	IngressGateway     string `json:"ingressGateway,omitempty"`
	IngressServiceName string `json:"ingressService,omitempty"`
}

type VirtualServiceBuilder struct {
	ingressConfig *IngressConfig
}

func NewVirtualServiceBuilder(config *corev1.ConfigMap) *VirtualSrviceBuilder {

}
