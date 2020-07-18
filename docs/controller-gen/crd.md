# CRDの生成

[api/v1/tenant_types.go](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/tenant/api/v1/tenant_types.go)

## SubResource

```go
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// Tenant is the Schema for the tenants API
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}
```


`+kubebuilder:object:root=true`: `Tenant`というstructがAPIのrootオブジェクトであることを表すマーカーです。
`+kubebuilder:resource:scope=Cluster`: `Tenant`がcluster-scopeのカスタムリソースであることを表すマーカーです。



## Spec

```go
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
```

```go
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="ADMIN",type="string",JSONPath=".spec.admin.name"
// +kubebuilder:printcolumn:name="PREFIX",type="string",JSONPath=".spec.namespacePrefix"

// Tenant is the Schema for the tenants API
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}
```

## Status

`Phase`フィールドを用意して現在の状態のみを格納するのではなく、`Conditions`フィールドで各状態を
判断できるようにしておくことが推奨されています。

https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

Tenantリソースでは状態遷移を扱う必要はないのですが、ここではConditionsフィールドを定義してみましょう。

```go
// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions is an array of conditions.
	// +optional
	Conditions []TenantCondition `json:"conditions,omitempty"`
}


type TenantCondition struct {
	// Type is the type fo the condition
	Type TenantConditionType `json:"type"`
	// Status is the status of the condition
	Status corev1.ConditionStatus `json:"status"`
	// Reason is a one-word CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Message is a human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
	// Message is a human-readable message indicating details about last transition.
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
}

// TenantConditionType is the type of Tenant condition.
// +kubebuilder:validation:Enum=Ready
type TenantConditionType string

// Valid values for TenantConditionType
const (
	ConditionReady TenantConditionType = "Ready"
)
```

```go
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
```

ここに`+kubebuilder:subresource:status`というマーカーを追加すると、
`status`フィールドがサブリソースとして扱われるようになります。

サブリソースを有効にすると`status`が独自のエンドポイントを持つようになります。
これによりTenantリソース全体を取得・更新しなくても、`status`のみを取得したり更新することが可能になります。
ただし、あくまでもメインのリソースに属するリソースなので、個別に作成や削除することはできません。

なお、CRDでは任意のサブリソースをもたせることはできず、`status`と`scale`の2つのフィールドのみに対応しています。

## Defaulting, Pruning

`apiextensions.k8s.io/v1beta1`

defaultingやpruning

`apiextensions.k8s.io/v1`

structural

```console
$ make manifests CRD_OPTIONS=crd:crdVersions=v1
```

