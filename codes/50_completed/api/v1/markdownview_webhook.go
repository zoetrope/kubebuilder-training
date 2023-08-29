/*
Copyright 2023.

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

//! [head]

package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var markdownviewlog = logf.Log.WithName("markdownview-resource")

func (r *MarkdownView) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//! [head]

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//! [webhook-defaulter]
//+kubebuilder:webhook:path=/mutate-view-zoetrope-github-io-v1-markdownview,mutating=true,failurePolicy=fail,sideEffects=None,groups=view.zoetrope.github.io,resources=markdownviews,verbs=create;update,versions=v1,name=mmarkdownview.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &MarkdownView{}

//! [webhook-defaulter]

//! [default]

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MarkdownView) Default() {
	markdownviewlog.Info("default", "name", r.Name)

	if len(r.Spec.ViewerImage) == 0 {
		r.Spec.ViewerImage = "peaceiris/mdbook:latest"
	}
}

//! [default]

//! [webhook-validator]
// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-view-zoetrope-github-io-v1-markdownview,mutating=false,failurePolicy=fail,sideEffects=None,groups=view.zoetrope.github.io,resources=markdownviews,verbs=create;update,versions=v1,name=vmarkdownview.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MarkdownView{}

//! [webhook-validator]

//! [validate]

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MarkdownView) ValidateCreate() (admission.Warnings, error) {
	markdownviewlog.Info("validate create", "name", r.Name)

	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MarkdownView) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	markdownviewlog.Info("validate update", "name", r.Name)

	return r.validate()
}

func (r *MarkdownView) validate() (admission.Warnings, error) {
	var errs field.ErrorList

	if r.Spec.Replicas < 1 || r.Spec.Replicas > 5 {
		errs = append(errs, field.Invalid(field.NewPath("spec", "replicas"), r.Spec.Replicas, "replicas must be in the range of 1 to 5."))
	}

	hasSummary := false
	for name := range r.Spec.Markdowns {
		if name == "SUMMARY.md" {
			hasSummary = true
		}
	}
	if !hasSummary {
		errs = append(errs, field.Required(field.NewPath("spec", "markdowns"), "markdowns must have SUMMARY.md."))
	}

	if len(errs) > 0 {
		err := apierrors.NewInvalid(schema.GroupKind{Group: GroupVersion.Group, Kind: "MarkdownView"}, r.Name, errs)
		markdownviewlog.Error(err, "validation error", "name", r.Name)
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MarkdownView) ValidateDelete() (admission.Warnings, error) {
	markdownviewlog.Info("validate delete", "name", r.Name)

	return nil, nil
}

//! [validate]
