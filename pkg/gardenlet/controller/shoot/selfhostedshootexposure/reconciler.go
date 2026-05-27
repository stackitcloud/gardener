// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package selfhostedshootexposure

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1beta1helper "github.com/gardener/gardener/pkg/api/core/v1beta1/helper"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	extensionsselfhostedshootexposure "github.com/gardener/gardener/pkg/component/extensions/selfhostedshootexposure"
)

const nodeRoleControlPlaneLabel = "node-role.kubernetes.io/control-plane"

// Reconciler keeps the SelfHostedShootExposure endpoints (or the external DNSRecord values in DNS-only setups) in
// sync with the self-hosted shoot's control-plane Node addresses.
type Reconciler struct {
	GardenClient  client.Client
	RuntimeClient client.Client
	ShootKey      types.NamespacedName
	Clock         clock.PassiveClock
}

// Reconcile recomputes the control-plane endpoints from the current Node state and patches the SelfHostedShootExposure
// resource or the external DNSRecord accordingly.
func (r *Reconciler) Reconcile(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
	log := logf.FromContext(ctx).WithName(ControllerName)

	shoot := &gardencorev1beta1.Shoot{}
	if err := r.GardenClient.Get(ctx, r.ShootKey, shoot); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("failed getting Shoot: %w", err)
	}

	if !v1beta1helper.IsShootSelfHosted(shoot.Spec.Provider.Workers) {
		return reconcile.Result{}, nil
	}

	enabled, err := r.endpointUpdatesEnabled(ctx, shoot)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !enabled {
		log.V(1).Info("Endpoint updates disabled by ControllerRegistration")
		return reconcile.Result{}, nil
	}

	nodes, err := r.listControlPlaneNodes(ctx)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed listing control-plane nodes: %w", err)
	}

	if v1beta1helper.HasExtensionExposure(shoot) {
		if err := r.reconcileSelfHostedShootExposure(ctx, log, shoot, nodes); err != nil {
			return reconcile.Result{}, err
		}
	} else {
		// TODO: implement DNS-only path (patch external DNSRecord values via FilterAddressesByIPFamily).
		log.V(1).Info("DNS-only endpoint updates not yet implemented")
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) endpointUpdatesEnabled(ctx context.Context, shoot *gardencorev1beta1.Shoot) (bool, error) {
	if !v1beta1helper.HasExtensionExposure(shoot) {
		return true, nil
	}

	registrations := &gardencorev1beta1.ControllerRegistrationList{}
	if err := r.GardenClient.List(ctx, registrations); err != nil {
		return false, fmt.Errorf("failed listing ControllerRegistrations: %w", err)
	}
	return v1beta1helper.SelfHostedShootExposureEndpointUpdateEnabled(registrations.Items, shseExtensionType(shoot)), nil
}

func (r *Reconciler) listControlPlaneNodes(ctx context.Context) ([]corev1.Node, error) {
	nodes := &corev1.NodeList{}
	if err := r.RuntimeClient.List(ctx, nodes, client.MatchingLabels{nodeRoleControlPlaneLabel: ""}); err != nil {
		return nil, err
	}
	return nodes.Items, nil
}

func (r *Reconciler) reconcileSelfHostedShootExposure(ctx context.Context, log logr.Logger, shoot *gardencorev1beta1.Shoot, nodes []corev1.Node) error {
	endpoints := make([]extensionsv1alpha1.ControlPlaneEndpoint, 0, len(nodes))
	for i := range nodes {
		node := nodes[i].DeepCopy()
		if len(node.Status.Addresses) == 0 {
			return fmt.Errorf("node %q has no addresses", node.Name)
		}
		endpoints = append(endpoints, extensionsv1alpha1.ControlPlaneEndpoint{
			NodeName:  node.Name,
			Addresses: node.Status.Addresses,
		})
	}

	values := &extensionsselfhostedshootexposure.Values{
		Name:      shoot.Name,
		Namespace: v1beta1helper.ControlPlaneNamespaceForShoot(shoot),
		Type:      shseExtensionType(shoot),
		Endpoints: endpoints,
	}
	if v1beta1helper.HasManagedInfrastructure(shoot) {
		values.CredentialsRef = &corev1.ObjectReference{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
			Name:       v1beta1constants.SecretNameCloudProvider,
			Namespace:  values.Namespace,
		}
	}

	c := extensionsselfhostedshootexposure.New(log, r.RuntimeClient, values)
	if r.Clock != nil {
		c.Clock = r.Clock
	}
	return c.Deploy(ctx)
}

func shseExtensionType(shoot *gardencorev1beta1.Shoot) string {
	pool := v1beta1helper.ControlPlaneWorkerPoolForShoot(shoot.Spec.Provider.Workers)
	return ptr.Deref(pool.ControlPlane.Exposure.Extension.Type, shoot.Spec.Provider.Type)
}
