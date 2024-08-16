// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	operatorv1alpha1 "github.com/gardener/gardener/pkg/apis/operator/v1alpha1"
)

var _ = Describe("Garden metrics", func() {
	var (
		ctx       context.Context
		k8sClient client.Client

		c      prometheus.Collector
		garden *operatorv1alpha1.Garden
	)

	BeforeEach(func() {
		testScheme := runtime.NewScheme()
		Expect(operatorv1alpha1.AddToScheme(testScheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().
			WithScheme(testScheme).
			WithStatusSubresource(&operatorv1alpha1.Garden{}).
			Build()

		c = newGardenCollector(k8sClient, logr.Discard())

		garden = &operatorv1alpha1.Garden{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
		}
		Expect(k8sClient.Create(ctx, garden)).To(Succeed())

		garden.Status = operatorv1alpha1.GardenStatus{
			LastOperation: &gardencorev1beta1.LastOperation{
				State: gardencorev1beta1.LastOperationStateSucceeded,
			},
			Conditions: []gardencorev1beta1.Condition{
				{Type: operatorv1alpha1.RuntimeComponentsHealthy, Status: gardencorev1beta1.ConditionTrue},
				{Type: operatorv1alpha1.VirtualComponentsHealthy, Status: gardencorev1beta1.ConditionFalse},
			},
		}
		Expect(k8sClient.Status().Update(ctx, garden)).To(Succeed())

	})

	It("should collect condition metrics", func() {
		expected := strings.NewReader(`# HELP gardener_operator_garden_condition Condition state of the Garden.
# TYPE gardener_operator_garden_condition gauge
gardener_operator_garden_condition{condition="RuntimeComponentsHealthy",name="foo",status="False"} 0
gardener_operator_garden_condition{condition="RuntimeComponentsHealthy",name="foo",status="Progressing"} 0
gardener_operator_garden_condition{condition="RuntimeComponentsHealthy",name="foo",status="True"} 1
gardener_operator_garden_condition{condition="RuntimeComponentsHealthy",name="foo",status="Unknown"} 0
gardener_operator_garden_condition{condition="VirtualComponentsHealthy",name="foo",status="False"} 1
gardener_operator_garden_condition{condition="VirtualComponentsHealthy",name="foo",status="Progressing"} 0
gardener_operator_garden_condition{condition="VirtualComponentsHealthy",name="foo",status="True"} 0
gardener_operator_garden_condition{condition="VirtualComponentsHealthy",name="foo",status="Unknown"} 0
`)

		Expect(
			testutil.CollectAndCompare(c, expected, "gardener_operator_garden_condition"),
		).To(Succeed())
	})

	It("should collect operation succeeded metrics", func() {
		expected := strings.NewReader(`# HELP gardener_operator_garden_operation_succeeded Returns 1 if the last operation state is Succeeded.
# TYPE gardener_operator_garden_operation_succeeded gauge
gardener_operator_garden_operation_succeeded{name="foo"} 1
`)

		Expect(
			testutil.CollectAndCompare(c, expected, "gardener_operator_garden_operation_succeeded"),
		).To(Succeed())
	})

	It("should collect the metric for not succeeded gardens", func() {
		garden.Status.LastOperation.State = gardencorev1beta1.LastOperationStateError
		Expect(k8sClient.Status().Update(ctx, garden)).To(Succeed())

		expected := strings.NewReader(`# HELP gardener_operator_garden_operation_succeeded Returns 1 if the last operation state is Succeeded.
# TYPE gardener_operator_garden_operation_succeeded gauge
gardener_operator_garden_operation_succeeded{name="foo"} 0
`)

		Expect(
			testutil.CollectAndCompare(c, expected, "gardener_operator_garden_operation_succeeded"),
		).To(Succeed())
	})
})
