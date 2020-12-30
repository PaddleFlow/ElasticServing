package knative

import (
	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"

	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
	"ElasticServing/pkg/constants"
)

type ServiceConfig struct {
	Image string `json:"image,omitempty"`
	Port  int32  `json:"port,omitempty"`
}

type ServiceBuilder struct {
	serviceConfig *ServiceConfig
}

func NewServiceBuilder(configMap *core.ConfigMap) *ServiceBuilder {
	serviceConfig := &ServiceConfig{}
	paddleServiceConfig, err := elasticservingv1.NewPaddleServiceConfig(configMap)
	if err != nil {
		fmt.Printf("Failed to get paddle service config %s", err)
		panic("Failed to get paddle service config")
	}
	serviceConfig.Image = paddleServiceConfig.ContainerImage + ":" + paddleServiceConfig.Version
	serviceConfig.Port = paddleServiceConfig.Port
	return &ServiceBuilder{serviceConfig: serviceConfig}
}

func (r *ServiceBuilder) CreateService(serviceName string, paddlesvc *elasticservingv1.PaddleService) (*knservingv1.Service, error) {
	metadata := paddlesvc.ObjectMeta
	paddlesvcSpec := paddlesvc.Spec
	resources, err := r.buildResources(metadata, paddlesvcSpec)
	if err != nil {
		return nil, err
	}

	service := &knservingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: paddlesvc.Namespace,
			Labels:    paddlesvc.Labels,
		},
		Spec: knservingv1.ServiceSpec{
			ConfigurationSpec: knservingv1.ConfigurationSpec{
				Template: knservingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"PaddleService": paddlesvc.Name,
						},
					},
					Spec: knservingv1.RevisionSpec{
						PodSpec: core.PodSpec{
							Containers: []core.Container{
								{
									ImagePullPolicy: core.PullAlways,
									Name:            paddlesvc.Spec.RuntimeVersion,
									Image:           r.serviceConfig.Image,
									Ports: []core.ContainerPort{
										{ContainerPort: r.serviceConfig.Port, Name: "http", Protocol: "TCP"},
									},
									Resources: resources,
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

func (r *ServiceBuilder) buildResources(metadata metav1.ObjectMeta, paddlesvcSpec elasticservingv1.PaddleServiceSpec) (core.ResourceRequirements, error) {
	defaultResources := core.ResourceList{
		core.ResourceCPU:    resource.MustParse(constants.PaddleServiceDefaultCPU),
		core.ResourceMemory: resource.MustParse(constants.PaddleServiceDefaultMemory),
	}

	if paddlesvcSpec.Resources.Requests == nil {
		paddlesvcSpec.Resources.Requests = defaultResources
	} else {
		for name, value := range defaultResources {
			if _, ok := paddlesvcSpec.Resources.Requests[name]; !ok {
				paddlesvcSpec.Resources.Requests[name] = value
			}
		}
	}

	if paddlesvcSpec.Resources.Limits == nil {
		paddlesvcSpec.Resources.Limits = defaultResources
	} else {
		for name, value := range defaultResources {
			if _, ok := paddlesvcSpec.Resources.Limits[name]; !ok {
				paddlesvcSpec.Resources.Limits[name] = value
			}
		}
	}

	return paddlesvcSpec.Resources, nil
}
