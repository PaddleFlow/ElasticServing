package knative

import (
	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
)

type ServiceConfig struct {
	ContainerImage string `json:"containerImage,omitempty"`
	Port           int32  `json:"port,omitempty"`
}

type ServiceBuilder struct {
	serviceConfig *ServiceConfig
}

func NewServiceBuilder(paddlesvc *elasticservingv1.PaddleService) *ServiceBuilder {
	serviceConfig := &ServiceConfig{}
	serviceConfig.ContainerImage = paddlesvc.Spec.StorageURI
	serviceConfig.Port = paddlesvc.Spec.Port

	return &ServiceBuilder{serviceConfig: serviceConfig}
}
