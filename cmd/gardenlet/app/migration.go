// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener/pkg/utils/flow"
	versionutils "github.com/gardener/gardener/pkg/utils/version"
)

func (g *garden) runMigrations(ctx context.Context, log logr.Logger, _ client.Client) error {
	log.Info("Migrating deprecated failure-domain.beta.kubernetes.io labels to topology.kubernetes.io")
	if err := migrateDeprecatedTopologyLabels(ctx, log, g.mgr.GetClient(), g.mgr.GetConfig()); err != nil {
		return err
	}

	/*
		As we already run a modified apiserver-proxy component which uses the reversed-vpn port for connecting to the shoot controlplane,
			we don't need to validate if every shoot was reconciled with g/g 1.113 and has a reconfigured apiserver-proxy

		We also don't want to expose the vulnerable port again, requiring us the instantly enable the new "RemoveAPIServerProxyLegacyPort" feature gate when upgrading to g/g 1.113.
		This would not be possible with the feature gate validation intact.

		This check was only implemented upstream to ensure that the legacy proxy-protocol port (8443) will not be removed until every shoot switched to the new port

		if features.DefaultFeatureGate.Enabled(features.RemoveAPIServerProxyLegacyPort) {
			if err := verifyRemoveAPIServerProxyLegacyPortFeatureGate(ctx, gardenClient, g.config.SeedConfig.Name); err != nil {
				return err
			}
		}*/

	return nil
}

// TODO: Remove this function when Kubernetes 1.27 support gets dropped.
func migrateDeprecatedTopologyLabels(ctx context.Context, log logr.Logger, seedClient client.Client, restConfig *rest.Config) error {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed creating discovery client: %w", err)
	}

	version, err := discoveryClient.ServerVersion()
	if err != nil {
		return fmt.Errorf("failed reading the server version of seed cluster: %w", err)
	}

	seedVersion, err := semver.NewVersion(version.GitVersion)
	if err != nil {
		return fmt.Errorf("failed parsing server version to semver: %w", err)
	}

	//  PV node affinities were immutable until Kubernetes 1.27, see https://github.com/kubernetes/kubernetes/pull/115391
	if !versionutils.ConstraintK8sGreaterEqual127.Check(seedVersion) {
		return nil
	}

	persistentVolumeList := &corev1.PersistentVolumeList{}
	if err := seedClient.List(ctx, persistentVolumeList); err != nil {
		return fmt.Errorf("failed listing persistent volumes for migrating deprecated topology labels: %w", err)
	}

	var taskFns []flow.TaskFn

	for _, pv := range persistentVolumeList.Items {
		persistentVolume := pv

		taskFns = append(taskFns, func(ctx context.Context) error {
			patch := client.MergeFrom(persistentVolume.DeepCopy())

			if persistentVolume.Spec.NodeAffinity == nil {
				// when PV is very old and has no node affinity, we just replace the topology labels
				if v, ok := persistentVolume.Labels[corev1.LabelFailureDomainBetaRegion]; ok {
					persistentVolume.Labels[corev1.LabelTopologyRegion] = v
				}
				if v, ok := persistentVolume.Labels[corev1.LabelFailureDomainBetaZone]; ok {
					persistentVolume.Labels[corev1.LabelTopologyZone] = v
				}
			} else if persistentVolume.Spec.NodeAffinity.Required != nil {
				// when PV has node affinity then we do not need the labels but just need to replace the topology keys
				// in the node selector term match expressions
				for i, term := range persistentVolume.Spec.NodeAffinity.Required.NodeSelectorTerms {
					for j, expression := range term.MatchExpressions {
						if expression.Key == corev1.LabelFailureDomainBetaRegion {
							persistentVolume.Spec.NodeAffinity.Required.NodeSelectorTerms[i].MatchExpressions[j].Key = corev1.LabelTopologyRegion
						}

						if expression.Key == corev1.LabelFailureDomainBetaZone {
							persistentVolume.Spec.NodeAffinity.Required.NodeSelectorTerms[i].MatchExpressions[j].Key = corev1.LabelTopologyZone
						}
					}
				}
			}

			// either new topology labels were added above, or node affinity keys were adjusted
			// in both cases, the old, deprecated topology labels are no longer needed and can be removed
			delete(persistentVolume.Labels, corev1.LabelFailureDomainBetaRegion)
			delete(persistentVolume.Labels, corev1.LabelFailureDomainBetaZone)

			// prevent sending empty patches
			if data, err := patch.Data(&persistentVolume); err != nil {
				return fmt.Errorf("failed getting patch data for PV %s: %w", persistentVolume.Name, err)
			} else if string(data) == `{}` {
				return nil
			}

			log.Info("Migrating deprecated topology labels", "persistentVolumeName", persistentVolume.Name)
			return seedClient.Patch(ctx, &persistentVolume, patch)
		})
	}

	return flow.Parallel(taskFns...)(ctx)
}

// TODO(Wieneo): Remove this function when feature gate RemoveAPIServerProxyLegacyPort is removed
/*
func verifyRemoveAPIServerProxyLegacyPortFeatureGate(ctx context.Context, gardenClient client.Client, seedName string) error {
	shootList := &gardencorev1beta1.ShootList{}
	if err := gardenClient.List(ctx, shootList); err != nil {
		return err
	}

	for _, k := range shootList.Items {
		if specSeedName, statusSeedName := gardenerutils.GetShootSeedNames(&k); gardenerutils.GetResponsibleSeedName(specSeedName, statusSeedName) != seedName {
			continue
		}

		// we need to ignore shoots under the following conditions:
		// - it is workerless
		// - it is not yet picked up by gardenlet or still in phase "Creating"
		//
		// this is needed bcs. the constraint "ShootAPIServerProxyUsesHTTPProxy" is only set once the apiserver-proxy component is deployed to the shoot
		// this will never happen if the shoot is workerless or the component could still be missing, if the gardenlet is restarted during the creation of a shoot
		if v1beta1helper.IsWorkerless(&k) {
			continue
		}

		if k.Status.LastOperation == nil || (k.Status.LastOperation.Type == gardencorev1beta1.LastOperationTypeCreate && k.Status.LastOperation.State != gardencorev1beta1.LastOperationStateSucceeded) {
			continue
		}

		if cond := v1beta1helper.GetCondition(k.Status.Constraints, gardencorev1beta1.ShootAPIServerProxyUsesHTTPProxy); cond == nil || cond.Status != gardencorev1beta1.ConditionTrue {
			return errors.New("the `proxy` port on the istio ingress gateway cannot be removed until all api server proxies in all shoots on this seed have been reconfigured to use the `tls-tunnel` port instead, i.e., the `RemoveAPIServerProxyLegacyPort` feature gate can only be enabled once all shoots have the `APIServerProxyUsesHTTPProxy` constraint with status `true`")
		}
	}

	return nil
}
*/
