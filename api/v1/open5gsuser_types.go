/*
Copyright 2024.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Open5GSReference defines the reference to an Open5GS instance
type Open5GSReference struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// Open5GSUserSpec defines the desired state of Open5GSUser
type Open5GSUserSpec struct {
	IMSI    string           `json:"imsi,omitempty"`
	Key     string           `json:"key,omitempty"`
	OPC     string           `json:"opc,omitempty"`
	SD      string           `json:"sd,omitempty"`
	SST     string           `json:"sst,omitempty"`
	APN     string           `json:"apn,omitempty"`
	Open5GS Open5GSReference `json:"open5gs,omitempty"`
}

// Open5GSUserStatus defines the observed state of Open5GSUser
type Open5GSUserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Open5GSUser is the Schema for the open5gsusers API
type Open5GSUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Open5GSUserSpec   `json:"spec,omitempty"`
	Status Open5GSUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Open5GSUserList contains a list of Open5GSUser
type Open5GSUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Open5GSUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Open5GSUser{}, &Open5GSUserList{})
}
