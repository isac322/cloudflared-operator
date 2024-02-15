package controller

import (
	corev1 "k8s.io/api/core/v1"
)

type Condition[T ~string, C any] interface {
	Equals(C) bool
	GetConditionType() T
}

type Status[T ~string, C Condition[T, C]] interface {
	GetCondition(condType T) C
	SetCondition(condition C)
}

func UpdateConditionIfChanged[T ~string, C Condition[T, C]](status Status[T, C], cond C) bool {
	conditionType := cond.GetConditionType()
	prevCond := status.GetCondition(conditionType)
	if prevCond.GetConditionType() != conditionType {
		status.SetCondition(cond)
		return true
	}

	if prevCond.Equals(cond) {
		return false
	}
	status.SetCondition(cond)
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
