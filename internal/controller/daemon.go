package controller

import (
	"context"
	"fmt"
	"maps"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
	"github.com/isac322/cloudflared-operator/internal/cloudflare"
)

func getDaemonVersion(ctx context.Context, tunnel *v1.Tunnel) (string, error) {
	version := tunnel.Spec.DaemonDeployment.DaemonVersion
	if version == "latest" {
		return cloudflare.GetLatestDaemonVersion(ctx)
	}

	isValid, err := cloudflare.VerifyDaemonVersion(ctx, version)
	if err != nil {
		return "", err
	}
	if !isValid {
		return "", fmt.Errorf("invalid daemon version: %s", version)
	}
	return version, nil
}

func buildDaemon(daemonVersion string, tunnel *v1.Tunnel, tunnelConfig TunnelConfig) (client.Object, error) {
	image := "cloudflare/cloudflared:" + daemonVersion

	configHash, err := tunnelConfig.Hash()
	if err != nil {
		return nil, err
	}

	var terminationGracePeriodSeconds *int64
	if tunnel.Spec.TunnelRunParameters != nil && tunnel.Spec.TunnelRunParameters.GracePeriod != nil {
		*terminationGracePeriodSeconds = int64(tunnel.Spec.TunnelRunParameters.GracePeriod.Seconds()) + 5
	}

	podAnnotations := maps.Clone(tunnel.Spec.DaemonDeployment.PodAnnotations)
	if podAnnotations == nil {
		podAnnotations = make(map[string]string, 1)
	}
	podAnnotations["cloudflared-operator.bhyoo.com/config-hash"] = configHash

	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      fillLabels(tunnel.Spec.DaemonDeployment.PodLabels, tunnel.Name, daemonVersion),
			Annotations: podAnnotations,
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "credential",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: tunnel.Spec.CredentialSecretName(),
							Items: []corev1.KeyToPath{{
								Key:  fileNameCredential,
								Path: fileNameCredential,
							}},
						},
					},
				},
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: tunnel.Spec.ConfigName()},
							Items: []corev1.KeyToPath{{
								Key:  fileNameConfig,
								Path: fileNameConfig,
							}},
						},
					},
				},
			},
			InitContainers: nil,
			Containers: []corev1.Container{{
				Name:    "cloudflared",
				Image:   image,
				Command: nil,
				Args: []string{
					"tunnel",
					"--no-autoupdate",
					"--metrics",
					"0.0.0.0:2000",
					"--config",
					"/etc/cloudflared/" + fileNameConfig,
					"--credentials-file",
					"/etc/cloudflared/creds/" + fileNameCredential,
					"run",
					tunnel.Spec.Name,
				},
				WorkingDir:    "",
				Ports:         nil,
				EnvFrom:       nil,
				Env:           nil,
				Resources:     tunnel.Spec.DaemonDeployment.Resources,
				ResizePolicy:  nil,
				RestartPolicy: nil,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "config",
						ReadOnly:  true,
						MountPath: "/etc/cloudflared",
					},
					{
						Name:      "credential",
						ReadOnly:  true,
						MountPath: "/etc/cloudflared/creds",
					},
				},
				VolumeDevices: nil,
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/ready",
							Port: intstr.FromInt32(2000),
						},
					},
					InitialDelaySeconds: 10,
					PeriodSeconds:       10,
					FailureThreshold:    1,
				},
				ReadinessProbe:           nil,
				StartupProbe:             nil,
				Lifecycle:                nil,
				TerminationMessagePath:   "",
				TerminationMessagePolicy: "",
				ImagePullPolicy:          "",
				SecurityContext: &corev1.SecurityContext{
					Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
					ReadOnlyRootFilesystem:   ptr.To(true),
					AllowPrivilegeEscalation: ptr.To(false),
				},
				Stdin:     false,
				StdinOnce: false,
				TTY:       false,
			}},
			EphemeralContainers:           nil,
			RestartPolicy:                 "",
			TerminationGracePeriodSeconds: terminationGracePeriodSeconds,
			ActiveDeadlineSeconds:         nil,
			DNSPolicy:                     tunnel.Spec.DaemonDeployment.DNSPolicy,
			NodeSelector:                  tunnel.Spec.DaemonDeployment.NodeSelector,
			ServiceAccountName:            "",
			AutomountServiceAccountToken:  nil,
			NodeName:                      "",
			ShareProcessNamespace:         nil,
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:    ptr.To(int64(65532)),
				RunAsNonRoot: ptr.To(true),
				//Sysctls: []corev1.Sysctl{
				//	{
				//		Name:  "net.core.rmem_max",
				//		Value: "2500000",
				//	},
				//	{
				//		Name:  "net.core.wmem_max",
				//		Value: "2500000",
				//	},
				//},
			},
			ImagePullSecrets:          nil,
			Hostname:                  "",
			Subdomain:                 "",
			Affinity:                  tunnel.Spec.DaemonDeployment.Affinity,
			SchedulerName:             "",
			Tolerations:               tunnel.Spec.DaemonDeployment.Tolerations,
			HostAliases:               nil,
			PriorityClassName:         "",
			Priority:                  nil,
			DNSConfig:                 nil,
			ReadinessGates:            nil,
			RuntimeClassName:          nil,
			EnableServiceLinks:        nil,
			PreemptionPolicy:          nil,
			Overhead:                  nil,
			TopologySpreadConstraints: nil,
			SetHostnameAsFQDN:         nil,
			OS:                        nil,
			HostUsers:                 nil,
			SchedulingGates:           nil,
			ResourceClaims:            nil,
		},
	}

	labelSelector := map[string]string{
		"app.kubernetes.io/name":     "cloudflared",
		"app.kubernetes.io/instance": tunnel.Name,
		"app.kubernetes.io/part-of":  "cloudflared",
	}

	daemonAnnotations := maps.Clone(tunnel.Spec.DaemonDeployment.Annotations)
	if daemonAnnotations == nil {
		daemonAnnotations = make(map[string]string, 1)
	}
	daemonAnnotations["cloudflared-operator.bhyoo.com/config-hash"] = configHash

	if tunnel.Spec.DaemonDeployment.Kind == v1.DeploymentKindDeployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        buildDaemonName(tunnel),
				Namespace:   tunnel.Namespace,
				Labels:      fillLabels(tunnel.Spec.DaemonDeployment.Labels, tunnel.Name, daemonVersion),
				Annotations: daemonAnnotations,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: tunnel.Spec.DaemonDeployment.Replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: labelSelector,
				},
				Template:                podTemplateSpec,
				Strategy:                tunnel.Spec.DaemonDeployment.DeploymentStrategy,
				MinReadySeconds:         tunnel.Spec.DaemonDeployment.MinReadySeconds,
				RevisionHistoryLimit:    tunnel.Spec.DaemonDeployment.RevisionHistoryLimit,
				Paused:                  false,
				ProgressDeadlineSeconds: nil,
			},
		}, nil
	}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "cloudflared-" + tunnel.Name + "-" + tunnel.Spec.Name,
			Namespace:   tunnel.Namespace,
			Labels:      tunnel.Spec.DaemonDeployment.Labels,
			Annotations: daemonAnnotations,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
			Template:             podTemplateSpec,
			UpdateStrategy:       tunnel.Spec.DaemonDeployment.DaemonSetUpdateStrategy,
			MinReadySeconds:      tunnel.Spec.DaemonDeployment.MinReadySeconds,
			RevisionHistoryLimit: tunnel.Spec.DaemonDeployment.RevisionHistoryLimit,
		},
	}, nil
}

