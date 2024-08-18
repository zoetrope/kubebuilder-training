/*
Copyright 2024.

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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viewv1 "github.com/zoetrope/markdown-view/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("MarkdownView Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "sample"
		const testNamespace = "test"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: testNamespace,
		}

		BeforeEach(func() {
			By("creating the namespace for the test")
			ns := &corev1.Namespace{}
			ns.Name = testNamespace
			err := k8sClient.Create(context.Background(), ns)
			Expect(err).NotTo(HaveOccurred())

			By("creating the custom resource for the Kind MarkdownView")
			markdownview := &viewv1.MarkdownView{}
			err = k8sClient.Get(ctx, typeNamespacedName, markdownview)
			if err != nil && apierrors.IsNotFound(err) {
				resource := &viewv1.MarkdownView{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: testNamespace,
					},
					Spec: viewv1.MarkdownViewSpec{
						Markdowns: map[string]string{
							"SUMMARY.md": `summary`,
							"page1.md":   `page1`,
						},
						Replicas:    3,
						ViewerImage: "peaceiris/mdbook:0.4.10",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &viewv1.MarkdownView{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance MarkdownView")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			err = k8sClient.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace(testNamespace))
			Expect(err).NotTo(HaveOccurred())
			err = k8sClient.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace(testNamespace))
			Expect(err).NotTo(HaveOccurred())
			svcs := &corev1.ServiceList{}
			err = k8sClient.List(ctx, svcs, client.InNamespace(testNamespace))
			Expect(err).NotTo(HaveOccurred())
			for _, svc := range svcs.Items {
				err := k8sClient.Delete(ctx, &svc)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &MarkdownViewReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Making sure the ConfigMap created successfully")
			cm := corev1.ConfigMap{}
			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "markdowns-sample"}, &cm)
			Expect(err).NotTo(HaveOccurred())
			Expect(cm.Data).Should(HaveKey("SUMMARY.md"))
			Expect(cm.Data).Should(HaveKey("page1.md"))

			By("Making sure the Deployment created successfully")
			dep := appsv1.Deployment{}
			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "viewer-sample"}, &dep)
			Expect(err).NotTo(HaveOccurred())
			Expect(dep.Spec.Replicas).Should(Equal(ptr.To[int32](3)))
			Expect(dep.Spec.Template.Spec.Containers[0].Image).Should(Equal("peaceiris/mdbook:0.4.10"))

			By("Making sure the Service created successfully")
			svc := corev1.Service{}
			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "viewer-sample"}, &svc)
			Expect(err).NotTo(HaveOccurred())
			Expect(svc.Spec.Ports[0].Port).Should(Equal(int32(80)))
			Expect(svc.Spec.Ports[0].TargetPort).Should(Equal(intstr.FromInt32(3000)))

			By("Making sure the Status updated successfully")
			updated := viewv1.MarkdownView{}
			err = k8sClient.Get(ctx, typeNamespacedName, &updated)
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Status.Conditions).ShouldNot(BeEmpty(), "status should be updated")
		})
	})
})
