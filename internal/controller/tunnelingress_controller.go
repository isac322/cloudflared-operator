/*
Copyright 2024 Byeonghoon Yoo.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

// TunnelIngressReconciler reconciles a TunnelIngress object
type TunnelIngressReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock  clock.PassiveClock
}

//+kubebuilder:rbac:groups=cloudflared-operator.bhyoo.com,resources=tunnelingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloudflared-operator.bhyoo.com,resources=tunnelingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloudflared-operator.bhyoo.com,resources=tunnelingresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *TunnelIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var ingress v1.TunnelIngress
	if err := r.Get(ctx, req.NamespacedName, &ingress); err != nil {
		l.Error(err, "unable to fetch TunnelIngress")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch ingress.Spec.TunnelRef.Kind {
	case v1.TunnelKindTunnel:
		tunnel, err := r.getTunnelFromIngress(ctx, &ingress)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Tunnel is not found, we'll wait for it to be created
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
			l.Error(err, "unable to fetch Tunnel")
			return ctrl.Result{}, err
		}
		if err = r.reconcileDNSRecord(ctx, &ingress, tunnel); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil

	default:
		return ctrl.Result{}, errors.New("unsupported tunnel type")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TunnelIngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.TunnelIngress{}).
		Complete(r)
}

func (r *TunnelIngressReconciler) buildConditionRecorder(
	ctx context.Context,
	ingress *v1.TunnelIngress,
	condType v1.TunnelIngressConditionType,
) func(err error) error {
	return func(err error) (cause error) {
		defer func() {
			if errors.Is(err, reconcile.TerminalError(nil)) &&
				!errors.Is(cause, reconcile.TerminalError(nil)) {
				cause = reconcile.TerminalError(cause)
			}
		}()

		cause = err
		var reason v1.TunnelIngressConditionReason = ""
		var withReason ReasonedError[v1.TunnelIngressConditionReason]
		if errors.As(err, &withReason) {
			cause = withReason.Cause()
			reason = withReason.Reason
		}

		newCond := v1.TunnelIngressStatusCondition{
			Type:               condType,
			Status:             corev1.ConditionFalse,
			Message:            "",
			Error:              fmt.Sprintf("%+v", cause),
			LastTransitionTime: metav1.Time{Time: r.Clock.Now()},
			Reason:             reason,
		}

		if status, ok := cause.(apierrors.APIStatus); ok || errors.As(cause, &status) {
			newCond.Error = string(status.Status().Reason)
			newCond.Message = status.Status().Message
		}

		if !UpdateConditionIfChanged(&ingress.Status, newCond) {
			return cause
		}

		if updateErr := r.Status().Update(ctx, ingress); updateErr != nil {
			return errors.Join(cause, updateErr)
		}
		return cause
	}
}

func (r *TunnelIngressReconciler) getTunnelFromIngress(ctx context.Context, ingress *v1.TunnelIngress) (*v1.Tunnel, error) {
	if ingress.Spec.TunnelRef.Kind != v1.TunnelKindTunnel {
		return nil, nil
	}

	var tunnel v1.Tunnel
	err := r.Get(ctx, client.ObjectKey{Namespace: ingress.GetNamespace(), Name: ingress.Spec.TunnelRef.Name}, &tunnel)
	if err != nil {
		return nil, err
	}

	return &tunnel, nil
}
