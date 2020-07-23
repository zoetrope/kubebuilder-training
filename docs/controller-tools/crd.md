# CRDマニフェストの生成

コントローラでカスタムリソースを扱うためには、そのリソースのCRD(Custom Resource Definition)を定義する必要があります。このCRDはOpenAPI v3.0の形式で書く必要があるのですが、人間が記述するには少々大変です。

- [CRDの例](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/tenant/config/crd/bases/multitenancy.example.com_tenants.yaml)

そこでKubebuilderでは、controller-genというツールを利用して、Goで記述したstructからCRDを生成することが可能になっています。

## Spec

[import:"spec"](../../codes/tenant/api/v1/tenant_types.go)

テナントに属するnamespaceの一覧を
namespace名のプリフィックスを指定するための`NamespacePrefix`
テナントの管理ユーザーを指定するために`Admin`フィールド

また、各フィールドの上に`// +kubebuilder`という文字列から始まるマーカーと呼ばれるコメントが記述されています。
これらのマーカーによって、生成されるCRDの内容を制御することができます。

付与できるマーカーは`controller-gen crd -w`コマンドで確認することができます。

Required, Optional
各種バリデーション
Enum型
デフォルト値の設定


## Status

`Phase`フィールドを用意して現在の状態のみを格納するのではなく、`Conditions`フィールドで各状態を
判断できるようにしておくことが推奨されています。

[API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties)

Tenantリソースでは状態遷移を扱う必要はないのですが、ここではConditionsフィールドを定義してみましょう。

[import:"status"](../../codes/tenant/api/v1/tenant_types.go)

## Tenant

[import:"tenant"](../../codes/tenant/api/v1/tenant_types.go)

`+kubebuilder:object:root=true`: `Tenant`というstructがAPIのrootオブジェクトであることを表すマーカーです。
`+kubebuilder:resource:scope=Cluster`: `Tenant`がcluster-scopeのカスタムリソースであることを表すマーカーです。

### subresource

`+kubebuilder:subresource:status`というマーカーを追加すると、`status`フィールドがサブリソースとして扱われるようになります。

サブリソースを有効にすると`status`が独自のエンドポイントを持つようになります。
これによりTenantリソース全体を取得・更新しなくても、`status`のみを取得したり更新することが可能になります。
ただし、あくまでもメインのリソースに属するリソースなので、個別に作成や削除することはできません。

サブリソース化しておかないと、クライアントでの編集
基本的にはstatusはサブリソースにしておくのがよいでしょう。

なお、CRDでは任意のサブリソースをもたせることはできず、`status`と`scale`の2つのフィールドのみに対応しています。

### printcolumn

表示対象のフィールドはJSONPathで指定することが可能です。これにより
例えば、`JSONPath=".status.conditions[?(@.type=='Ready')].status"`と記述すると、

kubectlでTenantリソースを取得すると、下記のようにPREFIXやREADYの値が表示されていることが確認できます。

```
$ kubectl get tenant
NAME            ADMIN     PREFIX           READY
tenant-sample   default   tenant-sample-   True
```

## Defaulting, Pruning

Defaulting機能を利用するためには、Structural SchemeかつPruningが有効になっている必要があります。

Pruningを有効にするためには
CRDの`spec.preserveUnknownFields: false`にするか、
v1にすればいい。


`apiextensions.k8s.io/v1beta1`

defaultingやpruning

`apiextensions.k8s.io/v1`

structural

```console
$ make manifests CRD_OPTIONS=crd:crdVersions=v1
```

## 既存リソースの利用

上記の例では`TenantSpec`の`Admin`フィールドに、Kubernetesが標準で用意している[rbac/v1/Subject](https://pkg.go.dev/k8s.io/api/rbac/v1?tab=doc#Subject)型を利用していました。
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
