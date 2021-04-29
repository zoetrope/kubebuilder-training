package controllers

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	multitenancyv1 "github.com/zoetrope/kubebuilder-training/codes/api/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Tenant controller", func() {

	const (
		tenantName = "sample"
	)

	Context("when creating Tenant resource", func() {
		It("Should create namespaces", func() {
			ctx := context.Background()
			tenant := &multitenancyv1.Tenant{
				TypeMeta: metav1.TypeMeta{
					APIVersion: multitenancyv1.GroupVersion.String(),
					Kind:       "Tenant",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: tenantName,
				},
				Spec: multitenancyv1.TenantSpec{
					Namespaces: []string{
						"test1",
						"test2",
					},
					NamespacePrefix: "",
					Admin: rbacv1.Subject{
						Kind:      "ServiceAccount",
						Name:      "default",
						Namespace: "default",
					},
				},
			}
			err := k8sClient.Create(ctx, tenant)
			Expect(err).Should(Succeed())

			createdTenant := &multitenancyv1.Tenant{}
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: tenantName}, createdTenant)
				if err != nil {
					return err
				}
				cond := meta.FindStatusCondition(createdTenant.Status.Conditions, multitenancyv1.ConditionReady)
				if cond == nil {
					return errors.New("condition not found")
				}
				if cond.Status != metav1.ConditionTrue {
					return errors.New(cond.Reason)
				}
				return nil
			}).Should(Succeed())

			nsList := &corev1.NamespaceList{}
			err = k8sClient.List(ctx, nsList)
			Expect(err).Should(Succeed())

			var namespaces []string
			for _, ns := range nsList.Items {
				namespaces = append(namespaces, ns.Name)
			}
			Expect(namespaces).Should(ContainElement("test1"))
			Expect(namespaces).Should(ContainElement("test2"))
		})
	})
})
