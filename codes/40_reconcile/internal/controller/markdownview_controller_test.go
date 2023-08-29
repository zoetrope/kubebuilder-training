package controller

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	viewv1 "github.com/zoetrope/markdown-view/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MarkdownView controller", func() {

	ctx := context.Background()
	var stopFunc func()

	BeforeEach(func() {
		err := k8sClient.DeleteAllOf(ctx, &viewv1.MarkdownView{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		svcs := &corev1.ServiceList{}
		err = k8sClient.List(ctx, svcs, client.InNamespace("test"))
		Expect(err).NotTo(HaveOccurred())
		for _, svc := range svcs.Items {
			err := k8sClient.Delete(ctx, &svc)
			Expect(err).NotTo(HaveOccurred())
		}
		time.Sleep(100 * time.Millisecond)

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme,
		})
		Expect(err).ToNot(HaveOccurred())

		reconciler := MarkdownViewReconciler{
			Client: k8sClient,
			Scheme: scheme,
		}
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel := context.WithCancel(ctx)
		stopFunc = cancel
		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		stopFunc()
		time.Sleep(100 * time.Millisecond)
	})

	It("should create ConfigMap", func() {
		mdView := newMarkdownView()
		err := k8sClient.Create(ctx, mdView)
		Expect(err).NotTo(HaveOccurred())

		cm := corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "markdowns-sample"}, &cm)
		}).Should(Succeed())
		Expect(cm.Data).Should(HaveKey("SUMMARY.md"))
		Expect(cm.Data).Should(HaveKey("page1.md"))
	})

	It("should create Deployment", func() {
		mdView := newMarkdownView()
		err := k8sClient.Create(ctx, mdView)
		Expect(err).NotTo(HaveOccurred())

		dep := appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "viewer-sample"}, &dep)
		}).Should(Succeed())
		Expect(dep.Spec.Replicas).Should(Equal(pointer.Int32Ptr(3)))
		Expect(dep.Spec.Template.Spec.Containers[0].Image).Should(Equal("peaceiris/mdbook:0.4.10"))
	})

	It("should create Service", func() {
		mdView := newMarkdownView()
		err := k8sClient.Create(ctx, mdView)
		Expect(err).NotTo(HaveOccurred())

		svc := corev1.Service{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "viewer-sample"}, &svc)
		}).Should(Succeed())
		Expect(svc.Spec.Ports[0].Port).Should(Equal(int32(80)))
		Expect(svc.Spec.Ports[0].TargetPort).Should(Equal(intstr.FromInt(3000)))
	})

	It("should update status", func() {
		mdView := newMarkdownView()
		err := k8sClient.Create(ctx, mdView)
		Expect(err).NotTo(HaveOccurred())

		updated := viewv1.MarkdownView{}
		Eventually(func() error {
			err := k8sClient.Get(ctx, client.ObjectKey{Namespace: "test", Name: "sample"}, &updated)
			if err != nil {
				return err
			}
			if updated.Status == "" {
				return errors.New("status should be updated")
			}
			return nil
		}).Should(Succeed())
	})
})

func newMarkdownView() *viewv1.MarkdownView {
	return &viewv1.MarkdownView{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "test",
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
}
