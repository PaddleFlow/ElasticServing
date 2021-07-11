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

	if err := loadFromConfigMap(configMap, &paddleServiceConfig, key); err != nil {
		return nil, err
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

	if err := loadFromConfigMap(configMap, &ingressConfig, key); err != nil {
		return nil, err
	}

	return &ingressConfig, nil
}

func loadFromConfigMap(configMap *core.ConfigMap, config interface{}, key string) error {
	if data, ok := configMap.Data[key]; ok {
		return json.Unmarshal([]byte(data), config)
	}
	return nil
}