func isDaemonEqualTo(a, b client.Object, kind v1.DeploymentKind) bool {
	switch kind {
	case v1.DeploymentKindDaemonSet:
		aa, ok := a.(*appsv1.DaemonSet)
		if !ok {
			return false
		}
		bb, ok := b.(*appsv1.DaemonSet)
		if !ok {
			return false
		}
		return reflect.DeepEqual(aa.Spec, bb.Spec)
	case v1.DeploymentKindDeployment:
		aa, ok := a.(*appsv1.Deployment)
		if !ok {
			return false
		}
		bb, ok := b.(*appsv1.Deployment)
		if !ok {
			return false
		}
		return reflect.DeepEqual(aa.Spec, bb.Spec)
	default:
		return false
	}
}

func buildDaemonName(tunnel *v1.Tunnel) string {
	return "cloudflared-" + tunnel.Name + "-" + tunnel.Spec.Name
}

func fillLabels(labels map[string]string, tunnelName, version string) map[string]string {
	dest := maps.Clone(labels)
	if dest == nil {
		dest = make(map[string]string, 5)
	}
	dest["app.kubernetes.io/name"] = "cloudflared"
	dest["app.kubernetes.io/instance"] = tunnelName
	dest["app.kubernetes.io/component"] = "daemon"
	dest["app.kubernetes.io/part-of"] = "cloudflared"
	dest["app.kubernetes.io/version"] = version
	return dest
}
