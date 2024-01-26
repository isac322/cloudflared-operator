package controller

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

var (
	errNotFoundConfigKey = errors.New("config key is not found")
	errUnReadableConfig  = errors.New("unreadable config")
)

func (r *TunnelReconciler) reconcileConfig(ctx context.Context, tunnel *v1.Tunnel) error {
	l := log.FromContext(ctx)

	recordConditionFrom := r.buildConditionRecorder(ctx, tunnel, v1.TunnelConditionTypeConfig)

	config, err := r.buildConfig(ctx, tunnel)
	if err != nil {
		return recordConditionFrom(err)
	}

	var dirtyStatus bool

	var configMap corev1.ConfigMap
	err = r.Get(ctx, client.ObjectKey{Namespace: tunnel.Namespace, Name: tunnel.Spec.ConfigName()}, &configMap)
	switch {
	case err == nil:
		prevConfig, err := readTunnelConfig(configMap)
		if err != nil || !prevConfig.Equals(config) {
			l.Info("outdated config. force update")
			marshaledConfig, err := yaml.Marshal(config)
			if err != nil {
				return recordConditionFrom(WrapError(err, v1.ConfigReasonInvalidConfig))
			}
			if configMap.Data == nil {
				configMap.Data = make(map[string]string, 1)
			}
			delete(configMap.BinaryData, fileNameConfig)
			configMap.Data[fileNameConfig] = string(marshaledConfig)
			if err := r.Update(ctx, &configMap); err != nil {
				return recordConditionFrom(WrapError(err, v1.ConfigReasonFailedToUpdateConfigMap))
			}
		}

	case apierrors.IsNotFound(err):
		dirtyStatus = true

		marshaledConfig, err := yaml.Marshal(config)
		if err != nil {
			return recordConditionFrom(WrapError(err, v1.ConfigReasonFailedToCreateConfigMap))
		}

		configMap = corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        tunnel.Spec.ConfigName(),
				Namespace:   tunnel.Namespace,
				Labels:      nil,
				Annotations: nil,
			},
			Data: map[string]string{fileNameConfig: string(marshaledConfig)},
		}
		if err := ctrl.SetControllerReference(tunnel, &configMap, r.Scheme); err != nil {
			return recordConditionFrom(WrapError(err, v1.ConfigReasonFailedToCreateConfigMap))
		}
		if err := r.Create(ctx, &configMap); err != nil {
			return recordConditionFrom(WrapError(err, v1.ConfigReasonFailedToCreateConfigMap))
		}

	// unknown errors
	default:
		return recordConditionFrom(WrapError(err, v1.ConfigReasonFailedToGetExistingConfig))
	}

	if dirtyStatus || SetTunnelConditionIfDiff(tunnel, v1.TunnelStatusCondition{
		Type:               v1.TunnelConditionTypeConfig,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: r.Clock.Now()},
	}) {
		return r.Status().Update(ctx, tunnel)
	}
	return nil
}

func (r *TunnelReconciler) buildConfig(ctx context.Context, tunnel *v1.Tunnel) (TunnelConfig, error) {
	var bindings v1.TunnelBindingList
	err := r.List(ctx, &bindings, client.MatchingLabels{"cloudflared-operator.bhyoo.com/tunnel-name": tunnel.Name})
	if err != nil {
		return TunnelConfig{}, err
	}

	config := TunnelConfig{
		TunnelRunParameters: tunnel.Spec.TunnelRunParameters,
		OriginRequestConfig: OriginRequestConfig{},
		Ingress:             make([]TunnelConfigIngress, 0, len(bindings.Items)+1),
	}
	for _, binding := range bindings.Items {
		config.Ingress = append(config.Ingress, TunnelConfigIngress{
			Hostname:      binding.Spec.Hostname,
			Path:          binding.Spec.Path,
			Service:       binding.Spec.Service,
			OriginRequest: OriginRequestConfig{},
		})
	}
	config.Ingress = append(config.Ingress, TunnelConfigIngress{Service: "http_status:404"})

	if tunnel.Spec.OriginConfiguration != nil {
		config.OriginRequestConfig.OriginTLSSettings = tunnel.Spec.OriginConfiguration.TLSSettings
		config.OriginRequestConfig.OriginHTTPSettings = tunnel.Spec.OriginConfiguration.HTTPSettings
		config.OriginRequestConfig.OriginConnectionSettings = tunnel.Spec.OriginConfiguration.ConnectionSettings
		config.OriginRequestConfig.OriginAccessSettings = tunnel.Spec.OriginConfiguration.AccessSettings
	}

	return config, nil
}

func readTunnelConfig(cm corev1.ConfigMap) (config TunnelConfig, err error) {
	v, ok := cm.Data[fileNameConfig]
	bytesV := []byte(v)
	if !ok {
		if bytesV, ok = cm.BinaryData[fileNameConfig]; !ok {
			return TunnelConfig{}, errNotFoundConfigKey
		}
	}

	if err := yaml.Unmarshal(bytesV, &config); err != nil {
		return TunnelConfig{}, errors.Join(errUnReadableConfig, err)
	}
	return config, nil
}
