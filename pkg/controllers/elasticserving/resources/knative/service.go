package knative

import (
	"fmt"
	"strconv"

	"ElasticServing/pkg/constants"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/serving/pkg/apis/autoscaling"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"

	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
)

type EndpointConfig struct {
	Image    string `json:"image,omitempty"`
	Port     int32  `json:"port,omitempty"`
	Argument string `json:"arg,omitempty"`
}

type ServiceBuilder struct {
	defaultEndpointConfig *EndpointConfig
	canaryEndpointConfig  *EndpointConfig
}

func NewServiceBuilder(paddlesvc *elasticservingv1.PaddleService) *ServiceBuilder {
	defaultEndpointConfig := &EndpointConfig{}
	defaultEndpointConfig.Image = paddlesvc.Spec.Default.ContainerImage + ":" + paddlesvc.Spec.Default.Tag
	defaultEndpointConfig.Port = paddlesvc.Spec.Default.Port
	defaultEndpointConfig.Argument = paddlesvc.Spec.Default.Argument
	if paddlesvc.Spec.Canary == nil {
		return &ServiceBuilder{
			defaultEndpointConfig: defaultEndpointConfig,
			canaryEndpointConfig:  nil,
		}
	} else {
		canaryEndpointConfig := &EndpointConfig{}
		canaryEndpointConfig.Image = paddlesvc.Spec.Default.ContainerImage + ":" + paddlesvc.Spec.Default.Tag
		canaryEndpointConfig.Port = paddlesvc.Spec.Default.Port
		canaryEndpointConfig.Argument = paddlesvc.Spec.Default.Argument
		return &ServiceBuilder{
			defaultEndpointConfig: defaultEndpointConfig,
			canaryEndpointConfig:  canaryEndpointConfig,
		}
	}
}

