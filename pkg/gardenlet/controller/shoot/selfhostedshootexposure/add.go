// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package selfhostedshootexposure

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ControllerName is the name of this controller.
const ControllerName = "self-hosted-shoot-exposure"

// AddToManager adds Reconciler to the given manager.
func (r *Reconciler) AddToManager(mgr manager.Manager, gardenCluster cluster.Cluster) error {
	if r.GardenClient == nil {
		r.GardenClient = gardenCluster.GetClient()
	}
	if r.RuntimeClient == nil {
		r.RuntimeClient = mgr.GetClient()
	}
	if r.Clock == nil {
		r.Clock = clock.RealClock{}
	}

	return builder.
		ControllerManagedBy(mgr).
		Named(ControllerName).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		WatchesRawSource(source.Kind[client.Object](mgr.GetCache(),
			&corev1.Node{},
			r.EventHandler(),
			r.NodePredicate(),
		)).
		Complete(r)
}

// EventHandler returns a handler that enqueues the configured shoot key for every relevant Node event.
func (r *Reconciler) EventHandler() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(_ context.Context, _ client.Object) []reconcile.Request {
		return []reconcile.Request{{NamespacedName: r.ShootKey}}
	})
}

// NodePredicate returns a predicate that accepts only control-plane Nodes and ignores updates that don't change
// Node.Status.Addresses.
func (r *Reconciler) NodePredicate() predicate.Predicate {
	return predicate.And(controlPlaneNodePredicate(), nodeAddressChangedPredicate())
}

func controlPlaneNodePredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		_, ok := obj.GetLabels()[nodeRoleControlPlaneLabel]
		return ok
	})
}

func nodeAddressChangedPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(_ event.CreateEvent) bool { return true },
		DeleteFunc:  func(_ event.DeleteEvent) bool { return true },
		GenericFunc: func(_ event.GenericEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode, ok := e.ObjectOld.(*corev1.Node)
			if !ok {
				return true
			}
			newNode, ok := e.ObjectNew.(*corev1.Node)
			if !ok {
				return true
			}
			return !nodeAddressesEqual(oldNode.Status.Addresses, newNode.Status.Addresses)
		},
	}
}

func nodeAddressesEqual(a, b []corev1.NodeAddress) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
