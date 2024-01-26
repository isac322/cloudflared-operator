package controller

import (
	"context"
	"errors"

	"github.com/goccy/go-json"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
	"github.com/isac322/cloudflared-operator/internal/cloudflare"
)

var (
	errNotFoundAPITokenKey = errors.New("api token key is not found")
)

func (r *TunnelReconciler) reconcileCredential(ctx context.Context, tunnel *v1.Tunnel) error {
	recordConditionFrom := r.buildConditionRecorder(ctx, tunnel, v1.TunnelConditionTypeCredential)

	cfClient, err := r.getCloudflareClient(ctx, tunnel)
	if err != nil {
		return recordConditionFrom(err)
	}

	var dirtyStatus bool

	var credentialSecret corev1.Secret
	err = r.Get(
		ctx,
		client.ObjectKey{Namespace: tunnel.Namespace, Name: tunnel.Spec.CredentialSecretName()},
		&credentialSecret,
	)
	switch {
	case err == nil:
		tunnelID, err := r.verifyAndUpdateCred(
			ctx,
			&credentialSecret,
			cfClient,
			tunnel.Spec.AccountID,
			tunnel.Spec.Name,
		)
		if err != nil {
			return recordConditionFrom(err)
		}

		dirtyStatus = tunnel.Status.TunnelID != tunnelID
		tunnel.Status.TunnelID = tunnelID

	case apierrors.IsNotFound(err):
		dirtyStatus = true
		if err := r.updateConditionIfDiff(
			ctx,
			tunnel,
			v1.TunnelStatusCondition{
				Type:               v1.TunnelConditionTypeCredential,
				Status:             corev1.ConditionFalse,
				LastTransitionTime: metav1.Time{Time: r.Clock.Now()},
				Reason:             v1.CredentialReasonCreating,
			},
		); err != nil {
			return err
		}
		tunnel.Status.TunnelID, err = r.createCredSecret(ctx, cfClient, tunnel)
		if err != nil {
			return recordConditionFrom(err)
		}

	// unknown errors
	default:
		return recordConditionFrom(WrapError(err, v1.CredentialReasonFailedToGetExistingCredential))
	}

	if dirtyStatus || SetTunnelConditionIfDiff(tunnel, v1.TunnelStatusCondition{
		Type:               v1.TunnelConditionTypeCredential,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: r.Clock.Now()},
	}) {
		return r.Status().Update(ctx, tunnel)
	}
	return nil
}

func (r *TunnelReconciler) verifyAndUpdateCred(
	ctx context.Context,
	credentialSecret *corev1.Secret,
	cfClient cloudflare.Client,
	accountID, tunnelName string,
) (string, error) {
	l := log.FromContext(ctx)

	tunnelID, err := verifyCredSecret(ctx, credentialSecret, cfClient)
	if err != nil || tunnelID != "" {
		return tunnelID, err
	}

	l.Info("outdated credential. reconciling it...")
	tunnelID, err = fillCredSecret(ctx, credentialSecret, cfClient, accountID, tunnelName)
	if err != nil {
		return "", err
	}
	if err = r.Update(ctx, credentialSecret); err != nil {
		return "", err
	}
	return tunnelID, nil
}

func (r *TunnelReconciler) createCredSecret(
	ctx context.Context,
	cfClient cloudflare.Client,
	tunnel *v1.Tunnel,
) (string, error) {
	credentialSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnel.Spec.CredentialSecretName(),
			Namespace: tunnel.Namespace,
		},
		Immutable: ptr.To(true),
	}
	if err := ctrl.SetControllerReference(tunnel, credentialSecret, r.Scheme); err != nil {
		return "", WrapError(err, v1.CredentialReasonFailedToCreateSecret)
	}
	tunnelID, err := fillCredSecret(ctx, credentialSecret, cfClient, tunnel.Spec.AccountID, tunnel.Spec.Name)
	if err != nil {
		return "", err
	}
	if err := r.Create(ctx, credentialSecret); err != nil {
		return "", WrapError(err, v1.CredentialReasonFailedToCreateSecret)
	}

	return tunnelID, nil
}

func fillCredSecret(
	ctx context.Context,
	credentialSecret *corev1.Secret,
	cfClient cloudflare.Client,
	accountID, tunnelName string,
) (string, error) {
	credential, err := cfClient.GetOrCreateTunnel(ctx, accountID, tunnelName)
	if err != nil {
		return "", WrapError(err, v1.CredentialReasonFailedToCreateTunnelOnCF)
	}
	marshaledCredential, err := json.MarshalNoEscape(credential)
	if err != nil {
		return "", WrapError(err, v1.CredentialReasonInvalidCredential)
	}

	delete(credentialSecret.StringData, fileNameCredential)
	credentialSecret.Data[fileNameCredential] = marshaledCredential
	return credential.TunnelID, nil
}

func verifyCredSecret(
	ctx context.Context,
	credentialSecret *corev1.Secret,
	cfClient cloudflare.Client,
) (string, error) {
	bytesCredential, bytesOk := credentialSecret.Data[fileNameCredential]
	strCredential, strOk := credentialSecret.StringData[fileNameCredential]
	if !bytesOk && !strOk {
		return "", nil
	}

	if strOk {
		bytesCredential = []byte(strCredential)
	}

	var credential cloudflare.TunnelCredential
	if err := json.UnmarshalNoEscape(bytesCredential, &credential); err != nil {
		return "", nil
	}

	isValid, err := cfClient.ValidateTunnelCredential(ctx, credential)
	if err != nil {
		return "", WrapError(err, v1.CredentialReasonFailedToValidate)
	}
	if isValid {
		return credential.TunnelID, nil
	}
	return "", nil
}

func (r *TunnelReconciler) getCloudflareClient(ctx context.Context, tunnel *v1.Tunnel) (cloudflare.Client, error) {
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
		err = WrapError(err, v1.CredentialReasonNoToken)
		if apierrors.IsNotFound(err) {
			err = reconcile.TerminalError(err)
		}
		return nil, err
	}

	secretKey := ptr.Deref(tunnel.Spec.APITokenSecretRef.Key, apiTokenKey)
	bytesToken, exists := secret.Data[secretKey]
	var token string
	if !exists {
		token, exists = secret.StringData[secretKey]
		if !exists {
			return nil, WrapError(errNotFoundAPITokenKey, v1.CredentialReasonNoToken)
		}
	} else {
		token = string(bytesToken)
	}

	cli, err := cloudflare.NewClient(token)
	if err != nil {
		l.Error(err, "failed to connect to Cloudflare")
		return nil, WrapError(errNotFoundAPITokenKey, v1.CredentialReasonFailedToConnectCF)
	}
	return cli, nil
}
