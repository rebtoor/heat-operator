/*
Copyright 2022.

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

package functional_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/openstack-k8s-operators/lib-common/modules/common/test/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	heatv1 "github.com/openstack-k8s-operators/heat-operator/api/v1beta1"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
)

var _ = Describe("Heat controller", func() {

	var heatName types.NamespacedName
	var heatTransportURLName types.NamespacedName
	var heatConfigMapName types.NamespacedName

	BeforeEach(func() {

		heatName = types.NamespacedName{
			Name:      "heat",
			Namespace: namespace,
		}
		heatTransportURLName = types.NamespacedName{
			Namespace: namespace,
			Name:      heatName.Name + "-heat-transport",
		}
		heatConfigMapName = types.NamespacedName{
			Namespace: namespace,
			Name:      heatName.Name + "-config-data",
		}

		err := os.Setenv("OPERATOR_TEMPLATES", "../../templates")
		Expect(err).NotTo(HaveOccurred())
	})

	When("A Heat instance is created", func() {
		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateHeat(heatName, GetDefaultHeatSpec()))
		})

		It("should have the Spec fields initialized", func() {
			Heat := GetHeat(heatName)
			Expect(Heat.Spec.DatabaseInstance).Should(Equal("openstack"))
			Expect(Heat.Spec.DatabaseUser).Should(Equal("heat"))
			Expect(Heat.Spec.RabbitMqClusterName).Should(Equal("rabbitmq"))
			Expect(Heat.Spec.ServiceUser).Should(Equal("heat"))
			Expect(*(Heat.Spec.HeatAPI.Replicas)).Should(Equal(int32(1)))
			Expect(*(Heat.Spec.HeatCfnAPI.Replicas)).Should(Equal(int32(1)))
			Expect(*(Heat.Spec.HeatEngine.Replicas)).Should(Equal(int32(1)))
		})

		It("should have the Status fields initialized", func() {
			Heat := GetHeat(heatName)
			Expect(Heat.Status.Hash).To(BeEmpty())
			Expect(Heat.Status.DatabaseHostname).To(Equal(""))
			Expect(Heat.Status.TransportURLSecret).To(Equal(""))
			Expect(Heat.Status.HeatAPIReadyCount).To(Equal(int32(0)))
			Expect(Heat.Status.HeatCfnAPIReadyCount).To(Equal(int32(0)))
			Expect(Heat.Status.HeatEngineReadyCount).To(Equal(int32(0)))
		})

		It("should have input not ready and unknown Conditions initialized", func() {
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.InputReadyCondition,
				corev1.ConditionFalse,
			)

			for _, cond := range []condition.Type{
				heatv1.HeatRabbitMqTransportURLReadyCondition,
				condition.ServiceConfigReadyCondition,
				condition.DBReadyCondition,
				condition.DBSyncReadyCondition,
				heatv1.HeatStackDomainReadyCondition,
				heatv1.HeatAPIReadyCondition,
				heatv1.HeatCfnAPIReadyCondition,
				heatv1.HeatEngineReadyCondition,
			} {
				th.ExpectCondition(
					heatName,
					ConditionGetterFunc(HeatConditionGetter),
					cond,
					corev1.ConditionUnknown,
				)
			}
		})

		It("should have a finalizer", func() {
			// the reconciler loop adds the finalizer so we have to wait for
			// it to run
			Eventually(func() []string {
				return GetHeat(heatName).Finalizers
			}, timeout, interval).Should(ContainElement("Heat"))
		})

		It("should not create a config map", func() {
			th.AssertConfigMapDoesNotExist(heatConfigMapName)
		})
	})

	When("The proper secret is provided", func() {
		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateHeat(heatName, GetDefaultHeatSpec()))
			DeferCleanup(
				k8sClient.Delete, ctx, CreateHeatSecret(namespace, SecretName))
		})

		It("should have input ready", func() {
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.InputReadyCondition,
				corev1.ConditionTrue,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				heatv1.HeatRabbitMqTransportURLReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ServiceConfigReadyCondition,
				corev1.ConditionUnknown,
			)
		})

		It("should not create a config map", func() {
			th.AssertConfigMapDoesNotExist(heatConfigMapName)
		})
	})

	When("TransportURL Created", func() {
		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateHeat(heatName, GetDefaultHeatSpec()))
			DeferCleanup(
				k8sClient.Delete, ctx, CreateHeatSecret(namespace, SecretName))
			DeferCleanup(
				k8sClient.Delete, ctx, CreateHeatMessageBusSecret(namespace, HeatMessageBusSecretName))
			th.SimulateTransportURLReady(heatTransportURLName)
		})

		It("should have transporturl ready", func() {
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				heatv1.HeatRabbitMqTransportURLReadyCondition,
				corev1.ConditionTrue,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ServiceConfigReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.DBSyncReadyCondition,
				corev1.ConditionUnknown,
			)
		})

		It("should not create a config map", func() {
			th.AssertConfigMapDoesNotExist(heatConfigMapName)
		})
	})

	When("keystoneAPI instance is available", func() {
		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateHeat(heatName, GetDefaultHeatSpec()))
			DeferCleanup(
				k8sClient.Delete, ctx, CreateHeatSecret(namespace, SecretName))
			DeferCleanup(
				k8sClient.Delete, ctx, CreateHeatMessageBusSecret(namespace, HeatMessageBusSecretName))
			th.SimulateTransportURLReady(heatTransportURLName)
			keystoneAPI := th.CreateKeystoneAPI(namespace)
			DeferCleanup(th.DeleteKeystoneAPI, keystoneAPI)
		})

		It("should have service config ready", func() {
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ServiceConfigReadyCondition,
				corev1.ConditionTrue,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ReadyCondition,
				corev1.ConditionFalse,
			)
		})

		It("should create a ConfigMap for heat.conf with the heat_domain_admin config option set", func() {
			cm := th.GetConfigMap(heatConfigMapName)

			Expect(cm.Data["heat.conf"]).Should(
				ContainSubstring("stack_domain_admin = heat_stack_domain_admin"))
		})
	})

	When("DB is created", func() {
		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateHeat(heatName, GetDefaultHeatSpec()))
			DeferCleanup(
				k8sClient.Delete, ctx, CreateHeatSecret(namespace, SecretName))
			DeferCleanup(
				k8sClient.Delete, ctx, CreateHeatMessageBusSecret(namespace, HeatMessageBusSecretName))
			th.SimulateTransportURLReady(heatTransportURLName)
			keystoneAPI := th.CreateKeystoneAPI(namespace)
			DeferCleanup(th.DeleteKeystoneAPI, keystoneAPI)
			DeferCleanup(
				th.DeleteDBService,
				th.CreateDBService(
					namespace,
					GetHeat(heatName).Spec.DatabaseInstance,
					corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{Port: 3306}},
					},
				),
			)
			th.SimulateMariaDBDatabaseCompleted(heatName)
		})

		It("should have db ready condition", func() {
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.DBReadyCondition,
				corev1.ConditionTrue,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.DBSyncReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				heatv1.HeatStackDomainReadyCondition,
				corev1.ConditionUnknown,
			)
		})
	})

	When("DB sync is completed", func() {
		BeforeEach(func() {
			DeferCleanup(th.DeleteInstance, CreateHeat(heatName, GetDefaultHeatSpec()))
			DeferCleanup(
				k8sClient.Delete, ctx, CreateHeatSecret(namespace, SecretName))
			DeferCleanup(
				k8sClient.Delete, ctx, CreateHeatMessageBusSecret(namespace, HeatMessageBusSecretName))
			th.SimulateTransportURLReady(heatTransportURLName)
			keystoneAPI := th.CreateKeystoneAPI(namespace)
			DeferCleanup(th.DeleteKeystoneAPI, keystoneAPI)
			DeferCleanup(
				th.DeleteDBService,
				th.CreateDBService(
					namespace,
					GetHeat(heatName).Spec.DatabaseInstance,
					corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{Port: 3306}},
					},
				),
			)
			th.SimulateMariaDBDatabaseCompleted(heatName)
			dbSyncJobName := types.NamespacedName{
				Name:      "heat-db-sync",
				Namespace: namespace,
			}
			th.SimulateJobSuccess(dbSyncJobName)
		})

		It("should have db sync ready condition", func() {
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.ReadyCondition,
				corev1.ConditionFalse,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				condition.DBSyncReadyCondition,
				corev1.ConditionTrue,
			)
			th.ExpectCondition(
				heatName,
				ConditionGetterFunc(HeatConditionGetter),
				heatv1.HeatStackDomainReadyCondition,
				corev1.ConditionFalse,
			)
		})
	})
})
