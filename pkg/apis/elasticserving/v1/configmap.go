package v1

import (
	"encoding/json"

	core "k8s.io/api/core/v1"

	"ElasticServing/pkg/constants"
)

type PaddleServiceConfig struct {
	ContainerImage string `json:"containerImage"`
	Version        string `json:"version"`
	Port           int32  `json:"port"`
}

func NewPaddleServiceConfig(configMap *core.ConfigMap) (*PaddleServiceConfig, error) {
	paddleServiceConfig := PaddleServiceConfig{}
	key := constants.PaddleService

	if data, ok := configMap.Data[key]; ok {
		err := json.Unmarshal([]byte(data), &paddleServiceConfig)
		if err != nil {
			return nil, err
		}
	}

	return &paddleServiceConfig, nil
}

type IngressConfig struct {
	IngressGateway     string `json:"ingressGateway"`
	IngressServiceName string `json:"ingressServiceName"`
}

func NewIngressConfig(configMap *core.ConfigMap) (*IngressConfig, error) {
	ingressConfig := IngressConfig{}
	key := constants.Ingress

	if data, ok := configMap.Data[key]; ok {
		err := json.Unmarshal([]byte(data), &ingressConfig)
		if err != nil {
			return nil, err
		}
	}

	return &ingressConfig, nil
}
