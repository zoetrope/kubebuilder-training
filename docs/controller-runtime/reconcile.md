# Reconcile

Reconcileはカスタムコントローラのコアロジックになります。
あるべき状態(ユーザーが作成したカスタムリソース)と、実際のシステムの状態を比較し、差分があればそれを埋めるための処理を実行します。

## Reconcileはいつ呼ばれるのか

Reconcile処理は下記のタイミングで呼び出されます。

* コントローラが扱うリソースが作成、更新、削除されたとき
* Reconcileに失敗してリクエストが再度キューに積まれたとき
* コントローラの起動時
* 外部イベントが発生したとき
* キャッシュを再同期するとき(デフォルトでは10時間に1回)

このような様々なタイミングで呼び出されるので、Reconcile処理は必ず冪等(同じリクエストで何度呼び出しても同じ結果になること)でなければなりません。

なお、Reconcile処理はデフォルトでは1秒間に10回以上実行されないように制限されています。

### 監視対象の制御

Reconcileが呼ばれるタイミングを制御するために、`NewControllerManagedBy`関数を利用します。

[import:"managedby",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

`For`ではこのコントローラのReconcile対象となるリソースの型を指定します。

`Owns`にはこのコントローラが生成するリソースの型を指定します。
ここではテナントコントローラが生成するnamespaceとClusterRole,RoleBindingを指定しています。
これらのリソースに何らかの変更が発生した際にReconcileが呼び出されるようになります。
ただし、Ownsで指定した型のすべてのリソースの変更をウォッチするわけではなく、テナントリソースがownerReferenceに指定されているリソースのみが監視対象となります。

### イベントのフィルタリング

`WithEventFilter`では、`For`, `Owns`, `Watches`で監視対象としたリソースの変更イベントをまとめてフィルタリングすることができます。
後述しますが、`For` や `Owns` 個々に、より細かくフィルタリングすることもできます。

下記のような[predicate.Funcs](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/predicate?tab=doc#Funcs)を用意して、`WithEventFilter`関数で指定します。

[import:"pred",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

例えば`CreateFunc`でtrueを返し、`DeleteFunc`,`UpdateFunc`でfalseを返すようにすれば、リソースが作成されたときのみReconcileが呼び出されるようにできます。また、引数でイベントの詳細な情報が渡ってくるので、それを利用してより複雑なフィルタリングをおこなうことも可能です。
なお、`GenericFunc`は後述の外部イベントのフィルタリングに利用します。

重要な注意点として、kube-apiserver に発行した CREATE/UPDATE/PATCH 操作が一対一でイベントにはなりません。
たとえば CREATE 直後に UPDATE すると、イベントとしては CreateFunc しか呼び出されないことがあります。

`WithEventFilter`を利用すると`For`や`Owns`,`Watches`で指定したすべての監視対象にフィルターが適用されますが、下記のように`For`や`Owns`,`Watches`のオプションとして個別にフィルターを指定することも可能です。

```go
return ctrl.NewControllerManagedBy(mgr).
	For(&multitenancyv1.Tenant{}, builder.WithPredicates(pred1)).
	Owns(&corev1.Namespace{}, builder.WithPredicates(pred2)).
	Owns(&rbacv1.RoleBinding{}, builder.WithPredicates(pred3)).
	Watches(&src, &handler.EnqueueRequestForObject{}, builder.WithPredicates(pred4)).
	Complete(r)
```

### 外部イベントの監視

`Watches`では上記以外の外部イベントを監視したい場合に利用します。

Kubernetes内のリソースの変更だけでなく、外部イベントをトリガーにしてReconcileを呼び出したい場合があります。
例えばGitHubのWebhook呼び出しに応じて処理をおこないたい場合や、外部の状態をポーリングしてその状態の変化によって処理をおこないたい場合などが考えられます。

外部イベント監視の例として、ここでは10秒ごとに起動しテナントリソースがReady状態になっていればイベントを発行する仕組みを実装してみます。

[import](../../codes/tenant/controllers/external_event.go)

上記のwatcherを`mgr.Add`でマネージャに登録し、`GenericEvent`をウォッチします。

[import:"external-event,managedby",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

これにより、Ready状態のテナントリソースが存在すると10秒ごとにReconcileが呼び出されるようになります。

## Reconcileの実装

いよいよReconcileの本体を実装します。

### Reconciler

Reconcileは[reconcile.Reconciler](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Reconciler)インタフェースを実装することになります。

```go
type Reconciler interface {
	Reconcile(context.Context, Request) (Result, error)
}
```

引数として渡ってくる[reconcile.Request](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Request)には、`For`で指定した監視対象のNamespacedNameが含まれています。

このNamespacedNameを利用して、テナントリソースの取得をおこないます。

[import:"get",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

なお、`Owns`でnamespaceやClusterRole, RoleBindingを監視対象に設定しましたが、これらのリソースの変更によってReconcileが呼び出された場合でも、RequestのNamespacedNameにはこれらのリソースのownerであるテナントリソースの名前が入っています。

戻り値の[reconcile.Result](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Result)には、`Requeue`, `RequeueAfter`というフィールドが含まれています。
この戻り値を利用すると、指定した時間が経過したあとに再度Reconcileを呼び出させることが可能になります。
例えば何らかの時間がかかる処理(コンテナの起動など)を待つ場合に利用できます。

また、Recnocileがエラーを返した場合は、失敗するたびに待ち時間が指数関数的に増加します。

Reconcileは複数のリソースを管理しているため、1つのリソースを処理するために多くの時間をかけるべきではありません。
何らかの待ちが発生する場合は、`Requeue`や`RequeueAfter`を指定してReconcileをすぐに抜けるようにしましょう。

### reconcileNamespaces

テナントリソースに記述されたnamespaceを作成します。

[import:"reconcile-namespaces"](../../codes/tenant/controllers/tenant_controller.go)

### reconcileRBAC

ClusterRoleとRoleBindingを作成し、テナントの管理対象のnamespaceに管理者権限を付与します。

[import:"reconcile-rbac"](../../codes/tenant/controllers/tenant_controller.go)

### ステータスの更新

最初にステータスを更新するためのヘルパー関数を用意しておきます。

[import](../../codes/tenant/controllers/status.go)

コントローラが扱うリソースに何も変更が加えられなかった場合は、ステータスを更新する必要もないでしょう。
そこで下記のような関数を用意し、namespaceとRBACのどちらかに変更が加えられたことをわかるようにしておきます。

[import:"reconcile"](../../codes/tenant/controllers/tenant_controller.go)

上記の関数の戻り値に応じてステータスの更新をおこないます。

[import:"status",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

これにより、ユーザーはテナントリソースのステータスを確認することが可能になります。
