package controller

import (
	"context"
	"errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

func (r *TunnelIngressReconciler) deleteTunnelIngress(ctx context.Context, ingress *v1.TunnelIngress) error {
	l := log.FromContext(ctx)

	if ingress.Spec.Hostname == nil {
		return nil
	}

	var tunnel *v1.Tunnel
	switch ingress.Spec.TunnelRef.Kind {
	case v1.TunnelKindTunnel:
		var err error
		tunnel, err = r.getTunnelFromIngress(ctx, ingress)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			l.Error(err, "unable to fetch Tunnel")
			return err
		}

	default:
		return errors.New("unsupported tunnel type")
	}

	if tunnel.Status.TunnelID == "" {
		return nil
	}

	cfClient, err := r.getCloudflareClient(ctx, tunnel)
	if err != nil {
		return err
	}

	return cfClient.DeleteDNSRecord(ctx, tunnel.Spec.AccountID, *ingress.Spec.Hostname)
}
