package v1

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var tenantlog = logf.Log.WithName("tenant-resource")

func (r *Tenant) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//! [webhook-defaulter]
// +kubebuilder:webhook:path=/mutate-multitenancy-example-com-v1-tenant,mutating=true,failurePolicy=fail,sideEffects=None,groups=multitenancy.example.com,resources=tenants,verbs=create,versions=v1,name=mtenant.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Tenant{}

//! [webhook-defaulter]

//! [default]
// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Tenant) Default() {
	tenantlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Spec.NamespacePrefix == "" {
		r.Spec.NamespacePrefix = r.Name + "-"
	}
}

//! [default]

//! [webhook-validator]
// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-multitenancy-example-com-v1-tenant,mutating=false,failurePolicy=fail,sideEffects=None,groups=multitenancy.example.com,resources=tenants,verbs=update,versions=v1,name=vtenant.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Tenant{}

//! [webhook-validator]

//! [validate]
// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Tenant) ValidateCreate() error {
	tenantlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Tenant) ValidateUpdate(old runtime.Object) error {
	tenantlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	oldTenant := old.(*Tenant)
	if r.Spec.NamespacePrefix != oldTenant.Spec.NamespacePrefix {
		return errors.New("spec.namespacePrefix field should not be changed")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Tenant) ValidateDelete() error {
	tenantlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

//! [validate]
