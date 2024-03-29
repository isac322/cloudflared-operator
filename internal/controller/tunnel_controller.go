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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

const (
	apiTokenKey        = "token"
	secretNameField    = ".spec.apiTokenSecretRef.name"
	tunnelRefNameField = ".spec.tunnelRef.name"
	tunnelRefKindField = ".spec.tunnelRef.kind"
	fileNameCredential = "credential.json"
	fileNameConfig     = "config.yaml"

	tunnelFinalizerName = "tunnel.cloudflared-operator.bhyoo.com/finalizer"
)

// TunnelReconciler reconciles a Tunnel object
type TunnelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock  clock.PassiveClock
}

//+kubebuilder:rbac:groups=cloudflared-operator.bhyoo.com,resources=tunnels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cloudflared-operator.bhyoo.com,resources=tunnels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cloudflared-operator.bhyoo.com,resources=tunnels/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=apps,resources=daemonset,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=daemonset/status,verbs=get
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tunnel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *TunnelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("tunnelName", req.Name)
	ctx = log.IntoContext(ctx, l)

	var tunnel v1.Tunnel
	if err := r.Get(ctx, req.NamespacedName, &tunnel); err != nil {
		l.Error(err, "unable to fetch Tunnel")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if !tunnel.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&tunnel, tunnelFinalizerName) {
			return ctrl.Result{}, nil
		}
		if err := r.deleteTunnel(ctx, &tunnel); err != nil {
			// if fail to delete the external dependency here, return with error
			// so that it can be retried.
			return ctrl.Result{}, err
		}

		if controllerutil.RemoveFinalizer(&tunnel, tunnelFinalizerName) {
			return ctrl.Result{}, r.Update(ctx, &tunnel)
		}
		return ctrl.Result{}, nil
	}

	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object. This is equivalent
	// to registering our finalizer.
	if !controllerutil.ContainsFinalizer(&tunnel, tunnelFinalizerName) {
		controllerutil.AddFinalizer(&tunnel, tunnelFinalizerName)
		if err := r.Update(ctx, &tunnel); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.reconcileCredential(ctx, &tunnel); err != nil {
		return ctrl.Result{}, err
	}
	credCond := tunnel.Status.GetCondition(v1.TunnelConditionTypeCredential)
	if credCond.Status != corev1.ConditionTrue {
		err := errors.New("inconsistent state")
		l.Error(err, "credential reconciling was succeed with error")
		return ctrl.Result{}, err
	}

	tunnelConfig, err := r.reconcileConfig(ctx, &tunnel)
	if err != nil {
		return ctrl.Result{}, err
	}
	configCond := tunnel.Status.GetCondition(v1.TunnelConditionTypeConfig)
	if configCond.Status != corev1.ConditionTrue {
		err := errors.New("inconsistent state")
		l.Error(err, "config reconciling was succeed with error")
		return ctrl.Result{}, err
	}

	if err := r.reconcileDaemon(ctx, &tunnel, tunnelConfig); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TunnelReconciler) findObjectsForSecret(ctx context.Context, secret client.Object) []reconcile.Request {
	attachedTunnels := &v1.TunnelList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(secretNameField, secret.GetName()),
		Namespace:     secret.GetNamespace(),
	}
	err := r.List(ctx, attachedTunnels, listOps)
	if err != nil {
		l := log.FromContext(ctx)
		l.Error(err, "failed to listing tunnels matched with secret name")
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(attachedTunnels.Items))
	for i, item := range attachedTunnels.Items {
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *TunnelReconciler) findObjectsForTunnelIngress(
	_ context.Context,
	tunnelIngress client.Object,
) []reconcile.Request {
	ingress := tunnelIngress.(*v1.TunnelIngress)
	if ingress.Spec.TunnelRef.Kind != v1.TunnelKindTunnel {
		return nil
	}
	ingressCondition := ingress.Status.GetCondition(v1.TunnelIngressConditionTypeDNSRecord)
	if ingressCondition.Status != corev1.ConditionTrue {
		return nil
	}

	return []reconcile.Request{{NamespacedName: types.NamespacedName{
		Name:      ingress.Spec.TunnelRef.Name,
		Namespace: tunnelIngress.GetNamespace(),
	}}}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TunnelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()

	if err := mgr.GetFieldIndexer().IndexField(
		ctx,
		&v1.Tunnel{},
		secretNameField,
		func(rawObj client.Object) []string {
			tunnel := rawObj.(*v1.Tunnel)
			if tunnel.Spec.APITokenSecretRef.Name == "" {
				return nil
			}
			return []string{tunnel.Spec.APITokenSecretRef.Name}
		},
	); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(
		ctx,
		&v1.TunnelIngress{},
		tunnelRefNameField,
		func(rawObj client.Object) []string {
			tunnelIngress := rawObj.(*v1.TunnelIngress)
			if tunnelIngress.Spec.TunnelRef.Name == "" {
				return nil
			}
			return []string{tunnelIngress.Spec.TunnelRef.Name}
		},
	); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(
		ctx,
		&v1.TunnelIngress{},
		tunnelRefKindField,
		func(rawObj client.Object) []string {
			tunnelIngress := rawObj.(*v1.TunnelIngress)
			if tunnelIngress.Spec.TunnelRef.Kind == "" {
				return nil
			}
			return []string{string(tunnelIngress.Spec.TunnelRef.Kind)}
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Tunnel{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForSecret),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Watches(
			&v1.TunnelIngress{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForTunnelIngress),
		).
		Complete(r)
}

func (r *TunnelReconciler) buildConditionRecorder(
	ctx context.Context,
	tunnel *v1.Tunnel,
	condType v1.TunnelConditionType,
) func(err error) error {
	return func(err error) (cause error) {
		defer func() {
			if errors.Is(err, reconcile.TerminalError(nil)) &&
				!errors.Is(cause, reconcile.TerminalError(nil)) {
				cause = reconcile.TerminalError(cause)
			}
		}()

		cause = err
		var reason v1.TunnelConditionReason = ""
		var withReason ReasonedError[v1.TunnelConditionReason]
		if errors.As(err, &withReason) {
			cause = withReason.Cause()
			reason = withReason.Reason
		}

		newCond := v1.TunnelStatusCondition{
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

		if !UpdateConditionIfChanged(&tunnel.Status, newCond) {
			return cause
		}

		if updateErr := r.Status().Update(ctx, tunnel); updateErr != nil {
			return errors.Join(cause, updateErr)
		}
		return cause
	}
}

func (r *TunnelReconciler) updateConditionIfDiff(
	ctx context.Context,
	tunnel *v1.Tunnel,
	cond v1.TunnelStatusCondition,
) error {
	if UpdateConditionIfChanged(&tunnel.Status, cond) {
		return r.Status().Update(ctx, tunnel)
	}
	return nil
}
