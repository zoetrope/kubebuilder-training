# CRDマニフェストの生成

コントローラでカスタムリソースを扱うためには、そのリソースのCRD(Custom Resource Definition)を定義する必要があります。
下記の例の様にCRDは長くなりがちで、人間が記述するには少々大変です。

- [CRDの例](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/tenant/config/crd/bases/multitenancy.example.com_tenants.yaml)

そこでKubebuilderではcontroller-genというツールを利用して、Goで記述したstructからCRDを生成する方式を採用しています。

`kubebuilder create api`コマンドで生成された[api/v1/tenant_types.go](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/tenant/api/v1/tenant_types.go)を見てみると、`TenantSpec`, `TenantStatus`, `Tenant`, `TenantList`という構造体が定義されており、たくさんの`// +kubebuilder:`から始まるマーカーコメントが付与されています。
controller-genは、これらの構造体とマーカーを頼りにCRDの生成をおこないます。

`Tenant`がカスタムリソースの本体となる構造体です。`TenantList`は`Tenant`のリストを表す構造体です。これら2つの構造体は基本的に変更することはありません。
`TenantSpec`と`TenantStatus`は`Tenant`構造体を構成する要素です。この2つの構造体を書き換えてカスタムリソースを定義していきます。

一般的にカスタムリソースの`Spec`はユーザーが記述するもので、システムのあるべき状態をユーザーからコントローラに伝えるために利用されます。
一方の`Status`は、コントローラが処理した結果をユーザーや他のシステムに伝えるために利用されます。

## TenantSpec

さっそく`TenantSpec`を定義していきましょう。

[作成するカスタムコントローラ](../introduction/sample.md)において、テナントコントローラが扱うカスタムリソースとして、下記のようなマニフェストを検討しました。

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  name: sample
spec:
  namespaces:
    - test1
    - test2
  namespacePrefix: sample-
  admin:
    kind: User
    name: test
    namespace: default
    apiGroup: rbac.authorization.k8s.io
```

上記のマニフェストを取り扱うための構造体を用意しましょう。

[import:"spec"](../../codes/tenant/api/v1/tenant_types.go)

まず下記の3つのフィールドを定義します。

- `Namespaces`: テナントに属するnamespaceの一覧を指定
- `NamespacePrefix`: namespace名のプリフィックスを指定
- `Admin`: テナントの管理ユーザーを指定

各フィールドの上に`// +kubebuilder`という文字列から始まるマーカーと呼ばれるコメントが記述されています。
これらのマーカーによって、生成されるCRDの内容を制御することができます。

付与できるマーカーは`controller-gen crd -w`コマンドで確認することができます。

### Required/Optional

`Namespaces`と`Admin`フィールドには`+kubebuiler:validation:Required`マーカーが付与されています。
これはこのフィールドが必須項目であることを示しており、ユーザーがマニフェストを記述する際にこの項目を省略することができません。
一方の`NamespacePrefix`には`+optional`が付与されており、この項目が省略可能であることを示しています。

マーカーを指定しなかった場合はデフォルトでRequiredなフィールドになります。
ファイル内に下記のマーカーを配置すると、デフォルトの挙動をOptionalに変更することができます。

```
// +kubebuilder:validation:Optional
```

`+optional`マーカーを付与しなくても、フィールドの後ろのJSONタグに`omitempty`を付与した場合は、自動的にOptionalなフィールドとなります。

```go
type SampleSpec struct {
	Value string `json:"value,omitempty"`
}
```

Optionalなフィールドは、以下のようにフィールドの型をポインタにすることができます。
これによりマニフェストで値を指定しなかった場合の挙動が異なります。
ポインタ型にした場合はnullが入り、実体にした場合はその型の初期値(intの場合は0)が入ります。

```go
type SampleSpec struct {
	// +optional
	Value1 int  `json:"value1"`
	// +optional
	Value2 *int `json:"value2"`
}
```

### Validation

