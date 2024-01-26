package controller

import v1 "github.com/isac322/cloudflared-operator/api/v1"

func GetTunnelCondition(status v1.TunnelStatus, condType v1.TunnelConditionType) *v1.TunnelStatusCondition {
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
