/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PaddleServiceSpec defines the desired state of PaddleService
type PaddleServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:MaxLength=64
	DeploymentName string `json:"deploymentName"`
	// Argument for Service
	Argument string `json:"arg,omitempty"`
	// Version of the service
	RuntimeVersion string `json:"runtimeVersion,omitempty"`
	// Defaults to requests and limits of 1CPU, 2Gb MEM.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// +optional
	Service ServiceSpec `json:"service,omitempty"`
}

// ServiceSpec defines the configuration for Knative Service.
type ServiceSpec struct {
	//+optional
	Autoscaler Autoscaler `json:"autoscaler,omitempty"`
	//+optional
	Metric AutoscalerMetric `json:"metric,omitempty"`
	//+optional
	Window string `json:"window,omitempty"`
	//+optional
	PanicWindow string `json:"panicWindow,omitempty"`
	//+optional
	PanicThreshold string `json:"panicThreshold,omitempty"`
	//+optional
	MinScale *int `json:"minScale,omitempty"`
	//+optional
	MaxScale int `json:"maxScale,omitempty"`
	//+optional
	Target int `json:"target,omitempty"`
	//+optional
	TargetUtilization string `json:"targetUtilization,omitempty"`
}

// Autoscaler defines the autoscaler class
//+kubebuilder:validation:Enum=kpa.autoscaling.knative.dev;hpa.autoscaling.knative.dev
type Autoscaler string

// AutoscalerMetric defines the metric for the autoscaler
//+kubebuilder:validation:Enum=concurrency;rps;cpu
type AutoscalerMetric string

// PaddleServiceStatus defines the observed state of PaddleService
type PaddleServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	duckv1beta1.Status `json:",inline"`
	// URL of the PaddleService
	URL string `json:"url,omitempty"`
	// Statuses for the default endpoints of the PaddleService
	Default *StatusConfigurationSpec `json:"default,omitempty"`
	// Addressable URL for eventing
	Address *duckv1beta1.Addressable `json:"address,omitempty"`

	// +optional
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas,omitempty"`
}

// +kubebuilder:object:root=true

// PaddleService is the Schema for the paddles API
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
type PaddleService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PaddleServiceSpec   `json:"spec,omitempty"`
	Status PaddleServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PaddleServiceList contains a list of PaddleService
type PaddleServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PaddleService `json:"items"`
}

// StatusConfigurationSpec describes the state of the configuration receiving traffic.
type StatusConfigurationSpec struct {
	// Latest revision name that is in ready state
	Name string `json:"name,omitempty"`
	// Host name of the service
	Hostname string `json:"host,omitempty"`
}

func init() {
	SchemeBuilder.Register(&PaddleService{}, &PaddleServiceList{})
}
