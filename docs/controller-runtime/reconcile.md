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

TODO: その他の方法については応用編へ。

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

なお、`Owns`でnamespaceやClusterRole, RoleBindingを監視対象に設定しましたが、これらのリソースの変更によってReconcileが呼び出された場合でも、
RequestのNamespacedNameにはこれらのリソースのownerであるテナントリソースの名前が入っています。

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

最後に、テナントリソースの状況をユーザーに知らせるためにステータスの更新をおこないます。

コントローラが扱うリソースに何も変更が加えられなかった場合は、ステータスを更新する必要もないでしょう。
そこで下記のように更新をおこなったかどうかをフラグで返すようにしておきます。

[import:"reconcile"](../../codes/tenant/controllers/tenant_controller.go)

上記の関数の戻り値に応じてステータスの更新をおこないます。
Conditionsの更新には、`meta.SetStatusCondition()`という関数が用意されています。
この関数を利用すると、同じタイプのConditionがすでに存在する場合は値を更新し、存在しない場合は追加してくれます。
また、Condition.Statusの値が変化したときだけ`LastTransitionTime`が現在の時刻で更新されます。

[import:"status",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

これにより、ユーザーはテナントリソースのステータスを確認することが可能になります。
