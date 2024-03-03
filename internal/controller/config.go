package controller

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
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

func (c TunnelConfig) Hash() (string, error) {
	marshal, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	md5Sum := md5.Sum(marshal)
	return hex.EncodeToString(md5Sum[:]), nil
}
