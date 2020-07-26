package v1

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Tenant Webhook", func() {
	It("should create a valid tenant", func() {
		ctx := context.Background()
		tenant := &Tenant{
			TypeMeta: metav1.TypeMeta{
				APIVersion: GroupVersion.String(),
				Kind:       "Tenant",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: TenantSpec{
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

		mutatedTenant := &Tenant{}
		err = k8sClient.Get(ctx, client.ObjectKey{Name: "test"}, mutatedTenant)
		Expect(err).Should(Succeed())
		Expect(mutatedTenant.Spec.NamespacePrefix).Should(Equal("test-"))
	})
})
