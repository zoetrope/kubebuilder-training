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

func testMutating(input string, output string) {
	ctx := context.Background()

	y, err := os.ReadFile(input)
	Expect(err).NotTo(HaveOccurred())
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(y), 4096)
	inputView := &MarkdownView{}
	err = d.Decode(inputView)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(ctx, inputView)
	Expect(err).NotTo(HaveOccurred())

	ret := &MarkdownView{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: inputView.GetName(), Namespace: inputView.GetNamespace()}, ret)
	Expect(err).NotTo(HaveOccurred())

	y, err = os.ReadFile(output)
	Expect(err).NotTo(HaveOccurred())
	d = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(y), 4096)
	outputView := &MarkdownView{}
	err = d.Decode(outputView)
	Expect(err).NotTo(HaveOccurred())

	Expect(ret.Spec).Should(Equal(outputView.Spec))
}

func testValidating(file string, valid bool) {
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

	Context("When creating MarkdownView under Defaulting Webhook", func() {
		It("Should fill in the default value if a required field is empty", func() {
			testMutating(filepath.Join("testdata", "mutating", "input.yaml"), filepath.Join("testdata", "mutating", "output.yaml"))
		})
	})

	Context("When creating MarkdownView under Validating Webhook", func() {
		It("Should deny if a required field is empty or invalid", func() {
			testValidating(filepath.Join("testdata", "validating", "empty-markdowns.yaml"), false)
			testValidating(filepath.Join("testdata", "validating", "invalid-replicas.yaml"), false)
			testValidating(filepath.Join("testdata", "validating", "without-summary.yaml"), false)
		})

		It("Should admit if all required fields are valid", func() {
			testValidating(filepath.Join("testdata", "validating", "valid.yaml"), true)
		})
	})

})
