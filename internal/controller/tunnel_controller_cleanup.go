package controller

import (
	"context"
	"errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

func (r *TunnelReconciler) deleteTunnel(ctx context.Context, tunnel *v1.Tunnel) error {
	if tunnel.Status.TunnelID == "" {
		return nil
	}

	client, err := r.getCloudflareClient(ctx, tunnel)
	if err != nil {
		if errors.Is(err, errNotFoundAPITokenKey) || apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return client.DeleteTunnel(ctx, tunnel.Spec.AccountID, tunnel.Status.TunnelID)
}
