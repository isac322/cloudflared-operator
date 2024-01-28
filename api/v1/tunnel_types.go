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

type SecretKeyRef struct {
	// Name is Kubernetes's secret name that contains API token.
	//
	//+kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Key is Secret key. Defaults to token.
	//
	// +optional
	//+kubebuilder:default:=token
	Key *string `json:"key,omitempty"`

	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

// TunnelRunParameters represents the configurable options for Cloudflare Tunnel.
type TunnelRunParameters struct {
	// EdgeBindAddress defines the address that `cloudflared` will bind to on the Cloudflare edge.
	//
	// +optional
	//EdgeBindAddress *string `json:"edgeBindAddress,omitempty"`

	// EdgeIPVersion sets the IP version for edge connections.
	//
	// +optional
	//EdgeIPVersion *string `json:"edgeIPVersion,omitempty"`

	// GracePeriod specifies the time to wait for connections to close gracefully before exiting.
	//
	// +optional
	GracePeriod *metav1.Duration `json:"grace-period,omitempty"`

	// Logfile sets the path to the log file.
	//
	// +optional
	Logfile *string `json:"logfile,omitempty"`

	// Loglevel defines the level of logging (e.g., debug, info).
	//
	// +optional
	//+kubebuilder:validation:Enum:=debug;info;warn;error;fatal
	Loglevel *string `json:"loglevel,omitempty"`

	// Metrics sets the address for exposing the metrics reporting endpoint.
	// +optional
	//Metrics *string `json:"metrics,omitempty"`

	// NoAutoupdate indicates whether autoupdates are disabled.
	//
	// +optional
	//NoAutoupdate *bool `json:"noAutoupdate,omitempty"`

	// Origincert specifies the path to the certificate file for authentication with Cloudflare.
	//
	// +optional
	//Origincert *string `json:"origincert,omitempty"`

	// Pidfile sets the path to the PID file.
	// +optional
	Pidfile *string `json:"pidfile,omitempty"`

	// Protocol specifies the protocol for the tunnel.
	//
	// +optional
	Protocol *string `json:"protocol,omitempty"`

	// Region sets the preferred region for the edge connection.
	//
	// +optional
	Region *string `json:"region,omitempty"`

	// Retries specifies the number of retries for the tunnel connection.
	//
	// +optional
	Retries *int `json:"retries,omitempty"`

	// Tag contains tags for the tunnel in the format of key-value pairs.
	//
	// +optional
	Tag map[string]string `json:"tag,omitempty"`

	// Token specifies the token for authentication with Cloudflare.
	//
	// +optional
	//Token *string `json:"token,omitempty"`
}

// TunnelSpec defines the desired state of Tunnel
type TunnelSpec struct {
	// The tunnel name. It wil show up in Cloudflared's dashboard.
	//
	// +kubebuilder:validation:Pattern:="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	Name string `json:"name"`

	// Cloudflared's account id to create tunnel.
	// Refer https://developers.cloudflare.com/fundamentals/setup/find-account-and-zone-ids/ to find the value.
	//
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:MinLength=1
	AccountID string `json:"accountID"`

	// Reference to secret resource that contains Cloudflare API token.
	APITokenSecretRef SecretKeyRef `json:"apiTokenSecretRef"`

	// SecretName is for generated credential file. Defaults to cloudflare-tunnel-credential-<TUNNEL_NAME>
	//
	// +kubebuilder:validation:Pattern:="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// +optional
	SecretName *string `json:"secretName,omitempty"`

	// ConfigMapName is for generated config file Defaults to cloudflare-tunnel-<TUNNEL_NAME>
	//
	// +kubebuilder:validation:Pattern:="[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*"
	// +optional
	ConfigMapName *string `json:"configMapName,omitempty"`

	DaemonDeployment Deployment `json:"daemonDeployment"`

	// OriginConfiguration represents the configuration settings for cloudflared proxy to an origin server.
	// Refer https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/configure-tunnels/origin-configuration for details.
	//
	// +optional
	OriginConfiguration *OriginConfiguration `json:"originConfiguration,omitempty"`

	// +optional
	TunnelRunParameters *TunnelRunParameters `json:"tunnelRunParameters,omitempty"`
}

func (s *TunnelSpec) CredentialSecretName() string {
	if s.SecretName != nil {
		return *s.SecretName
	}
	return "cloudflare-tunnel-credential-" + s.Name
}

func (s *TunnelSpec) ConfigName() string {
	if s.ConfigMapName != nil {
		return *s.ConfigMapName
	}
	return "cloudflare-tunnel-" + s.Name
}

// TunnelConditionType ...
// +kubebuilder:validation:Enum=Daemon;Credential;Config
type TunnelConditionType string

const (
	TunnelConditionTypeDaemon     TunnelConditionType = "Daemon"
	TunnelConditionTypeCredential TunnelConditionType = "Credential"
	TunnelConditionTypeConfig     TunnelConditionType = "Config"
)

// TunnelConditionReason ...
// +kubebuilder:validation:Enum=CredentialRequired;ConfigRequired;FailedToDeleteOrphans;FailedToDeploy;DeletingOrphans;Creating;NoToken;FailedToConnectCloudflare;FailedToCreateTunnelOnCloudflare;FailedToCreateSecret;InvalidCredential;FailedToValidate;FailedToGetExistingCredential;FailedToBuildConfigFromSpec;FailedToGetExistingConfig;FailedToCreateConfigMap;FailedToUpdateConfigMap;InvalidConfig
type TunnelConditionReason string

const (
	DaemonReasonCredentialRequired    TunnelConditionReason = "CredentialRequired"
	DaemonReasonConfigRequired        TunnelConditionReason = "ConfigRequired"
	DaemonReasonFailedToDeleteOrphans TunnelConditionReason = "FailedToDeleteOrphans"
	DaemonReasonFailedToDeploy        TunnelConditionReason = "FailedToDeploy"
	DaemonReasonDeletingOrphans       TunnelConditionReason = "DeletingOrphans"

	CredentialReasonCreating                      TunnelConditionReason = "Creating"
	CredentialReasonNoToken                       TunnelConditionReason = "NoToken"
	CredentialReasonFailedToConnectCF             TunnelConditionReason = "FailedToConnectCloudflare"
	CredentialReasonFailedToCreateTunnelOnCF      TunnelConditionReason = "FailedToCreateTunnelOnCloudflare"
	CredentialReasonFailedToCreateSecret          TunnelConditionReason = "FailedToCreateSecret"
	CredentialReasonInvalidCredential             TunnelConditionReason = "InvalidCredential"
	CredentialReasonFailedToValidate              TunnelConditionReason = "FailedToValidate"
	CredentialReasonFailedToGetExistingCredential TunnelConditionReason = "FailedToGetExistingCredential"

	ConfigReasonFailedToBuildConfigFromSpec TunnelConditionReason = "FailedToBuildConfigFromSpec"
	ConfigReasonFailedToGetExistingConfig   TunnelConditionReason = "FailedToGetExistingConfig"
	ConfigReasonFailedToCreateConfigMap     TunnelConditionReason = "FailedToCreateConfigMap"
	ConfigReasonFailedToUpdateConfigMap     TunnelConditionReason = "FailedToUpdateConfigMap"
	ConfigReasonInvalidConfig               TunnelConditionReason = "InvalidConfig"
)

type TunnelStatusCondition struct {
	// Type of condition for a component.
	// Valid value: "Daemon", "Credential", "Config"
	Type TunnelConditionType `json:"type"`

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
	Reason TunnelConditionReason `json:"reason,omitempty"`
}

func (c TunnelStatusCondition) Equals(o TunnelStatusCondition) bool {
	return c.Type == o.Type && c.Status == o.Status && c.Message == o.Message &&
		c.Error == o.Error && c.Reason == o.Reason
}

// TunnelStatus defines the observed state of Tunnel
type TunnelStatus struct {
	Conditions []TunnelStatusCondition `json:"conditions,omitempty"`

	// +optional
	TunnelID string `json:"tunnelID,omitempty"`

	// +optional
	DaemonVersion string `json:"daemonVersion,omitempty"`
}

// Tunnel is the Schema for the tunnels API
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Account ID",type=string,JSONPath=`.spec.accountID`
// +kubebuilder:printcolumn:name="Tunnel Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Tunnel ID",type=string,JSONPath=`.status.tunnelID`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.daemonVersion`
type Tunnel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TunnelSpec   `json:"spec,omitempty"`
	Status TunnelStatus `json:"status,omitempty"`
}

// TunnelList contains a list of Tunnel
//
// +kubebuilder:object:root=true
type TunnelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tunnel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tunnel{}, &TunnelList{})
}