func (r *ServiceBuilder) CreateService(serviceName string, paddlesvc *elasticservingv1.PaddleService, isCanary bool) (*knservingv1.Service, error) {
	arg := r.defaultEndpointConfig.Argument
	containerImage := r.defaultEndpointConfig.Image
	containerPort := r.defaultEndpointConfig.Port

	if isCanary && r.canaryEndpointConfig == nil {
		return nil, nil
	} else if isCanary && r.canaryEndpointConfig != nil {
		arg = r.canaryEndpointConfig.Argument
		containerImage = r.canaryEndpointConfig.Image
		containerPort = r.canaryEndpointConfig.Port
	}

	metadata := paddlesvc.ObjectMeta
	paddlesvcSpec := paddlesvc.Spec

	resources, err := r.buildResources(metadata, paddlesvcSpec)
	if err != nil {
		return nil, err
	}

	annotations, err := r.buildAnnotations(metadata, paddlesvcSpec)
	if err != nil {
		return nil, err
	}
	concurrency := int64(paddlesvcSpec.Service.Target)

	command := []string{"/bin/bash", "-c"}
	args := []string{
		arg,
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
						Annotations: annotations,
					},
					Spec: knservingv1.RevisionSpec{
						TimeoutSeconds:       &constants.PaddleServiceDefaultTimeout,
						ContainerConcurrency: &concurrency,
						PodSpec: core.PodSpec{
							Containers: []core.Container{
								{
									ImagePullPolicy: core.PullAlways,
									Name:            paddlesvc.Spec.RuntimeVersion,
									Image:           containerImage,
									Ports: []core.ContainerPort{
										{ContainerPort: containerPort,
											Name:     constants.PaddleServiceDefaultPodName,
											Protocol: core.ProtocolTCP,
										},
									},
									Command: command,
									Args:    args,
									ReadinessProbe: &core.Probe{
										InitialDelaySeconds: constants.ReadinessInitialDelaySeconds,
										FailureThreshold:    constants.ReadinessFailureThreshold,
										PeriodSeconds:       constants.ReadinessPeriodSeconds,
										TimeoutSeconds:      constants.ReadinessTimeoutSeconds,
										Handler: core.Handler{
											TCPSocket: &core.TCPSocketAction{
												Port: intstr.FromInt(0),
											},
										},
									},
									LivenessProbe: &core.Probe{
										InitialDelaySeconds: constants.LivenessInitialDelaySeconds,
										FailureThreshold:    constants.LivenessFailureThreshold,
										PeriodSeconds:       constants.LivenessPeriodSeconds,
										Handler: core.Handler{
											TCPSocket: &core.TCPSocketAction{
												Port: intstr.FromInt(0),
											},
										},
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

func (r *ServiceBuilder) buildAnnotations(metadata metav1.ObjectMeta, paddlesvcSpec elasticservingv1.PaddleServiceSpec) (map[string]string, error) {
	annotations := make(map[string]string)

	// Autoscaler
	if paddlesvcSpec.Service.Autoscaler == "" {
		annotations[autoscaling.ClassAnnotationKey] = constants.PaddleServiceDefaultScalingClass
	} else {
		annotations[autoscaling.ClassAnnotationKey] = string(paddlesvcSpec.Service.Autoscaler)
	}

	// Metric
	if paddlesvcSpec.Service.Metric == "" {
		annotations[autoscaling.MetricAnnotationKey] = constants.PaddleServiceDefaultScalingMetric
	} else {
		annotations[autoscaling.MetricAnnotationKey] = string(paddlesvcSpec.Service.Metric)
	}

	// Target
	if paddlesvcSpec.Service.Target == 0 {
		annotations[autoscaling.TargetAnnotationKey] = fmt.Sprint(constants.PaddleServiceDefaultScalingTarget)
	} else {
		annotations[autoscaling.TargetAnnotationKey] = strconv.Itoa(paddlesvcSpec.Service.Target)
	}

	// Target utilization
	if paddlesvcSpec.Service.TargetUtilization == "" {
		annotations[autoscaling.TargetUtilizationPercentageKey] = constants.PaddleServiceDefaultTargetUtilizationPercentage
	} else {
		annotations[autoscaling.TargetUtilizationPercentageKey] = paddlesvcSpec.Service.TargetUtilization
	}

	// Window
	if paddlesvcSpec.Service.Window == "" {
		annotations[autoscaling.WindowAnnotationKey] = constants.PaddleServiceDefaultWindow
	} else {
		annotations[autoscaling.WindowAnnotationKey] = paddlesvcSpec.Service.Window
	}

	// Panic window
	if paddlesvcSpec.Service.PanicWindow == "" {
		annotations[autoscaling.PanicWindowPercentageAnnotationKey] = constants.PaddleServiceDefaultPanicWindow
	} else {
		annotations[autoscaling.PanicWindowPercentageAnnotationKey] = paddlesvcSpec.Service.PanicWindow
	}

	// Panic threshold
	if paddlesvcSpec.Service.PanicThreshold == "" {
		annotations[autoscaling.PanicThresholdPercentageAnnotationKey] = constants.PaddleServiceDefaultPanicThreshold
	} else {
		annotations[autoscaling.PanicThresholdPercentageAnnotationKey] = paddlesvcSpec.Service.PanicThreshold
	}

	// Min replicas
	if paddlesvcSpec.Service.MinScale == nil {
		annotations[autoscaling.MinScaleAnnotationKey] = fmt.Sprint(constants.PaddleServiceDefaultMinScale)
	} else {
		annotations[autoscaling.MinScaleAnnotationKey] = strconv.Itoa(*paddlesvcSpec.Service.MinScale)
	}

	// Max replicas
	if paddlesvcSpec.Service.MaxScale == 0 {
		annotations[autoscaling.MaxScaleAnnotationKey] = fmt.Sprint(constants.PaddleServiceDefaultMaxScale)
	} else {
		annotations[autoscaling.MaxScaleAnnotationKey] = strconv.Itoa(paddlesvcSpec.Service.MaxScale)
	}

	return annotations, nil

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
