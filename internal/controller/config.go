package controller

import (
	"reflect"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

type TunnelConfig struct {
	*v1.TunnelRunParameters `json:",inline"`
	OriginRequestConfig     `json:",inline"`
	Ingress                 []TunnelConfigIngress `json:"ingress"`
}

func (c TunnelConfig) Equals(o TunnelConfig) bool {
	return reflect.DeepEqual(c, o)
}

type TunnelConfigIngress struct {
	Hostname      string              `json:"hostname,omitempty"`
	Path          string              `json:"path,omitempty"`
	Service       string              `json:"service,omitempty"`
	OriginRequest OriginRequestConfig `json:"originRequest"`
}

type OriginRequestConfig struct {
	*v1.OriginTLSSettings        `json:",inline"`
	*v1.OriginHTTPSettings       `json:",inline"`
	*v1.OriginConnectionSettings `json:",inline"`
	*v1.OriginAccessSettings     `json:",inline"`
}
