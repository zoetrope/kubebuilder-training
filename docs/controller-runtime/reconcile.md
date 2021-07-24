# Reconcile

Reconcileはカスタムコントローラのコアロジックになります。
あるべき状態(ユーザーが作成したカスタムリソース)と、実際のシステムの状態を比較し、差分があればそれを埋めるための処理を実行します。

## Reconcilerの仕組み

### Reconcilerインタフェース

Reconcile処理は[reconcile.Reconciler](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Reconciler)インタフェースを実装することになります。

```go
type Reconciler interface {
	Reconcile(context.Context, Request) (Result, error)
}
```

引数として渡ってくる[reconcile.Request](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Request)には、`For`で指定した監視対象のNamespacedNameが含まれています。


なお、`Owns`でnamespaceやClusterRole, RoleBindingを監視対象に設定しましたが、これらのリソースの変更によってReconcileが呼び出された場合でも、
RequestのNamespacedNameにはこれらのリソースのownerであるテナントリソースの名前が入っています。

戻り値の[reconcile.Result](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Result)には、`Requeue`, `RequeueAfter`というフィールドが含まれています。
この戻り値を利用すると、指定した時間が経過したあとに再度Reconcileを呼び出させることが可能になります。
例えば何らかの時間がかかる処理(コンテナの起動など)を待つ場合に利用できます。

また、Recnocileがエラーを返した場合は、失敗するたびに待ち時間が指数関数的に増加します。

Reconcileは複数のリソースを管理しているため、1つのリソースを処理するために多くの時間をかけるべきではありません。
何らかの待ちが発生する場合は、`Requeue`や`RequeueAfter`を指定してReconcileをすぐに抜けるようにしましょう。

### Reconcileの実行タイミング

Reconcile処理は下記のタイミングで呼び出されます。

* コントローラが扱うリソースが作成、更新、削除されたとき
* Reconcileに失敗してリクエストが再度キューに積まれたとき
* コントローラの起動時
* 外部イベントが発生したとき
* キャッシュを再同期するとき(デフォルトでは10時間に1回)

このような様々なタイミングで呼び出されるので、Reconcile処理は必ず冪等(同じリクエストで何度呼び出しても同じ結果になること)でなければなりません。

なお、Reconcile処理はデフォルトでは1秒間に10回以上実行されないように制限されています。

また、これらのイベントが高い頻度で発生する場合は、Reconciliation Loopを並列実行するように設定することも可能です

### 監視対象の制御

前節で、Reconcileが呼ばれるタイミングとして

> - コントローラが扱うリソースが作成、更新、削除されたとき

この、コントローラが扱うリソースを伝える方法が
Reconcileが呼ばれるタイミングを制御するために、`NewControllerManagedBy`関数を利用します。

[import:"managedby",unindent:"true"](../../codes/markdown-viewer/controllers/markdownview_controller.go)

`For`ではこのコントローラのReconcile対象となるリソースの型を指定します。

`Owns`にはこのコントローラが生成するリソースの型を指定します。
ここではテナントコントローラが生成するnamespaceとClusterRole,RoleBindingを指定しています。
これらのリソースに何らかの変更が発生した際にReconcileが呼び出されるようになります。
ただし、Ownsで指定した型のすべてのリソースの変更をウォッチするわけではなく、テナントリソースがownerReferenceに指定されているリソースのみが監視対象となります。

TODO: その他の方法については応用編へ。

## Reconcileの実装

いよいよReconcileの本体を実装します。

### Reconcile処理の流れ

[import:"reconcile",unindent:"true"](../../codes/markdown-viewer/controllers/markdownview_controller.go)

このNamespacedNameを利用して、テナントリソースの取得をおこないます。

このとき、NotFoundだった場合
Reconcileが呼び出されたのに、引数で渡されたRequestの対象のリソースはもう存在しない場合。
リソースを削除した場合に発生することがある。
ここでエラーを返すとエラーログがうるさくなるので、`Requeue: true`で返しておくとよいでしょう。

また、`DeletionTimestamp.IsZero()`は、リソースの削除中。
後述するようにFinalizerで自前の終了処理を実装することもできます。

reconcile

最後にupdateStatusでステータスの更新をおこないます。

### reconcileConfigMap

テナントリソースに記述されたnamespaceを作成します。

[import:"reconcile-configmap"](../../codes/markdown-viewer/controllers/markdownview_controller.go)


### reconcileDeployment, reconcileService

CreateOrUpdateを利用した場合、DeploymentやServiceを適切に作成することは意外と面倒だったりします。


[import:"reconcile-service"](../../codes/markdown-viewer/controllers/markdownview_controller.go)

### ステータスの更新

最後に、テナントリソースの状況をユーザーに知らせるためにステータスの更新をおこないます。

[import:"update-status"](../../codes/markdown-viewer/controllers/markdownview_controller.go)
