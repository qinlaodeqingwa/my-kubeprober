/*
Copyright 2026.

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

// ProbeSpec defines the desired state of Probe
type ProbeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of Probe. Edit probe_types.go to remove/update
	// +optional
	Policy   Policy                 `json:"policy,omitempty"`
	Template corev1.PodTemplateSpec `json:"template,omitempty"`
	configs  []Config               `json:"configs,omitempty"`
}
type Policy struct {
	RunInterval int `json:"runInterval,omitempty"`
}
type Config struct {
	Name string          `json:"name,omitempty"`
	Env  []corev1.EnvVar `json:"env,omitempty"`
}

// ProbeStatus defines the observed state of Probe.
type ProbeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the Probe resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Phase       string       `json:"phase,omitempty"`
	LastRunTime *metav1.Time `json:"lastRunTime,omitempty"`
	Message     string       `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Probe is the Schema for the probes API
type Probe struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of Probe
	// +required
	Spec ProbeSpec `json:"spec,omitempty"`

	// status defines the observed state of Probe
	// +optional
	Status ProbeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProbeList contains a list of Probe
type ProbeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Probe `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Probe{}, &ProbeList{})
}
