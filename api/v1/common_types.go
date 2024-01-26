package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OriginTLSSettings holds the TLS specific settings.
type OriginTLSSettings struct {
	// OriginServerName is the server name used in the origin server certificate.
	// +optional
	OriginServerName *string `json:"originServerName,omitempty"`

	// CAPool is the path to the certificate authority pool.
	// +optional
	CAPool *string `json:"caPool,omitempty"`

	// NoTLSVerify controls whether TLS verification is bypassed.
	// +optional
	NoTLSVerify *bool `json:"noTLSVerify,omitempty"`

	// TLSTimeout is the timeout for TLS connections.
	// +optional
	//+kubebuilder:default:="10s"
	TLSTimeout *metav1.Duration `json:"tlsTimeout,omitempty"`

	// HTTP2Origin enables HTTP/2 support to the origin.
	// +optional
	HTTP2Origin *bool `json:"http2Origin,omitempty"`
}

// OriginHTTPSettings holds settings specific to HTTP protocol.
type OriginHTTPSettings struct {
	// HTTPHostHeader is the HTTP Host header to use in requests to the origin.
	// +optional
	HTTPHostHeader *string `json:"httpHostHeader,omitempty"`

	// DisableChunkedEncoding determines whether chunked encoding is disabled.
	// +optional
	DisableChunkedEncoding *bool `json:"disableChunkedEncoding,omitempty"`
}

// OriginConnectionSettings contains settings related to network connections.
type OriginConnectionSettings struct {
	// ConnectTimeout is the timeout for establishing new connections.
	// +optional
	//+kubebuilder:default:="30s"
	ConnectTimeout *string `json:"connectTimeout,omitempty"`

	// NoHappyEyeballs disables "Happy Eyeballs" for IPv4/IPv6 fallback.
	// +optional
	NoHappyEyeballs *bool `json:"noHappyEyeballs,omitempty"`

	// ProxyType is the type of proxy to use (e.g., HTTP, SOCKS).
	// +optional
	ProxyType *string `json:"proxyType,omitempty"`

	// ProxyAddress is the address of the proxy server.
	// +optional
	//+kubebuilder:default:="127.0.0.1"
	ProxyAddress *string `json:"proxyAddress,omitempty"`

	// ProxyPort is the port of the proxy server.
	// +optional
	ProxyPort *int `json:"proxyPort,omitempty"`

	// KeepAliveTimeout is the timeout for keeping connections alive.
	// +optional
	//+kubebuilder:default:="1m30s"
	KeepAliveTimeout *metav1.Duration `json:"keepAliveTimeout,omitempty"`

	// KeepAliveConnections is the maximum number of keep-alive connections.
	// +optional
	//+kubebuilder:default:=100
	KeepAliveConnections *int `json:"keepAliveConnections,omitempty"`

	// TCPKeepAlive is the keep-alive time for TCP connections.
	// +optional
	//+kubebuilder:default:="30s"
	TCPKeepAlive *metav1.Duration `json:"tcpKeepAlive,omitempty"`
}

type OriginAccessSettingsAccess struct {
	// Required indicates if access control is required.
	// +optional
	Required *bool `json:"required,omitempty"`

	// TeamName specifies the team name for access control.
	// +optional
	TeamName *string `json:"teamName,omitempty"`

	// AudTag is a list of audit tags for access control.
	// +optional
	AudTag []string `json:"audTag,omitempty"`
}

// OriginAccessSettings contains settings related to access control.
type OriginAccessSettings struct {
	// Access struct holds the access control settings.
	// +optional
	Access *OriginAccessSettingsAccess `json:"access,omitempty"`
}

// OriginConfiguration represents the configuration settings for cloudflared proxy to an origin server.
// Refer https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/configure-tunnels/origin-configuration for details.
type OriginConfiguration struct {
	// TLSSettings holds the TLS specific settings.
	// +optional
	TLSSettings *OriginTLSSettings `json:"tlsSettings,omitempty"`

	// HTTPSettings holds settings specific to HTTP protocol.
	// +optional
	HTTPSettings *OriginHTTPSettings `json:"httpSettings,omitempty"`

	// ConnectionSettings contains settings related to network connections.
	// +optional
	ConnectionSettings *OriginConnectionSettings `json:"connectionSettings,omitempty"`

	// AccessSettings contains settings related to access control.
	// +optional
	AccessSettings *OriginAccessSettings `json:"accessSettings,omitempty"`
}

// DeploymentKind ...
// +kubebuilder:validation:Enum=DaemonSet;Deployment
type DeploymentKind string

const (
	DeploymentKindDaemonSet  DeploymentKind = "DaemonSet"
	DeploymentKindDeployment DeploymentKind = "Deployment"
)

type Deployment struct {
	// DaemonVersion specify Cloudfalred version to deploy. Defaults to latest.
	// Refer https://github.com/cloudflare/cloudflared/releases to available versions.
	//
	// +optional
	//+kubebuilder:default:=latest
	DaemonVersion string `json:"daemonVersion,omitempty"`

	//+kubebuilder:default:=Deployment
	Kind DeploymentKind `json:"kind"`

	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1. (only applies when kind == Deployment)
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// The deployment strategy to use to replace existing pods with new ones. (only applies when kind == Deployment)
	// +optional
	// +patchStrategy=retainKeys
	DeploymentStrategy appsv1.DeploymentStrategy `json:"DeploymentStrategy,omitempty" patchStrategy:"retainKeys"`

	// An update strategy to replace existing DaemonSet pods with new pods. (only applies when kind == DaemonSet)
	// +optional
	DaemonSetUpdateStrategy appsv1.DaemonSetUpdateStrategy `json:"updateStrategy,omitempty"`

	// The minimum number of seconds for which a newly created DaemonSet pod should
	// be ready without any of its container crashing, for it to be considered
	// available. Defaults to 0 (pod will be considered available as soon as it
	// is ready).
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`

	// The number of old history to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels
	// +optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// PodAnnotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations
	// +optional
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// Set DNS policy for the pod.
	// Defaults to "ClusterFirst".
	// Valid values are 'ClusterFirstWithHostNet', 'ClusterFirst', 'Default' or 'None'.
	// DNS parameters given in DNSConfig will be merged with the policy selected with DNSPolicy.
	// To have DNS options set along with hostNetwork, you have to specify DNS policy
	// explicitly to 'ClusterFirstWithHostNet'.
	// +optional
	DNSPolicy corev1.DNSPolicy `json:"dnsPolicy,omitempty"`

	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type TunnelConfigIngress struct {
	Hostname      *string              `json:"hostname,omitempty"`
	Path          *string              `json:"path,omitempty"`
	Service       string               `json:"service,omitempty"`
	OriginRequest *OriginRequestConfig `json:"originRequest,omitempty"`
}

type OriginRequestConfig struct {
	*OriginTLSSettings        `json:",inline"`
	*OriginHTTPSettings       `json:",inline"`
	*OriginConnectionSettings `json:",inline"`
	*OriginAccessSettings     `json:",inline"`
}
