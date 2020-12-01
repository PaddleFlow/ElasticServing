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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PaddleSpec defines the desired state of Paddle
type PaddleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Paddle. Edit Paddle_types.go to remove/update
	Foo string `json:"foo,omitempty"`
	// The URI for the saved model
	StorageURI string `json:"storageUri,omitempty"`
	// Docker image version
	RuntimeVersion string `json:"runtimeVersion,omitempty"`
	// Defaults to requests and limits of 1CPU, 2Gb MEM.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	Replicas *int32 `json:"replicas,omitempty"`
}

// PaddleStatus defines the observed state of Paddle
type PaddleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// URL of the InferenceService
	URL string `json:"url,omitempty"`
	// Traffic percentage that goes to default services
	Traffic int `json:"traffic,omitempty"`
	// Traffic percentage that goes to canary services
	CanaryTraffic int `json:"canaryTraffic,omitempty"`

	Replicas int32 `json:"replicas,omitempty"`
}

// +kubebuilder:object:root=true

// Paddle is the Schema for the paddles API
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
type Paddle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PaddleSpec   `json:"spec,omitempty"`
	Status PaddleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PaddleList contains a list of Paddle
type PaddleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Paddle `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Paddle{}, &PaddleList{})
}
