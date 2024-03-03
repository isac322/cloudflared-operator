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
	"fmt"
	"reflect"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1 "github.com/isac322/cloudflared-operator/api/v1"
)

const (
	HostNameAnnotation          = "cloudflared-operator.bhyoo.com/host-name"
	PortTunnelMappingAnnotation = "cloudflared-operator.bhyoo.com/port-"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=services/finalizers,verbs=update
//+kubebuilder:rbac:groups=cloudflared-operator.bhyoo.com,resources=tunnels,verbs=get;list;watch

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
	logger := log.FromContext(ctx)

	var service corev1.Service
	if err := r.Get(ctx, req.NamespacedName, &service); err != nil {
		logger.Error(err, "unable to fetch Service")
		return ctrl.Result{}, err
	}

	annotations := service.Annotations
	hostName := annotations[HostNameAnnotation]
	ports := service.Spec.Ports
	var tunnel v1.Tunnel

	for _, port := range ports {
		tunnelName := annotations[PortTunnelMappingAnnotation+port.Name]
		if err := r.Get(ctx, client.ObjectKey{Name: tunnelName, Namespace: service.Namespace}, &tunnel); err != nil {
			logger.Error(err, "No such Tunnel object exists with given name")
			continue
		}
		tunnelIngress := createTunnelIngress(tunnelName, hostName, service, port)
		if err := ctrl.SetControllerReference(&tunnel, tunnelIngress, r.Scheme); err != nil {
			// TODO: error handling
			return ctrl.Result{}, err
		}
		if err := ctrl.SetControllerReference(&service, tunnelIngress, r.Scheme); err != nil {
			// TODO: error handling
			return ctrl.Result{}, err
		}
		if err := r.Update(ctx, tunnelIngress); err != nil {
			// TODO: error handling
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, tunnelIngress); err != nil {
			logger.Error(err, "Failed to create TunnelIngress")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func createTunnelIngress(tunnelName, hostName string, service corev1.Service, port corev1.ServicePort) *v1.TunnelIngress {
	portStr := portToString(port)
	tunnelIngress := &v1.TunnelIngress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelName,
			Namespace: service.Namespace,
		},
		Spec: v1.TunnelIngressSpec{
			TunnelConfigIngress: v1.TunnelConfigIngress{
				Hostname: &hostName,
				Service:  fmt.Sprintf("%s.%s.svc.cluster.local:%s", service.Name, service.Namespace, portStr),
			},
			TunnelRef: v1.TunnelRef{
				Kind: v1.TunnelKindTunnel,
				Name: tunnelName,
			},
		},
	}
	return tunnelIngress
}

func portToString(port corev1.ServicePort) string {
	portNum := int64(port.Port)
	portStr := strconv.FormatInt(portNum, 10)
	return portStr
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		//Watches(
		//	&v1.Tunnel{},
		//	handler.EnqueueRequestsFromMapFunc(r.findRelatedServiceObject),
		//	//builder.WithPredicates(onlyResponseOnTunnelCreation()),
		//).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: checkServiceUpdateForReconciliation(),
			CreateFunc: checkServiceCreationForReconciliation(),
		}).
		Complete(r)
}

//func onlyResponseOnTunnelCreation() predicate.Funcs {
//	return predicate.Funcs{
//		CreateFunc: func(e event.CreateEvent) bool {
//			return true
//		},
//		DeleteFunc: func(e event.DeleteEvent) bool {
//			return false
//		},
//		UpdateFunc: func(e event.UpdateEvent) bool {
//			return false
//		},
//		GenericFunc: func(e event.GenericEvent) bool {
//			return false
//		},
//	}
//}

func checkServiceCreationForReconciliation() func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) bool {
		annotations := e.Object.GetAnnotations()

		if _, hostNameExists := findHostName(annotations); !hostNameExists {
			return false
		}

		newService, isService := e.Object.(*corev1.Service)
		if !isService {
			return false
		}

		for _, port := range newService.Spec.Ports {
			// check if all required annotations are present
			if _, tunnelNameExists := findTunnelMappingByPortName(annotations, port.Name); tunnelNameExists {
				return true
			}
		}
		return false
	}
}

func checkServiceUpdateForReconciliation() func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) bool {
		oldAnnotations := e.ObjectOld.GetAnnotations()
		newAnnotations := e.ObjectNew.GetAnnotations()

		// if no changes in annotations, return false
		if reflect.DeepEqual(oldAnnotations, newAnnotations) {
			return false
		}

		hostName, hostNameExists := findHostName(newAnnotations)
		if !hostNameExists {
			return false
		}

		updatedService, isService := e.ObjectNew.(*corev1.Service)
		if !isService {
			return false
		}

		for _, port := range updatedService.Spec.Ports {
			portName := port.Name
			// check if portToTunnel mapping annotation exists with given port name
			portToTunnel, mappingExists := findTunnelMappingByPortName(newAnnotations, portName)
			if !mappingExists {
				// continue to see if other ports have correct annotations to create TunnelIngress object
				continue
			}
			// check if there are any changes in tunnel mapping or host name for any port.
			if portToTunnel != oldAnnotations[PortTunnelMappingAnnotation+portName] || hostName != oldAnnotations[HostNameAnnotation] {
				return true
			}
		}
		return false
	}
}

//func (r *ServiceReconciler) findRelatedServiceObject(ctx context.Context, tunnel client.Object) []reconcile.Request {
//	tunnelName := tunnel.GetName()
//	tunnelNamespace := tunnel.GetNamespace()
//
//	var list corev1.ServiceList
//	if err := r.List(
//		ctx,
//		&list,
//		client.InNamespace(tunnelNamespace),
//	); err != nil {
//		return []reconcile.Request{}
//	}
//
//	for _, item := range list.Items {
//		annotations := item.Annotations
//		if _, hostNameExists := findHostName(annotations); !hostNameExists {
//			continue
//		}
//		for _, port := range item.Spec.Ports {
//			if tunnelName == annotations[PortTunnelMappingAnnotation+port.Name] {
//				// if there's a service that already defines annotations for that specific Tunnel,
//				// create Reconcile request
//				return []reconcile.Request{{
//					NamespacedName: types.NamespacedName{
//						Name:      item.Name,
//						Namespace: item.Namespace,
//					},
//				}}
//			}
//		}
//	}
//	return []reconcile.Request{}
//}

// Check if the service has a hostname annotation ("cloudflared-operator.bhyoo.com/host-name") specified.
// This annotation is crucial for identifying the target host name for the TunnelIngress creation.
// Without a specified host name, the TunnelIngress cannot be created
func findHostName(newAnnotations map[string]string) (string, bool) {
	hostName, hostNameExists := newAnnotations[HostNameAnnotation]
	return hostName, hostNameExists
}

func findTunnelMappingByPortName(annotations map[string]string, portName string) (string, bool) {
	portToTunnel, mappingExists := annotations[PortTunnelMappingAnnotation+portName]
	return portToTunnel, mappingExists
}
