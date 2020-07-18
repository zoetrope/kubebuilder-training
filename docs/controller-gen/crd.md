# CRDの生成

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
