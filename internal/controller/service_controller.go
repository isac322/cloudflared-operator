/*
Copyright 2024 Jinha Jeong.

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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	TunnelNameAnnotation = "tunnel.name"
	HostNameAnnotation   = "host.name"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme // 리소스의 타입 정보를 관리
}

//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=services/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Service object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	instance := &corev1.Service{} //nolint:staticcheck
	if instance != nil {          //nolint:staticcheck
		log.Info("service exists", "annotated tunnel name", instance.Annotations["tunnel.name"])
	}
	log.Info("Reconciling Service", "service", req.NamespacedName)
	// TODO(user): your logic here
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldAnnotations := e.ObjectOld.GetAnnotations()
				newAnnotations := e.ObjectNew.GetAnnotations()

				// if no changes in annotations, return false
				if reflect.DeepEqual(oldAnnotations, newAnnotations) {
					return false
				}

				tunnelName, tunnelNameExists := newAnnotations[TunnelNameAnnotation]
				hostName, hostNameExists := newAnnotations[HostNameAnnotation]

				// both are required to create a proper TunnelIngress object
				if !tunnelNameExists || !hostNameExists {
					return false
				}

				// if changes in annotations is not about the purpose of creating a TunnelIngress object, return false
				if tunnelName == oldAnnotations[TunnelNameAnnotation] && hostName == oldAnnotations[HostNameAnnotation] {
					return false
				}
				return true
			},
			CreateFunc: func(e event.CreateEvent) bool {
				annotations := e.Object.GetAnnotations()
				// check if all required annotations are present
				_, tunnelNameExists := annotations[TunnelNameAnnotation]
				_, hostNameExists := annotations[HostNameAnnotation]
				if !tunnelNameExists || !hostNameExists {
					return false
				}
				return true
			},
		}).
		Complete(r)
}
