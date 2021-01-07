package knative

import (
	elasticservingv1 "ElasticServing/pkg/apis/elasticserving/v1"
	"ElasticServing/pkg/constants"
	"testing"

	"github.com/google/go-cmp/cmp"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	image                      = "hub.baidubce.com/paddlepaddle/serving:latest"
	port                       = 9292
	ActualTestServiceName      = "test-service"
	PaddleServiceDefaultCPU    = "0.1"
	PaddleServiceDefaultMemory = "128Mi"
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
				},
				Spec: knservingv1.RevisionSpec{
					PodSpec: core.PodSpec{
						Containers: []core.Container{
							{
								ImagePullPolicy: core.PullAlways,
								Name:            paddlesvc.Spec.RuntimeVersion,
								Image:           image,
								Ports: []core.ContainerPort{
									{ContainerPort: port, Name: "http1", Protocol: "TCP"},
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
