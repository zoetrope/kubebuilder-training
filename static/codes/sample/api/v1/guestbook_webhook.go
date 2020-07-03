/*


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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var guestbooklog = logf.Log.WithName("guestbook-resource")

func (r *Guestbook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-webapp-example-com-v1-guestbook,mutating=true,failurePolicy=fail,groups=webapp.example.com,resources=guestbooks,verbs=create;update,versions=v1,name=mguestbook.kb.io

var _ webhook.Defaulter = &Guestbook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Guestbook) Default() {
	guestbooklog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-webapp-example-com-v1-guestbook,mutating=false,failurePolicy=fail,groups=webapp.example.com,resources=guestbooks,versions=v1,name=vguestbook.kb.io

var _ webhook.Validator = &Guestbook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Guestbook) ValidateCreate() error {
	guestbooklog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Guestbook) ValidateUpdate(old runtime.Object) error {
	guestbooklog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Guestbook) ValidateDelete() error {
	guestbooklog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
