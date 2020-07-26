# クライアントの使い方

controller-runtimeでは、Kubernetes APIにアクセスするためのクライアントとして[client.Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client?tab=doc#Client)を提供しています。

このクライアントは標準リソースとカスタムリソースを同じように扱うことができ、

## クライアントの作成

クライアントを作成するためにはまず[Scheme](https://pkg.go.dev/k8s.io/apimachinery/pkg/runtime?tab=doc#Scheme)を用意する必要があります。

SchemeはGoのstructとGroupVersionKindを相互に変換したり、異なるバージョン間でのSchemeの変換をおこなうための機能です。

[import:"init"](../../codes/tenant/main.go)

最初に`runtime.NewScheme()`で新しい`scheme`を作成します。
`clientgoscheme.AddToScheme`では、PodやServiceなどKubernetesの標準リソースの型をschemeに追加しています。
`multitenancyv1.AddToScheme`では、Tenantカスタムリソースの型をschemeに追加しています。

このSchemeを利用することで、標準リソースとTenantリソースを扱うことができるクライアントを作成できます。

続いてこのSchemeとConfigを利用してManagerを作成しClientを取得します。

```
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme: scheme,
})
client := mgr.GetClient()
```
[GetConfigOrDie](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config?tab=doc#GetConfigOrDie)でクライアントの設定を取得しています。

この関数はコマンドラインオプションの`--kubeconfig`や、環境変数`KUBECONFIG`で指定された設定ファイルを利用するか、またはKubernetesクラスタ上でPodとして動いているのであれば、Podが持つサービスアカウントの認証情報を利用します。
通常コントローラはKubernetesクラスタ上で動いているので、クラスタから割り当てられた設定が利用されます。

Managerから`GetClient()`でクライアントを取得することができます。
ただし、Managerの`Start()`を呼び出す前にClientを利用することはできないので注意しましょう。

## Get/List

### Getの使い方

[import:"get",unindent="true"](../../codes/tenant/controllers/tenant_controller.go)

### キャッシュ


### キャッシュの利用を避ける

なお後述するように、このクライアントは`Get()`や`List()`でリソースを取得すると、同一namespaceの同じKindのリソースをすべて取得してインメモリにキャッシュします。
このようなキャッシュの仕組みが必要ない場合は、`GetAPIReader()`でキャッシュを利用しないクライアントを取得することもできます。
基本的には`GetClient()`で取得するクライアントを利用すれば問題ありません。

### Listの使い方

LabelSelectorやNamespaceでフィルタリングすることができます。
Namespaceを指定しなかった場合は、全Namespaceのリソースを取得します。

[import:"list"](../../codes/misc/main.go)

`Limit`と`Continue`を利用することで、ページネーションをおこなうことも可能です。
下記の例では1回のAPI呼び出しで3件ずつリソースを取得して表示しています。

[import:"pagination"](../../codes/misc/main.go)

`.ListMeta.Continue`にトークンが入っているを利用して、続きのリソースを取得することができます。
トークンが空になるとすべてのリソースを取得したということになります。

### インデックス

index field: リソースごとに一意になっていればよい。 実態のフィールドの構成と一致していなくても良い。
informerはgvkごとに作られる。namespaceは自動的にキーに付与されるので、わざわざつけなくてもよい。
戻り値がスライスになっている、複数の値でインデクシングすることも可能。


リソース一覧を取得する際に、条件でフィルタリングしたいことがあるかと思います。
ループで回してもいいのですが、

インメモリキャッシュにインデックスを張ることができます。
インデックスを利用するためには事前に`GetFieldIndexer().IndexField()`を利用して、TenantリソースのConditionReadyの値に応じてインデックスを作成しておきます。

[import:"indexer"](../../codes/tenant/controllers/tenant_controller.go)
[import:"index-field",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

上記のようなインデックスを作成しておくと、`List()`を呼び出す際に特定のフィールドが指定した値と一致するリソースだけを取得することができます。
例えば以下の例であれば、ConditionReadyが"True"のTenantリソース一覧を取得することが可能です。

[import:"matching-fields",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

フィールド名には、どのフィールドを利用してインデックスを張っているのかを示す文字列を指定します。
実際にインデックスに利用しているフィールドのパスと一致していなくても問題はないのですが、なるべく一致させたほうが可読性がよくなるのでおすすめです。
なおinformerはGVKごとに作成されるので、異なるタイプのリソース間でフィールド名が同じになっても問題ありません。
またnamespaceスコープのリソースの場合は、自動的にフィールド名にnamespace名が付与されるので、明示的にフィールド名にnamespaceを含める必要はありません。

## Create/Update

[import:"namespace,create",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)


## CreateOrUpdate

Createはリソースがすでに存在していた場合には失敗
Updateはリソースが存在しない場合には失敗

CreateOrUpdateを利用すると、リソースが存在しなければ作成し、存在すれば更新してくれます。

[import:"create-or-update",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

## Patch

SSA

## Status.Update/Patch

Statusをサブリソース化している場合、これまで紹介した`Update`や`Patch`を利用してもステータスを更新することができません。
Status更新用のクライアントが用意されているのでそれを使いましょう。



逆にStatusをサブリソース化していない場合、これらの機能は利用できません。通常のUpdate/Patchを利用しましょう。

## Delete/DeleteOfAll

## ディスカバリーベースのクライアント

client-goを利用した場合、Kubernetes標準の

k8s.io/client-go/dynamic
k8s.io/apimachinery/pkg/apis/meta/v1/unstructured

`/apis/{group}/{version}/namespaces/{namespace}/{resource}/{name}`
`/apis/{group}/{version}/{resource}/{name}`

`/apis/multitenancy.example.com/v1/tenants/sample`

Clientでは、受け取るための構造体を渡すだけでAPIを呼び分けてくれている。
どのようにURLを解決できるのでしょう。

まずはリフレクションにより`Tenant`構造体から"Tenant"という文字列を取得します。これがKindになります。
さらに[api/v1/groupversion_info.go](../../codes/tenant/api/v1/groupversion_info.go)に埋め込まれた情報をもとにGroupとVersionを取得します。
これでGVKが取得できました。これらはSchemeの仕事です。

次にREST APIを叩くためにはリソース名やnamespace-scopedかどうかを解決する必要があります。
この情報はCRDに記述されているため、APIサーバーに問い合わせる必要があります。

discovery APIというものがあるわけではなく、api以下を何回も叩いてる

```
$ kubectl api-resources --api-group="multitenancy.example.com"
NAME      SHORTNAMES   APIGROUP                   NAMESPACED   KIND
tenants                multitenancy.example.com   false        Tenant
```

Clientはこのような仕組みによって、標準リソースとカスタムリソースを同じように扱うことができ、型安全で簡単に利用できるクライアントを実現しています。
