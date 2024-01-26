/*
Copyright 2024 Byeonghoon Yoo.

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

type TunnelRef struct {
	// Name is Tunnel name that bind to the TunnelIngress.
	Name string `json:"name"`

	// Kind is the type of the resource. Defaults to `Tunnel`.
	// +optional
	// +kubebuilder:validation:Enum=Tunnel
	// +kubebuilder:default:=Tunnel
	Kind string `json:"kind,omitempty"`
}

// TunnelIngressSpec defines the desired state of TunnelIngress
type TunnelIngressSpec struct {
	TunnelConfigIngress `json:",inline"`

	TunnelRef TunnelRef `json:"tunnelRef"`

	// +optional
	DNSRecordTTL *metav1.Duration `json:"dnsRecordTTL,omitempty"`
}

// TunnelIngressStatus defines the observed state of TunnelIngress
type TunnelIngressStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TunnelIngress is the Schema for the tunnelingresses API
type TunnelIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TunnelIngressSpec   `json:"spec,omitempty"`
	Status TunnelIngressStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TunnelIngressList contains a list of TunnelIngress
type TunnelIngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TunnelIngress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TunnelIngress{}, &TunnelIngressList{})
}
