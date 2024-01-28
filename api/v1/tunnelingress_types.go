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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TunnelKind ...
// +kubebuilder:validation:Enum=Tunnel
type TunnelKind string

const (
	TunnelKindTunnel TunnelKind = "Tunnel"
)

type TunnelRef struct {
	// Name is Tunnel name that bind to the TunnelIngress.
	Name string `json:"name"`

	// Kind is the type of the resource. Defaults to `Tunnel`.
	// +optional
	// +kubebuilder:default:=Tunnel
	Kind TunnelKind `json:"kind,omitempty"`
}

// TunnelIngressSpec defines the desired state of TunnelIngress
type TunnelIngressSpec struct {
	TunnelConfigIngress `json:",inline"`

	TunnelRef TunnelRef `json:"tunnelRef"`

	// +optional
	DNSRecordTTL *metav1.Duration `json:"dnsRecordTTL,omitempty"`
}

// TunnelIngressConditionType ...
// +kubebuilder:validation:Enum=DNSRecord
type TunnelIngressConditionType string

const (
	TunnelIngressConditionTypeDNSRecord TunnelIngressConditionType = "DNSRecord"
)

// TunnelIngressConditionReason ...
// +kubebuilder:validation:Enum=Creating;NoToken;FailedToConnectCloudflare
type TunnelIngressConditionReason string

const (
	DNSRecordReasonCreating             TunnelIngressConditionReason = "Creating"
	DNSRecordReasonNoToken              TunnelIngressConditionReason = "NoToken"
	DNSRecordReasonFailedToConnectCF    TunnelIngressConditionReason = "FailedToConnectCloudflare"
	DNSRecordReasonFailedToCreateRecord TunnelIngressConditionReason = "FailedToCreateRecord"
)

type TunnelIngressStatusCondition struct {
	// Type of condition for a component.
	// Valid value: "Daemon", "Credential", "Config"
	Type TunnelIngressConditionType `json:"type"`

	// Status of the condition for a component.
	// Valid values for "Daemon", "Credential", "Config": "True", "False", or "Unknown".
	Status corev1.ConditionStatus `json:"status"`

	// Message about the condition for a component.
	// For example, information about a health check.
	// +optional
	Message string `json:"message,omitempty"`

	// Error is Condition error code for a component.
	// For example, a health check error code.
	// +optional
	Error string `json:"error,omitempty"`

	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// +optional
	Reason TunnelIngressConditionReason `json:"reason,omitempty"`
}

func (c TunnelIngressStatusCondition) Equals(o TunnelIngressStatusCondition) bool {
	return c.Type == o.Type && c.Status == o.Status && c.Message == o.Message &&
		c.Error == o.Error && c.Reason == o.Reason
}

// TunnelIngressStatus defines the observed state of TunnelIngress
type TunnelIngressStatus struct {
	Conditions []TunnelIngressStatusCondition `json:"conditions,omitempty"`
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
