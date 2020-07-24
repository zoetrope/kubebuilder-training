# CRDマニフェストの生成(応用編)

テナントリソースでは利用していない応用的な機能が用意されているので

## Defaulting, Pruning

Defaulting機能を利用するためには、Structural SchemeかつPruningが有効になっている必要があります。

Pruningを有効にするためには
CRDの`spec.preserveUnknownFields: false`にするか、
v1にすればいい。

defaultにはoptionalをつけないと駄目。


`apiextensions.k8s.io/v1beta1`

defaultingやpruning

`apiextensions.k8s.io/v1`

structural

```console
$ make manifests CRD_OPTIONS=crd:crdVersions=v1
```

## Server Side Apply用のマーカー

https://kubernetes.io/docs/reference/using-api/api-concepts/#merge-strategy

Kubernetesには、リソース全体を更新するのではなく、一部のフィールドのみを

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  name: sample
spec:
  namespaces:
  - test1
  - test2
  syncResources:
  - apiVersion: v1
    kind: ConfigMap
    name: test
    namespace: sample1
    mode: remove
  - apiVersion: v1
    kind: ConfigMap
    name: test
    namespace: sample2
    mode: remove
```

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  name: sample
spec:
  syncResources:
  - apiVersion: v1
    kind: ConfigMap
    name: test
    namespace: sample2
    mode: ignore
```

```go
type TenantSpec struct {
	// +listType=map
	// +listMapKey=apiVersion
	// +listMapKey=kind
	// +listMapKey=name
	// +listMapKey=namespace
	SyncResources []SyncResource `json:"syncResources,omitempty"`
}
```

なお、`listMapKey`に指定したフィールドは、`Required`にするかデフォルト値を設定する


## 既存リソースの利用

前節の例では`TenantSpec`の`Admin`フィールドに、Kubernetesが標準で用意している[rbac/v1/Subject](https://pkg.go.dev/k8s.io/api/rbac/v1?tab=doc#Subject)型を利用していました。
このように標準で用意されているリソースをカスタムリソースに組み込むことが可能です。

しかし、標準リソースをカスタムリソースに組み込んだ際に問題が発生するケースがあります。

例えば、[core/v1/Container](https://pkg.go.dev/k8s.io/api/core/v1?tab=doc#Container)型を組み込んだ場合をみてみましょう。

Containerは、以下のようなPortsフィールドを保持しています。

```go
	// +optional
	// +patchMergeKey=containerPort
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=containerPort
	// +listMapKey=protocol
	Ports []ContainerPort `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"containerPort" protobuf:"bytes,6,rep,name=ports"`
```

このPortsフィールドには`+listMapKey=protocol`というマーカーが付与されているため、controller-genを実行すると下記のようなCRDが生成されます。

```yaml
ports:
  description: List of ports to expose from the container.
    Exposing a port here gives the system additional information
    about the network connections a container uses, but
    is primarily informational. Not specifying a port
    here DOES NOT prevent that port from being exposed.
    Any port which is listening on the default "0.0.0.0"
    address inside a container will be accessible from
    the network. Cannot be updated.
  items:
    description: ContainerPort represents a network port
      in a single container.
    properties:
      containerPort:
        description: Number of port to expose on the pod's
          IP address. This must be a valid port number,
          0 < x < 65536.
        format: int32
        type: integer
      hostIP:
        description: What host IP to bind the external
          port to.
        type: string
      hostPort:
        description: Number of port to expose on the host.
          If specified, this must be a valid port number,
          0 < x < 65536. If HostNetwork is specified,
          this must match ContainerPort. Most containers
          do not need this.
        format: int32
        type: integer
      name:
        description: If specified, this must be an IANA_SVC_NAME
          and unique within the pod. Each named port in
          a pod must have a unique name. Name for the
          port that can be referred to by services.
        type: string
      protocol:
        description: Protocol for port. Must be UDP, TCP,
          or SCTP. Defaults to "TCP".
        type: string
    required:
    - containerPort
    type: object
  type: array
  x-kubernetes-list-map-keys:
  - containerPort
  - protocol
  x-kubernetes-list-type: map
```

このCRDをv1.18のKubernetesクラスタに適用すると、下記のようなエラーが発生します。

```
Required value: this property is in x-kubernetes-list-map-keys, so it must have a default or be a required property
```

`x-kubernetes-list-map-keys`に指定されたフィールドは、Requiredであるかデフォルト値を持っている必要があるのですが、`protocol`フィールドはそのどちらでもないためです。
この問題はKubernetes 1.18とcontroller-gen 0.3.0時点では解決されていません。([参考](https://github.com/kubernetes/kubernetes/issues/91395))

現状では以下のように生成されたCRDを書き換えてデフォルト値を設定することで、この問題を回避することができます。

```diff
      protocol:
        description: Protocol for port. Must be UDP, TCP,
          or SCTP. Defaults to "TCP".
+       default: TCP
        type: string
```
