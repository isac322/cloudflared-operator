package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
	"github.com/isac322/cloudflared-operator/internal/cloudflare"
)

func (r *TunnelIngressReconciler) reconcileDNSRecord(
	ctx context.Context,
	ingress *v1.TunnelIngress,
	tunnel *v1.Tunnel,
) error {
	recordConditionFrom := r.buildConditionRecorder(ctx, ingress, v1.TunnelIngressConditionTypeDNSRecord)

	targetDomain := ingress.Spec.Hostname
	if targetDomain == nil {
		if UpdateConditionIfChanged(&ingress.Status, v1.TunnelIngressStatusCondition{
			Type:               v1.TunnelIngressConditionTypeDNSRecord,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Time{Time: r.Clock.Now()},
		}) {
			return r.Status().Update(ctx, ingress)
		}
		return nil
	}

	cfClient, err := r.getCloudflareClient(ctx, tunnel)
	if err != nil {
		return recordConditionFrom(err)
	}

	err = cfClient.CreateRoute(
		ctx,
		tunnel.Spec.AccountID,
		tunnel.Status.TunnelID,
		*targetDomain,
		ingress.Spec.OverwriteExistingDNS,
	)
	if err != nil {
		return recordConditionFrom(WrapError(err, v1.DNSRecordReasonFailedToCreateRecord))
	}

	if UpdateConditionIfChanged(&ingress.Status, v1.TunnelIngressStatusCondition{
		Type:               v1.TunnelIngressConditionTypeDNSRecord,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: r.Clock.Now()},
	}) {
		return r.Status().Update(ctx, ingress)
	}
	return nil
}

func (r *TunnelIngressReconciler) getCloudflareClient(
	ctx context.Context,
	tunnel *v1.Tunnel,
) (cloudflare.Client, error) {
	l := log.FromContext(ctx)

	var secret corev1.Secret
	if err := r.Get(
		ctx,
		client.ObjectKey{
			Namespace: ptr.Deref(tunnel.Spec.APITokenSecretRef.Namespace, tunnel.Namespace),
			Name:      tunnel.Spec.APITokenSecretRef.Name,
		},
		&secret,
	); err != nil {
		l.Error(err, "unable to fetch apiTokenSecretRef")
		err = WrapError(err, v1.DNSRecordReasonNoToken)
		if apierrors.IsNotFound(err) {
			err = reconcile.TerminalError(err)
		}
		return nil, err
	}

	secretKey := ptr.Deref(tunnel.Spec.APITokenSecretRef.Key, apiTokenKey)
	bytesToken, exists := GetDataFromSecret(&secret, secretKey)
	if !exists {
		return nil, WrapError(errNotFoundAPITokenKey, v1.DNSRecordReasonNoToken)
	}

	cli, err := cloudflare.NewClient(string(bytesToken))
	if err != nil {
		l.Error(err, "failed to create Cloudflare client")
		return nil, WrapError(err, v1.DNSRecordReasonFailedToConnectCF)
	}
	return cli, nil
}
