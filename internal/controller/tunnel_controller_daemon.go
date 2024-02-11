package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
	"github.com/isac322/cloudflared-operator/internal/cloudflare"
)

func (r *TunnelReconciler) reconcileDaemon(ctx context.Context, tunnel *v1.Tunnel) error {
	recordConditionFrom := r.buildConditionRecorder(ctx, tunnel, v1.TunnelConditionTypeDaemon)

	objectKey := client.ObjectKey{Namespace: tunnel.Namespace, Name: buildDaemonName(tunnel)}
	var deployNotExists, daemonSetNotExists bool
	var deployment appsv1.Deployment
	if err := r.Get(ctx, objectKey, &deployment); err != nil {
		if !apierrors.IsNotFound(err) {
			return recordConditionFrom(err)
		}
		deployNotExists = true
	}
	var daemonSet appsv1.DaemonSet
	if err := r.Get(ctx, objectKey, &daemonSet); err != nil {
		if !apierrors.IsNotFound(err) {
			return recordConditionFrom(err)
		}
		daemonSetNotExists = true
	}

	var target, orphan client.Object
	switch k := tunnel.Spec.DaemonDeployment.Kind; k {
	case v1.DeploymentKindDaemonSet:
		if !daemonSetNotExists {
			target = &daemonSet
		}
		if !deployNotExists {
			orphan = &deployment
		}

	case v1.DeploymentKindDeployment:
		if !deployNotExists {
			target = &deployment
		}
		if !daemonSetNotExists {
			orphan = &daemonSet
		}

	default:
		return recordConditionFrom(WrapError(
			fmt.Errorf("unknown kind: %s", k),
			v1.ConfigReasonFailedToGetExistingConfig,
		))
	}

	var dirtyStatus bool
	// delete orphan
	if orphan != nil {
		dirtyStatus = true
		if err := r.updateConditionIfDiff(ctx, tunnel, v1.TunnelStatusCondition{
			Type:               v1.TunnelConditionTypeDaemon,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Time{Time: r.Clock.Now()},
			Reason:             v1.DaemonReasonDeletingOrphans,
		}); err != nil {
			return err
		}
		if err := r.Delete(ctx, orphan); err != nil {
			return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeleteOrphans))
		}
	}

	daemon, daemonVersion, err := buildDaemon(ctx, tunnel)
	if err != nil {
		return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeploy))
	}
	if target != nil {
		dirtyStatus = true
		if err = ctrl.SetControllerReference(tunnel, daemon, r.Scheme); err != nil {
			return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeploy))
		}
		if err = r.Update(ctx, daemon); err != nil {
			return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeploy))
		}
	} else {
		if err = r.Create(ctx, daemon); err != nil {
			return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeploy))
		}
	}

	if SetTunnelConditionIfDiff(tunnel, v1.TunnelStatusCondition{
		Type:               v1.TunnelConditionTypeDaemon,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: r.Clock.Now()},
	}) {
		dirtyStatus = true
	}

	if tunnel.Status.DaemonVersion != daemonVersion {
		tunnel.Status.DaemonVersion = daemonVersion
		dirtyStatus = true
	}

	if dirtyStatus {
		return r.Status().Update(ctx, tunnel)
	}
	return nil
}

func buildDaemon(ctx context.Context, tunnel *v1.Tunnel) (client.Object, string, error) {
	version := tunnel.Spec.DaemonDeployment.DaemonVersion
	if version == "latest" {
		latestVersion, err := cloudflare.GetLatestDaemonVersion(ctx)
		if err != nil {
			return nil, "", err
		}
		version = latestVersion
	} else {
		isValid, err := cloudflare.VerifyDaemonVersion(ctx, version)
		if err != nil {
			return nil, "", err
		}
		if !isValid {
			return nil, "", fmt.Errorf("invalid daemon version: %s", version)
		}
	}
	image := "cloudflare/cloudflared:" + version

	var terminationGracePeriodSeconds *int64
	if tunnel.Spec.TunnelRunParameters != nil && tunnel.Spec.TunnelRunParameters.GracePeriod != nil {
		*terminationGracePeriodSeconds = int64(tunnel.Spec.TunnelRunParameters.GracePeriod.Seconds()) + 5
	}

	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      fillLabels(tunnel.Spec.DaemonDeployment.PodLabels, tunnel.Name, version),
			Annotations: tunnel.Spec.DaemonDeployment.PodAnnotations,
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

	if tunnel.Spec.DaemonDeployment.Kind == v1.DeploymentKindDeployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        buildDaemonName(tunnel),
				Namespace:   tunnel.Namespace,
				Labels:      fillLabels(tunnel.Spec.DaemonDeployment.Labels, tunnel.Name, version),
				Annotations: tunnel.Spec.DaemonDeployment.Annotations,
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
		}, version, nil
	}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "cloudflared-" + tunnel.Name + "-" + tunnel.Spec.Name,
			Namespace:   tunnel.Namespace,
			Labels:      tunnel.Spec.DaemonDeployment.Labels,
			Annotations: tunnel.Spec.DaemonDeployment.Annotations,
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
	}, version, nil
}

func buildDaemonName(tunnel *v1.Tunnel) string {
	return "cloudflared-" + tunnel.Name + "-" + tunnel.Spec.Name
}

func fillLabels(labels map[string]string, tunnelName, version string) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 5)
	}
	labels["app.kubernetes.io/name"] = "cloudflared"
	labels["app.kubernetes.io/instance"] = tunnelName
	labels["app.kubernetes.io/component"] = "daemon"
	labels["app.kubernetes.io/part-of"] = "cloudflared"
	labels["app.kubernetes.io/version"] = version
	return labels
}
