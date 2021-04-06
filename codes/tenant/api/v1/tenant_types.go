package v1

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//! [spec]
// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Namespaces are the names of the namespaces that belong to the tenant
	// +kubebuiler:validation:Required
	// +kubebuiler:validation:MinItems=1
	Namespaces []string `json:"namespaces"`
	// NamespacePrefix is the prefix for the name of namespaces
	// +optional
	NamespacePrefix string `json:"namespacePrefix,omitempty"`
	// Admin is the identity with admin for the tenant
	// +kubebuiler:validation:Required
	Admin rbacv1.Subject `json:"admin"`
}

//! [spec]

//! [status]
// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions is an array of conditions.
	// +optional
	Conditions []TenantCondition `json:"conditions,omitempty"`
}

type TenantCondition struct {
	// Type is the type for the condition
	Type TenantConditionType `json:"type"`
	// Status is the status of the condition
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a one-word CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Message is a human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
	// LastTransitionTime is the time of the last transition.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

// TenantConditionType is the type of Tenant condition.
// +kubebuilder:validation:Enum=Ready
type TenantConditionType string

// Valid values for TenantConditionType
const (
	ConditionReady TenantConditionType = "Ready"
)

//! [status]

//! [tenant]
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ADMIN",type="string",JSONPath=".spec.admin.name"
// +kubebuilder:printcolumn:name="PREFIX",type="string",JSONPath=".spec.namespacePrefix"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"

// Tenant is the Schema for the tenants API
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

//! [tenant]

// +kubebuilder:object:root=true

// TenantList contains a list of Tenant
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}
