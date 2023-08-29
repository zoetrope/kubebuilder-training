# クライアントの使い方

カスタムコントローラーを実装する前に、Kubernetes APIにアクセスするためのクライアントライブラリの使い方を確認しましょう。

controller-runtimeでは、Kubernetes APIにアクセスするためのクライアントライブラリ([client.Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client?tab=doc#Client))を提供しています。

このクライアントは標準リソースとカスタムリソースを同じように扱うことができ、型安全で簡単に利用できます。

## クライアントの作成

クライアントを作成するためにはまず[Scheme](https://pkg.go.dev/k8s.io/apimachinery/pkg/runtime?tab=doc#Scheme)を用意する必要があります。

SchemeはGoのstructとGroupVersionKindを相互に変換したり、異なるバージョン間でのSchemeの変換をおこなったりするための機能です。

kubebuilderが生成したコードでは、以下のように初期化処理をおこなっています。

[import:"init"](../../codes/30_client/cmd/main.go)

最初に`runtime.NewScheme()`で新しい`scheme`を作成します。
`clientgoscheme.AddToScheme`では、PodやServiceなどKubernetesの標準リソースの型をschemeに追加しています。
`viewv1.AddToScheme`では、MarkdownViewカスタムリソースの型をschemeに追加しています。

このSchemeを利用することで、標準リソースとMarkdownViewリソースを扱うことができるクライアントを作成できます。

つぎに[GetConfigOrDie](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client/config?tab=doc#GetConfigOrDie)でクライアントの設定を取得しています。

[import:"new-manager"](../../codes/30_client/cmd/main.go)

GetConfigOrDie関数は、下記のいずれかの設定を読み込みます。

- コマンドラインオプションの`--kubeconfig`指定された設定ファイル
- 環境変数`KUBECONFIG`で指定された設定ファイル
- Kubernetesクラスター上でPodとして動いているのであれば、カスタムコントローラーが持つサービスアカウントの認証情報を利用

カスタムコントローラーは通常Kubernetesクラスター上で動いているので、サービスアカウントの認証情報が利用されます。

このSchemeとConfigを利用してManagerを作成し、`GetClient()`でクライアントを取得できます。

以下のようにManagerから取得したクライアントを、MarkdownViewReconcilerに渡します。

[import:"init-reconciler"](../../codes/30_client/cmd/main.go)

ただし、Managerの`Start()`を呼び出す前にクライアントは利用できないので注意しましょう。

## Reconcile関数の中でクライアントの使い方

Managerから渡されたクライアントは、以下のようにMarkdownViewReconcilerの埋め込みフィールドとなります。

[import:"reconciler"](../../codes/30_client/internal/controller/markdownview_controller.go)

そのため、Reconcile関数内ではクライアントのメソッドを`r.Get(...)`や`r.Create(...)`のように呼び出すことができます。

また、クライアントを利用してDeploymentやServiceなどKubernetesの標準リソースを扱う際には、利用したいリソースのグループバージョンに
応じたパッケージをimportする必要があります。
例えばDeploymentリソースであれば`"k8s.io/api/apps/v1"`パッケージ、Serviceリソースであれば`"k8s.io/api/core/v1"`パッケージが必要となります。

しかし、これをそのままインポートすると`v1`というパッケージ名が衝突してしまうため、
`import appsv1 "k8s.io/api/apps/v1"`のようにエイリアスをつけてインポートするのが一般的です。

本ページのサンプルでは、以下のimportを利用します。

[import:"import"](../../codes/30_client/internal/controller/markdownview_controller.go)

## Get/List

クライアントを利用して、リソースを取得する方法を見ていきます。

### Getの使い方

リソースを取得するには、下記のように第2引数で欲しいリソースのnamespaceとnameを指定します。
そして第3引数に指定した変数で結果を受け取ることができます。
なお、どの種類のリソースを取得するのかは、第3引数に渡した変数の型で自動的に判別されます。

[import:"get"](../../codes/30_client/internal/controller/markdownview_controller.go)

### クライアントのキャッシュ機構

Kubernetes上ではいくつものコントローラーが動いており、そのコントローラーはそれぞれたくさんのリソースを扱っています。
これらのコントローラーが毎回APIサーバーにアクセスしてリソースの取得をおこなうと、APIサーバーやそのバックエンドにいるetcdの負荷が高まってしまうという問題があります。

そこで、controller-runtimeの提供するクライアントはキャッシュ機構を備えています。
このクライアントは`Get()`や`List()`でリソースを取得すると、同一namespace内の同じKindのリソースをすべて取得してインメモリにキャッシュします。
そして対象のリソースをWatchし、APIサーバー上でリソースの変更が発生した場合にキャッシュの更新をおこないます。

![cache](./img/cache.png)

このようなキャッシュの仕組みにより、コントローラーからAPIサーバーへのアクセスを減らすことが可能になっています。

なお、このようなキャッシュ機構を備えているため、実装上はGetしか呼び出していなくても、リソースのアクセス権限としてはListやWatchが必要となります。
[RBACマニフェストの生成](../controller-tools/rbac.md)で解説したように、リソースの取得をおこなう場合は`get, list, watch`の権限を付与しておきましょう。

キャッシュの仕組みが必要ない場合は、Managerの`GetAPIReader()`を利用してキャッシュ機能のないクライアントを取得できます。

### Listの使い方

Listでは条件を指定して複数のリソースを一度に取得できます。

下記の例では、LabelSelectorやNamespaceを指定してリソースの取得をおこなっています。
なお、Namespaceを指定しなかった場合は、全Namespaceのリソースを取得します。

[import:"list"](../../codes/30_client/internal/controller/markdownview_controller.go)

`Limit`と`Continue`を利用することで、ページネーションをおこなうことも可能です。
下記の例では1回のAPI呼び出しで3件ずつリソースを取得して表示しています。

[import:"pagination"](../../codes/30_client/internal/controller/markdownview_controller.go)

`.ListMeta.Continue`にトークンが入っているを利用して、続きのリソースを取得できます。
トークンが空になるとすべてのリソースを取得したということになります。

## Create/Update

リソースの作成は`Create()`、更新には`Update()`を利用します。
例えば、Deploymentリソースは以下のように作成できます。

[import:"create",unindent:"true"](../../codes/30_client/internal/controller/markdownview_controller.go)

なお、リソースがすでに存在する状態で`Create()`を呼んだり、リソースが存在しない状態で`Update()`を呼び出したりするとエラーになります。

## CreateOrUpdate

`Get()`でリソースを取得して、リソースが存在しなければ`Create()`を呼び、存在すれば`Update()`を呼び出すという処理は頻出パターンです。
そこで、controller-runtimeには`CreateOrUpdate()`という便利な関数が用意されています。

[import:"create-or-update",unindent:"true"](../../codes/30_client/internal/controller/markdownview_controller.go)

この関数の第3引数に渡すオブジェクトには、NameとNamespaceのみを指定します(ただしクラスターリソースの場合はNamespace不要)。

リソースが存在した場合、この第3引数で渡した変数に既存のリソースの値がセットされます。
その後、第4引数で渡した関数の中でその`svc`変数を書き換え、更新処理を実行します。

リソースが存在しない場合は、第4引数で渡した関数を実行した後、リソースの作成処理が実行されます。

## Patch

`Update()`や`CreateOrUpdate()`による更新処理は、リソースを取得してから更新するまでの間に、他の誰かがリソースを書き換えてしまう可能性があります
(これをTOCTTOU: Time of check to time of useと呼びます)。

すでに書き換えられたリソースを更新しようとすると、以下のようなエラーが発生してしまいます。

```
Operation cannot be fulfilled on deployments.apps "sample": the object has been modified; please apply your changes to the latest version and try again
```

そこで`Patch()`を利用すると、競合することなく変更したいフィールドの値だけを更新できます。

Patchには`client.MergeFrom`や`client.StrategicMergeFrom`を利用する方法と, Server-Side Applyを利用する方法があります。

`client.MergeFrom`と`client.StrategicMergeFrom`の違いは、リスト要素の更新方法です。
`client.MergeFrom`でリストを更新すると指定した要素で上書きされますが、`client.StrategicMergeFrom`ではリストはpatchStrategyに応じて
要素が追加されたり更新されたりします。

`client.MergeFrom`を利用してDeploymentのレプリカ数のみを更新する例を以下に示します。

[import:"patch-merge"](../../codes/30_client/internal/controller/markdownview_controller.go)

一方のServer-Side ApplyはKubernetes v1.14で導入されたリソースの更新方法です。
リソースの各フィールドを更新したコンポーネントを`.metadata.managedFields`で管理することで、
サーバーサイドでリソース更新の衝突を検出できます。

Server-Side Applyでは、以下のようにUnstructured型のパッチを用意してリソースの更新をおこないます。

なお、[公式ドキュメントに記述](https://kubernetes.io/docs/reference/using-api/server-side-apply/#using-server-side-apply-in-a-controller)されているように、
カスタムコントローラでServer-Side Applyをおこなう際には、常にForceオプションを有効にすることが推奨されています。

[import:"patch-apply"](../../codes/30_client/internal/controller/markdownview_controller.go)

上記のようにServer-Side ApplyはUnstructured型を利用するため、型安全なコードが記述できませんでした。

Kubernetes v1.21からApplyConfigurationが導入され、以下のように型安全なServer-Side Applyのコードが書けるようになりました。

[import:"patch-apply-config"](../../codes/30_client/internal/controller/markdownview_controller.go)

## Status.Update/Patch

Statusをサブリソース化している場合、これまで紹介した`Update()`や`Patch()`を利用してもステータスを更新できません。
Status更新用のクライアントを利用することになります。

`Status().Update()`と`Status().Patch()`は、メインリソースの`Update()`、`Patch()`と使い方は同じです。
以下のようにstatusフィールドを変更し、`Status().Update()`を呼び出します。
(このコードはあくまでもサンプルです。Deploymentリソースのステータスを勝手に書き換えるべきではありません。)

[import:"update-status"](../../codes/30_client/internal/controller/markdownview_controller.go)

## Delete/DeleteAllOf

最後にリソースを削除する`Delete`と`DeleteAllOf`を見てみましょう。

`Delete`と`DeleteAllOf`には`Preconditions`という特殊なオプションがあります。
以下のコードは`Preconditions`オプションを利用した例です。

[import:"cond"](../../codes/30_client/internal/controller/markdownview_controller.go)

リソースを削除する際、リソース取得してから削除のリクエストを投げるまでの間に、同じ名前の別のリソースが作り直される場合があります。
そのようなケースでは、NameとNamespaceのみを指定してDeleteを呼び出した場合、誤って新しく作成されたリソースを削除される可能性があります。
そこでこの例では再作成したリソースを間違って消してしまわないように、`Preconditions`オプションを利用してUIDとResourceVersionが一致するリソースを削除しています。

`DeleteAllOf`は、以下のように指定した種類のリソースをまとめて削除できます。

[import:"delete-all-of"](../../codes/30_client/internal/controller/markdownview_controller.go)

なお、Serviceリソースなど`DeleteAllOf`が利用できないリソースもあるので注意しましょう。
