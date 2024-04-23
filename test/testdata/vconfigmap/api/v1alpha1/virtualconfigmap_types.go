/*
Copyright 2024 The VCache Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VirtualConfigMapSpec defines the desired state of VirtualConfigMap
type VirtualConfigMapSpec struct {
	Data map[string]string `json:"data,omitempty"`
}

// VirtualConfigMapStatus defines the observed state of VirtualConfigMap
type VirtualConfigMapStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VirtualConfigMap is the Schema for the virtualconfigmaps API
type VirtualConfigMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualConfigMapSpec   `json:"spec,omitempty"`
	Status VirtualConfigMapStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VirtualConfigMapList contains a list of VirtualConfigMap
type VirtualConfigMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualConfigMap `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualConfigMap{}, &VirtualConfigMapList{})
}
