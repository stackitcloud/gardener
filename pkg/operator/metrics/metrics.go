// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	runtimemetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricPrefix = "gardener_operator"
)

type runnable struct {
	gardenCollector *gardenCollector
}

// AddToManager adds the custom metrics collectors to the metrics registry. This is done "inside" a `manager.Runnable`,
// because that guarantees that the cache informers are synced, before the metrics are added / scraped for the first
// time.
func AddToManager(_ context.Context, mgr manager.Manager) error {
	k8sClient := mgr.GetClient()
	return mgr.Add(&runnable{
		gardenCollector: newGardenCollector(k8sClient, mgr.GetLogger().WithName("operator-metrics")),
	})
}

func (r *runnable) Start(_ context.Context) error {
	runtimemetrics.Registry.MustRegister(r.gardenCollector)
	return nil
}
