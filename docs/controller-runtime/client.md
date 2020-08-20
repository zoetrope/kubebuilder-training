# クライアントの使い方

controller-runtimeでは、Kubernetes APIにアクセスするためのクライアントとして[client.Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client?tab=doc#Client)を提供しています。

このクライアントは標準リソースとカスタムリソースを同じように扱うことができ、型安全で簡単に利用することができます。

## クライアントの作成

クライアントを作成するためにはまず[Scheme](https://pkg.go.dev/k8s.io/apimachinery/pkg/runtime?tab=doc#Scheme)を用意する必要があります。

SchemeはGoのstructとGroupVersionKindを相互に変換したり、異なるバージョン間でのSchemeの変換をおこなうための機能です。

[import:"init"](../../codes/tenant/main.go)

最初に`runtime.NewScheme()`で新しい`scheme`を作成します。
`clientgoscheme.AddToScheme`では、PodやServiceなどKubernetesの標準リソースの型をschemeに追加しています。
`multitenancyv1.AddToScheme`では、Tenantカスタムリソースの型をschemeに追加しています。

このSchemeを利用することで、標準リソースとTenantリソースを扱うことができるクライアントを作成できます。

つぎに[GetConfigOrDie](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config?tab=doc#GetConfigOrDie)でクライアントの設定を取得しています。
この関数はコマンドラインオプションの`--kubeconfig`や、環境変数`KUBECONFIG`で指定された設定ファイルを利用するか、またはKubernetesクラスタ上でPodとして動いているのであれば、Podが持つサービスアカウントの認証情報を利用します。
通常コントローラはKubernetesクラスタ上で動いているので、サービスアカウントの認証情報が利用されます。

このSchemeとConfigを利用してManagerを作成し、`GetClient()`でクライアントを取得することができます。
ただし、Managerの`Start()`を呼び出す前にClientを利用することはできないので注意しましょう。

```
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme: scheme,
})
client := mgr.GetClient()
```

## Get/List

クライアントを利用して、リソースを取得する方法を見ていきます。

### Getの使い方

リソースを取得するには、下記のように第2引数で欲しいリソースのnamespaceとnameを指定します。
そして第3引数に指定した変数で結果を受け取ることができます。
なお、どの種類のリソースを取得するのかは、第3引数に渡した変数の型で自動的に判別されます。

[import:"get",unindent="true"](../../codes/tenant/controllers/tenant_controller.go)

### クライアントのキャッシュ機構

Kubernetes上ではいくつものコントローラが動いており、そのコントローラはそれぞれたくさんのリソースを扱っています。
これらのコントローラが毎回APIサーバーにアクセスしてリソースの取得をおこなうと、APIサーバーやそのバックエンドにいるetcdの負荷が高まってしまうという問題があります。

そこで、controller-runtimeの提供するクライアントはキャッシュ機構を備えています。
このクライアントは`Get()`や`List()`でリソースを取得すると、同一namespace内の同じKindのリソースをすべて取得してインメモリにキャッシュします。
そして対象のリソースをWatchし、APIサーバー上でリソースの変更が発生した場合にキャッシュの更新をおこないます。

このようなキャッシュの仕組みにより、コントローラからAPIサーバーへのアクセスを大幅に減らすことが可能になっています。

なお、このようなキャッシュ機構を備えているため、実装上はGetしか呼び出していなくても、リソースのアクセス権限としてはListやWatchが必要となります。
[RBACマニフェストの生成](../controller-tools/rbac.md)で解説したように、リソースの取得をおこなう場合は`get, list, watch`の権限を付与しておきましょう。

キャッシュの仕組みが必要ない場合は、Managerから`GetAPIReader()`でキャッシュを利用しないクライアントを取得することもできます。

### Listの使い方

Listでは条件を指定して複数のリソースを一度に取得することができます。

下記の例では、LabelSelectorやNamespaceを指定してリソースの取得をおこなっています。
なお、Namespaceを指定しなかった場合は、全Namespaceのリソースを取得します。

[import:"list"](../../codes/misc/main.go)

`Limit`と`Continue`を利用することで、ページネーションをおこなうことも可能です。
下記の例では1回のAPI呼び出しで3件ずつリソースを取得して表示しています。

[import:"pagination"](../../codes/misc/main.go)

`.ListMeta.Continue`にトークンが入っているを利用して、続きのリソースを取得することができます。
トークンが空になるとすべてのリソースを取得したということになります。

### インデックス

複数のリソースを取得する際にラベルやnamespaceだけでなく、特定のフィールドの値に応じてフィルタリングしたいことがあるかと思います。
controller-runtimeではインメモリキャッシュにインデックスを張る仕組みが用意されています。

インデックスを利用するためには事前に`GetFieldIndexer().IndexField()`を利用して、どのフィールドの値に基づいてインデックスを張るのかを指定しておきます。
下記の例ではnamespaceリソースに対して、ownerReferenceに指定されているTenantリソースの名前に応じてインデックスを作成しています。

[import:"indexer"](../../codes/tenant/controllers/tenant_controller.go)
[import:"index-field",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

フィールド名には、どのフィールドを利用してインデックスを張っているのかを示す文字列を指定します。
実際にインデックスに利用しているフィールドのパスと一致していなくても問題はないのですが、なるべく一致させたほうが可読性がよくなるのでおすすめです。
なおインデックスはGVKごとに作成されるので、異なるタイプのリソース間でフィールド名が同じになっても問題ありません。
またnamespaceスコープのリソースの場合は、内部的にフィールド名にnamespace名を付与して管理しているので、明示的にフィールド名にnamespaceを含める必要はありません。
インデクサーが返す値はスライスになっていることから分かるように、複数の値にマッチするようにインデックスを構成することも可能です。

上記のようなインデックスを作成しておくと、`List()`を呼び出す際に特定のフィールドが指定した値と一致するリソースだけを取得することができます。
例えば以下の例であれば、ownerReferenceに指定したTenantリソースがセットされているnamespaceだけを取得することができます。

[import:"matching-fields",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

## Create/Update

リソースの作成は以下のように`Create()`を利用します。更新処理の`Update()`も同じように利用できます。

[import:"namespace,create",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

なお、リソースが存在する状態で`Create()`を呼んだり、リソースが存在しない状態で`Update()`を呼び出すとエラーになります。

## CreateOrUpdate

`Get()`でリソースを取得して、リソースが存在しなければ`Create()`、存在すれば`Update()`を呼び出すという処理は頻出パターンです。
そこで、Clientには`CreateOrUpdate()`という便利な関数が用意されています。

[import:"create-or-update",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

リソースが存在した場合、`role`変数に取得したリソースの値が書き込まれます。
その後、第4引数で渡した関数の中でその`role`変数を書き換え、更新処理を実行します。

リソースが存在しない場合は、第4引数で渡した関数を実行した後、リソースの作成処理を実行します。

なお、AnnotationsなどのMetadataはKubernetesの標準コントローラが値を設定するので、以下のように初期化や上書きしてはいけません。

```go
op, err := ctrl.CreateOrUpdate(ctx, r.Client, role, func() error {
	role.Annotations = map[string]string{
		"an1": "test",
	}
	return nil
}
```

Annotationsを追加するときは、以下のようにしましょう。

```go
op, err := ctrl.CreateOrUpdate(ctx, r.Client, role, func() error {
	if role.Annotations == nil {
		role.Annotations = make(map[string]string)
	}
	role.Annotations["an1"] = "test"
	return nil
}
```

## Patch

`Update()`でリソースを更新するには、そのリソースのすべてのフィールドを埋めなくてはなりません。

`Patch()`を利用すると、変更したいフィールドの値を用意するだけでリソースの更新をおこなうことができます。

PatchにはMergePath方式とServer-Side Apply方式があります。
Server-Side Apply方式では、リソースの各フィールドごとに管理者を記録することにより、複数のコントローラやユーザーが同一のリソースを編集した場合に衝突を検知することが可能です。
MergePatch方式ではそのような衝突検知はおこなわれません。

ここではServer-Side Apply方式による`Patch()`の利用方法を紹介します。
以下の例では、Deploymentリソースの`spec.replicas`フィールドのみを更新しています。

[import:"patch"](../../codes/misc/main.go)

Server-Side Applyを利用するには、第3引数に`client.Apply`を指定し、オプションには`FieldManager`を指定する必要があります。
この`FieldManager`がフィールドごとの管理者の名前になるので、他のコントローラと被らないようにユニークな名前にしましょう。

なお、リストやマップをどのようにマージするのかは、Goの構造体に付与したマーカーで制御することが可能です。
詳しくは[Merge strategy](https://kubernetes.io/docs/reference/using-api/api-concepts/#merge-strategy)を参照してください。(TODO: あとで書く)

## Status.Update/Patch

Statusをサブリソース化している場合、これまで紹介した`Update()`や`Patch()`を利用してもステータスを更新することができません。
Status更新用のクライアントを利用することになります。

`Status.Update()`と`Status.Path()`は、メインリソースの`Update()`、`Patch()`と使い方は同じです。
ただし、現状カスタムリソースの Status サブリソースは Server-Side Apply による Patch をサポートされていません。

```go
tenant.Status = multitenancyv1.TenantStatus{
	Conditions: []multitenancyv1.TenantCondition{
		{
			Type: multitenancyv1.ConditionReady, 
			Status: corev1.ConditionTrue,
		},
	},
}
err := r.Status().Update(ctx, &tenant)
```

## Delete/DeleteOfAll

最後にリソースを削除する`Delete`と`DeleteOfAll`を見てみましょう。

`Delete`と`DeleteOfAll`には`Preconditions`と`PropagationPolicy`という特殊なオプションがあるのでそちらを紹介します。

まずは`Preconditions`オプションを利用した例です。

[import:"cond"](../../codes/misc/main.go)

リソースを取得してから削除のリクエストを投げるまでの間にリソースが作り直されてしまう可能性があります。
そこで再作成したリソースを間違って消してしまわないように、UIDとResourceVersionを指定して、確実に指定したリソースを削除しています。

つづいて`PropagationPolicy`オプションを利用した例です。

[import:"policy"](../../codes/misc/main.go)

[リソースの削除](deletion.md)で解説するように、Kubernetesでは親リソースを削除するとそのリソースに結びつく子リソースも一緒に削除されます。
この挙動を変えるためのオプションとして`PropagationPolicy`が用意されています。

上記のようにDeploymentリソースの削除時に`DeletePropagationOrphan`を指定すると、子のリソースであるReplicaSetやPodのリソースが削除されなくなります。

## ディスカバリーベースのクライアント

client-goを利用してCRDを扱う場合、[k8s.io/client-go/dynamic](https://pkg.go.dev/k8s.io/client-go/dynamic?tab=doc)や[k8s.io/apimachinery/pkg/apis/meta/v1/unstructured](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured?tab=doc)による動的型クライアントを利用するか、[kubernetes/code-generator](https://github.com/kubernetes/code-generator)を利用してコード生成をおこなう必要がありました。

しかし、controller-runtimeのClientでは、引数に構造体を渡すだけで標準リソースでもカスタムリソースでもAPIを呼び分けてくれています。
このClientはどのように仕組みになっているのでしょうか。

まずはリフレクションにより`Tenant`構造体から"Tenant"という文字列を取得します。これがKindになります。
さらに[api/v1/groupversion_info.go](../../codes/tenant/api/v1/groupversion_info.go)に埋め込まれた情報をもとにGroupとVersionを取得します。
これでGVKが取得できました。

次にREST APIを叩くためにはREST APIのパスを解決する必要があります。
REST APIのパスは、namespace-scopedのリソースであれば`/apis/{group}/{version}/namespaces/{namespace}/{resource}/{name}`、cluster-scopeのスコープであれば`/apis/{group}/{version}/{resource}/{name}`のようになります。
この情報はCRDに記述されているため、APIサーバーに問い合わせる必要があります。

これらの情報は`kubectl`でも確認することができます。以下のように実行してみましょう。

```
$ kubectl api-resources --api-group="multitenancy.example.com"
NAME      SHORTNAMES   APIGROUP                   NAMESPACED   KIND
tenants                multitenancy.example.com   false        Tenant
```

APIサーバーに問い合わせて取得した情報をもとにREST APIのパスが解決できました。
最後はこのパスに対してリクエストを発行します。

Clientはこのような仕組みによって、標準リソースとカスタムリソースを同じように扱うことができ、型安全で簡単に利用できるクライアントを実現しています。
