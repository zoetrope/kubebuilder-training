package v1

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func mutateTest(before string, after string) {
	ctx := context.Background()

	y, err := os.ReadFile(before)
	Expect(err).NotTo(HaveOccurred())
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(y), 4096)
	beforeView := &MarkdownView{}
	err = d.Decode(beforeView)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(ctx, beforeView)
	Expect(err).NotTo(HaveOccurred())

	ret := &MarkdownView{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: beforeView.GetName(), Namespace: beforeView.GetNamespace()}, ret)
	Expect(err).NotTo(HaveOccurred())

	y, err = os.ReadFile(after)
	Expect(err).NotTo(HaveOccurred())
	d = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(y), 4096)
	afterView := &MarkdownView{}
	err = d.Decode(afterView)
	Expect(err).NotTo(HaveOccurred())

	Expect(ret.Spec).Should(Equal(afterView.Spec))
}

func validateTest(file string, valid bool) {
	ctx := context.Background()
	y, err := os.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(y), 4096)
	view := &MarkdownView{}
	err = d.Decode(view)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(ctx, view)

	if valid {
		Expect(err).NotTo(HaveOccurred(), "MarkdownView: %v", view)
	} else {
		Expect(err).To(HaveOccurred(), "MarkdownView: %v", view)
		statusErr := &apierrors.StatusError{}
		Expect(errors.As(err, &statusErr)).To(BeTrue())
		expected := view.Annotations["message"]
		Expect(statusErr.ErrStatus.Message).To(ContainSubstring(expected))
	}
}

var _ = Describe("MarkdownView Webhook", func() {
	Context("mutating", func() {
		It("should mutate a MarkdownView", func() {
			mutateTest(filepath.Join("testdata", "mutating", "before.yaml"), filepath.Join("testdata", "mutating", "after.yaml"))
		})
	})
	Context("validating", func() {
		It("should create a valid MarkdownView", func() {
			validateTest(filepath.Join("testdata", "validating", "valid.yaml"), true)
		})
		It("should not create invalid MarkdownViews", func() {
			validateTest(filepath.Join("testdata", "validating", "empty-markdowns.yaml"), false)
			validateTest(filepath.Join("testdata", "validating", "invalid-replicas.yaml"), false)
			validateTest(filepath.Join("testdata", "validating", "without-summary.yaml"), false)
		})
	})
})