`Namespaces`フィールドには`// +kubebuiler:validation:MinItems=1`というマーカーが付与されています。
これは最低1つ以上のnamespaceを記述しないと、カスタムリソースを作成するときにバリデーションエラーとなることを示しています。

`MinItems`以外にも下記のようなバリデーションが用意されています。
詳しくは`controller-gen crd -w`コマンドで確認してください。

- リストの最小要素数、最大要素数
- 文字列の最小長、最大長
- 数値の最小値、最大値
- 正規表現にマッチするかどうか
- リスト内の要素がユニークかどうか


## TenantStatus

次にテナントリソースの状態を表現するために`TenantStatus`に`Conditions`フィールドを追加します。
このようなCondition型は様々なリソースで利用されている頻出パターンなので覚えておくとよいでしょう。

[import:"status"](../../codes/tenant/api/v1/tenant_types.go)

汎用的な Condition 型が Kubernetes 1.19 で追加されるので、こちらを使っていくのも良いです。

- https://pkg.go.dev/k8s.io/apimachinery@v0.19.0-rc.4/pkg/apis/meta/v1?tab=doc#Condition

`TenantConditionType`には`// +kubebuilder:validation:Enum=Ready`というマーカーが付与されています。
これにより`TenantConditionType`は列挙型となり、マーカーで列挙した値(ここでは"Ready")以外の値を指定できなくなります。

このStatusフィールドにより、ユーザーや他のシステム(モニタリングシステムなど)がテナントリソースの状態を確認することができるようになります。

テナントの作成に成功した場合には下記のようなStatusになります。

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  name: sample
spec: # 省略
status:
  conditions:
  - type: Ready
    status: True
    lastTransitionTime: "2020-07-18T09:01:02Z"
```

テナントの作成に失敗すると下記のようなStatusになります。

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  name: sample
spec: # 省略
status:
  conditions:
  - type: Ready
    status: False
    reason: Failed
    message: "failed to create 'test1' namespace"
    lastTransitionTime: "2020-07-18T10:15:34"
```

## Tenant

続いて`Tenant`構造体のマーカーを見てみましょう。

[import:"tenant"](../../codes/tenant/api/v1/tenant_types.go)

- `+kubebuilder:object:root=true`: `Tenant`構造体がカスタムリソースのrootオブジェクトであることを表すマーカーです。
- `+kubebuilder:resource:scope=Cluster`: `Tenant`がcluster-scopeのカスタムリソースであることを表すマーカーです。namespaced-scopeの場合は"scope=Namespaced"とするか、このマーカーを省略します。

上記に加えて`+kubebuilder:subresource`と`+kubebuilder:printcolumn`を付与します。

### subresource

`+kubebuilder:subresource:status`というマーカーを追加すると、`status`フィールドがサブリソースとして扱われるようになります。

サブリソースを有効にすると`status`が独自のエンドポイントを持つようになります。
これによりTenantリソース全体を取得・更新しなくても、`status`のみを取得したり更新することが可能になります。
ただし、あくまでもメインのリソースに属するリソースなので、個別に作成や削除することはできません。

ユーザーが`spec`フィールドを記述し、コントローラが`status`フィールドを記述するという役割分担を明確にすることができるので、基本的には`status`はサブリソースにしておくのがよいでしょう。

なお、CRDでは任意のフィールドをサブリソースにすることはできず、`status`と`scale`の2つのフィールドのみに対応しています。

### printcolumn

`+kubebuilder:printcolumn`マーカーを付与すると、kubectlでカスタムリソースを取得したときに表示する項目を指定することができます。

表示対象のフィールドはJSONPathで指定することが可能です。
これにより`JSONPath=".status.conditions[?(@.type=='Ready')].status"`と記述すると、`status.conditions`の中の`type`が"Ready"な要素の`status`の値を表示することができます。

kubectlでTenantリソースを取得すると、下記のようにPREFIXやREADYの値が表示されていることが確認できます。

```
$ kubectl get tenant
NAME            ADMIN     PREFIX           READY
tenant-sample   default   tenant-sample-   True
```
