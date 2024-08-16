// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	operatorv1alpha1 "github.com/gardener/gardener/pkg/apis/operator/v1alpha1"
)

const gardenSubsystem = "garden"

type gardenCollector struct {
	runtimeClient client.Reader
	log           logr.Logger

	condition          *prometheus.Desc
	operationSucceeded *prometheus.Desc
}

func newGardenCollector(k8sClient client.Reader, log logr.Logger) *gardenCollector {
	c := &gardenCollector{
		runtimeClient: k8sClient,
		log:           log,
	}
	c.setMetricDefinitions()
	return c
}

func (c *gardenCollector) setMetricDefinitions() {
	c.condition = prometheus.NewDesc(
		prometheus.BuildFQName(metricPrefix, gardenSubsystem, "condition"),
		"Condition state of the Garden.",
		[]string{
			"name",
			"condition",
			"status",
		},
		nil,
	)
	c.operationSucceeded = prometheus.NewDesc(
		prometheus.BuildFQName(metricPrefix, gardenSubsystem, "operation_succeeded"),
		"Returns 1 if the last operation state is Succeeded.",
		[]string{
			"name",
		},
		nil,
	)
}

func (c *gardenCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.condition
	ch <- c.operationSucceeded
}

func (c *gardenCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	gardenList := &operatorv1alpha1.GardenList{}
	if err := c.runtimeClient.List(ctx, gardenList); err != nil {
		c.log.Error(err, "Failed to list gardens")
		return
	}

	for _, garden := range gardenList.Items {
		c.collectConditionMetric(ch, &garden)
		c.collectOperationMetric(ch, &garden)
	}
}

func (c gardenCollector) collectConditionMetric(ch chan<- prometheus.Metric, garden *operatorv1alpha1.Garden) {
	for _, condition := range garden.Status.Conditions {
		if condition.Type == "" {
			continue
		}
		for _, status := range []gardencorev1beta1.ConditionStatus{
			gardencorev1beta1.ConditionFalse,
			gardencorev1beta1.ConditionTrue,
			gardencorev1beta1.ConditionProgressing,
			gardencorev1beta1.ConditionUnknown,
		} {
			val := float64(0)
			if condition.Status == status {
				val = 1
			}
			ch <- prometheus.MustNewConstMetric(
				c.condition,
				prometheus.GaugeValue,
				val,
				[]string{
					garden.Name,
					string(condition.Type),
					string(status),
				}...,
			)
		}
	}
}

func (c *gardenCollector) collectOperationMetric(ch chan<- prometheus.Metric, garden *operatorv1alpha1.Garden) {
	if garden.Status.LastOperation == nil {
		return
	}
	val := float64(0)
	if garden.Status.LastOperation.State == gardencorev1beta1.LastOperationStateSucceeded {
		val = 1
	}
	ch <- prometheus.MustNewConstMetric(
		c.operationSucceeded,
		prometheus.GaugeValue,
		val,
		[]string{
			garden.Name,
		}...,
	)
}
