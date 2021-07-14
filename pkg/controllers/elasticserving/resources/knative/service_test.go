package knative

import (
	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
	"ElasticServing/pkg/constants"
	"testing"

	"github.com/google/go-cmp/cmp"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	image                      = "hub.baidubce.com/paddlepaddle/serving:latest"
	port                       = 9292
	ActualTestServiceName      = "test-service"
	PaddleServiceDefaultCPU    = "0.1"
	PaddleServiceDefaultMemory = "128Mi"
)

var (
	command              = []string{"/bin/bash", "-c"}
	args                 = []string{""}
	containerConcurrency = int64(0)
	timeoutSeconds       = int64(300)

	readinessInitialDelaySeconds = 60
	readinessFailureThreshold    = 3
	readinessPeriodSeconds       = 10
	readinessTimeoutSeconds      = 180
	livenessInitialDelaySeconds  = 60
	livenessFailureThreshold     = 3
	livenessPeriodSeconds        = 10
)

var defaultResources = core.ResourceList{
	core.ResourceCPU:    resource.MustParse(PaddleServiceDefaultCPU),
	core.ResourceMemory: resource.MustParse(PaddleServiceDefaultMemory),
}

var configMapData = map[string]string{
	"paddleService": `{
		"containerImage": "hub.baidubce.com/paddlepaddle/serving",
		"version": "latest",
        "port": 9292
	}`,
}

var annotations = map[string]string{
	"autoscaling.knative.dev/class":                       "kpa.autoscaling.knative.dev",
	"autoscaling.knative.dev/maxScale":                    "10",
	"autoscaling.knative.dev/metric":                      "concurrency",
	"autoscaling.knative.dev/minScale":                    "1",
	"autoscaling.knative.dev/panicThresholdPercentage":    "200",
	"autoscaling.knative.dev/panicWindowPercentage":       "10",
	"autoscaling.knative.dev/target":                      "100",
	"autoscaling.knative.dev/targetUtilizationPercentage": "70",
	"autoscaling.knative.dev/window":                      "60s",
}

var paddlesvc = elasticservingv1.PaddleService{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "paddlesvc",
		Namespace: "default",
	},
	Spec: elasticservingv1.PaddleServiceSpec{
		DeploymentName: "depl-test",
		RuntimeVersion: "latest",
		Resources: core.ResourceRequirements{
			Requests: defaultResources,
			Limits:   defaultResources,
		},
	},
}

var defaultService = &knservingv1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      constants.DefaultServiceName("test"),
		Namespace: "default",
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
					ContainerConcurrency: &containerConcurrency,
					TimeoutSeconds:       &timeoutSeconds,
					PodSpec: core.PodSpec{
						Containers: []core.Container{
							{
								ImagePullPolicy: core.PullAlways,
								Name:            paddlesvc.Spec.RuntimeVersion,
								Image:           image,
								Ports: []core.ContainerPort{
									{ContainerPort: port, Name: "http1", Protocol: "TCP"},
								},
								Command: command,
								Args:    args,
								ReadinessProbe: &core.Probe{
									InitialDelaySeconds: int32(readinessInitialDelaySeconds),
									FailureThreshold:    int32(readinessFailureThreshold),
									PeriodSeconds:       int32(readinessPeriodSeconds),
									TimeoutSeconds:      int32(readinessTimeoutSeconds),
									Handler: core.Handler{
										TCPSocket: &core.TCPSocketAction{
											Port: intstr.FromInt(0),
										},
									},
								},
								LivenessProbe: &core.Probe{
									InitialDelaySeconds: int32(livenessInitialDelaySeconds),
									FailureThreshold:    int32(livenessFailureThreshold),
									PeriodSeconds:       int32(livenessPeriodSeconds),
									Handler: core.Handler{
										TCPSocket: &core.TCPSocketAction{
											Port: intstr.FromInt(0),
										},
									},
								},
								Resources: paddlesvc.Spec.Resources,
							},
						},
					},
				},
			},
		},
	},
}

func TestPaddleServiceToKnativeService(t *testing.T) {
	scenarios := map[string]struct {
		paddleService   elasticservingv1.PaddleService
		expectedDefault *knservingv1.Service
	}{
		"Test1": {
			paddleService:   paddlesvc,
			expectedDefault: defaultService,
		},
	}
	serviceBuilder := NewServiceBuilder(&core.ConfigMap{
		Data: configMapData,
	})

	for name, scenario := range scenarios {
		actualDefaultService, err := serviceBuilder.CreateService(ActualTestServiceName, &paddlesvc)
		if err != nil {
			t.Errorf("Test %q unexpected error %s", name, err.Error())
		}
		if diff := cmp.Diff(scenario.expectedDefault, actualDefaultService); diff != "" {
			t.Errorf("Test %q unexpected canary service (-want +got): %v", name, diff)
		}
	}
}
