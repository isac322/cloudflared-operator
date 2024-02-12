package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

func (r *TunnelReconciler) reconcileDaemon(ctx context.Context, tunnel *v1.Tunnel, tunnelConfig TunnelConfig) error {
	recordConditionFrom := r.buildConditionRecorder(ctx, tunnel, v1.TunnelConditionTypeDaemon)

	target, orphan, err := r.getExistingDaemons(ctx, tunnel)
	if err != nil {
		return recordConditionFrom(err)
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

	daemonVersion, err := getDaemonVersion(ctx, tunnel)
	if err != nil {
		return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeploy))
	}
	newTarget, err := buildDaemon(daemonVersion, tunnel, tunnelConfig)
	if err != nil {
		return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeploy))
	}
	if err = ctrl.SetControllerReference(tunnel, newTarget, r.Scheme); err != nil {
		return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeploy))
	}

	if target != nil {
		if err = r.Update(ctx, newTarget, client.DryRunAll); err != nil {
			return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeploy))
		}
		if !isDaemonEqualTo(target, newTarget, tunnel.Spec.DaemonDeployment.Kind) {
			dirtyStatus = true
			if err = r.Update(ctx, newTarget); err != nil {
				return recordConditionFrom(WrapError(err, v1.DaemonReasonFailedToDeploy))
			}
		}
	} else {
		if err = r.Create(ctx, newTarget); err != nil {
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

func (r *TunnelReconciler) getExistingDaemons(
	ctx context.Context,
	tunnel *v1.Tunnel,
) (target, orphan client.Object, err error) {
	objectKey := client.ObjectKey{Namespace: tunnel.Namespace, Name: buildDaemonName(tunnel)}
	var deployNotExists, daemonSetNotExists bool
	var deployment appsv1.Deployment
	if err := r.Get(ctx, objectKey, &deployment); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, nil, err
		}
		deployNotExists = true
	}
	var daemonSet appsv1.DaemonSet
	if err := r.Get(ctx, objectKey, &daemonSet); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, nil, err
		}
		daemonSetNotExists = true
	}

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
		return nil, nil, WrapError(
			fmt.Errorf("unknown kind: %s", k),
			v1.ConfigReasonFailedToGetExistingConfig,
		)
	}
	return target, orphan, nil
}
