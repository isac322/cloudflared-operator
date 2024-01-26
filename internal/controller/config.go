package controller

import (
	"reflect"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

type TunnelConfig struct {
	*v1.TunnelRunParameters `json:",inline"`
	v1.OriginRequestConfig  `json:",inline"`
	Ingress                 []v1.TunnelConfigIngress `json:"ingress"`
}

func (c TunnelConfig) Equals(o TunnelConfig) bool {
	return reflect.DeepEqual(c, o)
}
