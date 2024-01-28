package controller

import (
	corev1 "k8s.io/api/core/v1"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

func GetTunnelCondition(status v1.TunnelStatus, condType v1.TunnelConditionType) *v1.TunnelStatusCondition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condType {
			return &status.Conditions[i]
		}
	}
	return nil
}

func GetTunnelIngressCondition(
	status v1.TunnelIngressStatus,
	condType v1.TunnelIngressConditionType,
) *v1.TunnelIngressStatusCondition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condType {
			return &status.Conditions[i]
		}
	}
	return nil
}

func SetTunnelConditionIfDiff(tunnel *v1.Tunnel, cond v1.TunnelStatusCondition) bool {
	prevCond := GetTunnelCondition(tunnel.Status, cond.Type)
	if prevCond == nil {
		tunnel.Status.Conditions = append(tunnel.Status.Conditions, cond)
		return true
	}

	if prevCond.Equals(cond) {
		return false
	}
	*prevCond = cond
	return true
}

func SetTunnelIngressConditionIfDiff(ingress *v1.TunnelIngress, cond v1.TunnelIngressStatusCondition) bool {
	prevCond := GetTunnelIngressCondition(ingress.Status, cond.Type)
	if prevCond == nil {
		ingress.Status.Conditions = append(ingress.Status.Conditions, cond)
		return true
	}

	if prevCond.Equals(cond) {
		return false
	}
	*prevCond = cond
	return true
}

func GetDataFromSecret(secret *corev1.Secret, key string) ([]byte, bool) {
	if secret.Data == nil && secret.StringData == nil {
		return nil, false
	}

	if data, ok := secret.Data[key]; ok {
		return data, true
	}

	if data, ok := secret.StringData[key]; ok {
		return []byte(data), true
	}

	return nil, false
}

func GetDataFromConfigMap(configMap *corev1.ConfigMap, key string) ([]byte, bool) {
	if configMap.Data == nil && configMap.BinaryData == nil {
		return nil, false
	}

	if data, ok := configMap.BinaryData[key]; ok {
		return data, true
	}

	if data, ok := configMap.Data[key]; ok {
		return []byte(data), true
	}

	return nil, false
}
