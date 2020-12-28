package knative

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"

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

func (r *ServiceBuilder) CreateService(paddlesvc *elasticservingv1.PaddleService) (*knservingv1.Service, error) {
	serviceName := paddlesvc.ObjectMeta.Name + "-knativeSvc"
	service := &knservingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: paddlesvc.ObjectMeta.Namespace,
			Labels:    paddlesvc.ObjectMeta.Labels,
		},
		Spec: knservingv1.ServiceSpec{
			ConfigurationSpec: knservingv1.ConfigurationSpec{
				Template: knservingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"PaddleService": paddlesvc.ObjectMeta.Name,
						},
					},
					Spec: knservingv1.RevisionSpec{
						PodSpec: core.PodSpec{
							Containers: []core.Container{
								{
									Image: paddlesvc.Spec.StorageURI,
								},
							},
						},
					},
				},
			},
		},
	}
	return service, nil
}
