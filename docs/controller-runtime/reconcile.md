# Reconcile

Reconcileはカスタムコントローラーのコアロジックです。
あるべき状態(ユーザーが作成したカスタムリソース)と、実際のシステムの状態を比較し、差分があればそれを埋めるための処理を実行します。

## Reconcilerの仕組み

### Reconcilerインタフェース

controller-runtimeでは、Reconcile処理は[reconcile.Reconciler](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Reconciler)インタフェースを実装することになります。

```go
type Reconciler interface {
	Reconcile(context.Context, Request) (Result, error)
}
```

引数の[reconcile.Request](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Request)には、
このReconcilerが対象とするカスタムリソースのNamespaceとNameが入っています。

戻り値の[reconcile.Result](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Result)には、
`Requeue`, `RequeueAfter`というフィールドがあります。
RequeueにTrueを指定して戻り値を返すと、Reconcile処理がキューに積まれて再度実行されることになります。
RequeueAfterを指定した場合は、指定した時間が経過したあとに再度Reconcile処理が実行されます。

また、Recnocileがエラーを返した場合もReconcile処理がキューに積まれて再度実行されることになるのですが、
失敗するたびに待ち時間が指数関数的に増加します。

Reconcileは複数のリソースを管理しているため、1つのリソースを処理するために多くの時間をかけるべきではありません。
何らかの待ちが発生する場合は、`Requeue`や`RequeueAfter`を指定してReconcileをすぐに抜けるようにしましょう。

### Reconcileの実行タイミング

Reconcile処理は下記のタイミングで呼び出されます。

* コントローラーの扱うリソースが作成、更新、削除されたとき
* Reconcileに失敗してリクエストが再度キューに積まれたとき
* コントローラーの起動時
* 外部イベントが発生したとき
* キャッシュを再同期するとき(デフォルトでは10時間に1回)

このような様々なタイミングで呼び出されるので、Reconcile処理は必ず冪等(同じリクエストで何度呼び出しても同じ結果になること)でなければなりません。

なお、Reconcile処理はデフォルトでは1秒間に10回以上実行されないように制限されています。

また、これらのイベントが高い頻度で発生する場合は、Reconciliation Loopを並列実行するように設定可能です。

### 監視対象の制御

Reconcile処理は、コントローラーの扱うリソースが作成、更新、削除されたときに呼び出されると説明しました。
「コントローラーの扱うリソース」を指定するために、[NewControllerManagedBy](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/builder#ControllerManagedBy)関数を利用します。

[import:"managedby",unindent:"true"](../../codes/markdown-view/controllers/markdownview_controller.go)

#### For

`For`にはこのコントローラーのReconcile対象となるリソースの型を指定します。

今回はMarkdownViewカスタムリソースを指定します。
これによりMarkdownViewリソースの作成・変更・削除がおこなわれると、Reconcile関数が呼び出されることになります。
そして、Reconcile関数の引数で渡されるRequestは、MarkdownViewの情報になります。

なお、`For`に指定できるリソースは1種類だけです。

#### Owns

`Owns`にはこのコントローラーが生成するリソースの型を指定します。`For`とは異なり、`Owns`は複数指定が可能です。

MarkdownViewコントローラーは、ConfigMap, Deployment, Serviceリソースを作成することになるため、これらを`Owns`に指定します。

これにより、MarkdownViewコントローラーが作成したConfigMap, Deployment, Serviceリソースに何らかの変更が発生した際にReconcileが呼び出されるようになります。
ただしこのとき、コントローラーが作成したリソースの`ownerReferences`にMarkdownViewリソースを指定しなければなりません。
`ownerReferences`の設定方法は[リソースの削除](./deletion.md))を参照してください。

なお、`Owns`に指定したリソースの変更によってReconcileが呼び出された場合でも、
RequestにはこれらのリソースのownerであるMarkdownViewリソースの名前が入っています。

## Reconcileの実装

いよいよReconcileの本体を実装します。

### Reconcile処理の流れ

Reconcile処理のおおまかな流れを確認しましょう。

[import:"reconcile",unindent:"true"](../../codes/markdown-view/controllers/markdownview_controller.go)

Reconcileの引数として渡ってきたRequestを利用して、対象となるMarkdownViewリソースの取得をおこないます。

ここでMarkdownViewリソースが存在しなかった場合は、MarkdownViewリソースが削除されたということです。
終了処理をおこなって関数を抜けましょう。(ここではメトリクスの削除処理をおこなっています)

次に`DeletionTimestamp`の確認をしています。
`DeletionTimestamp`がゼロでない場合は、対象のリソースの削除が開始されたということです。(詳しくは[リソースの削除](./deletion.md)を参照してください。)
この場合もすぐに関数を抜けましょう。

そして、`reconcileConfigMap`, `reconcileDeployment`, `reconcileService`で、それぞれConfigMap, Deployment, Serviceリソースの作成・更新処理をおこないます。

最後に`updateStatus`でステータスの更新をおこないます。

### reconcileConfigMap

`reconcileConfigMap`では、MarkdownViewリソースに記述されたMarkdownの内容をもとに、ConfigMapリソースを作成します。

[import:"reconcile-configmap"](../../codes/markdown-view/controllers/markdownview_controller.go)

ここでは、[クライアントの使い方](./client.md)で紹介した`CreateOrUpdate`関数を利用しています。

### reconcileDeployment, reconcileService

`reconcileDeployment`, `reconcileService`では、それぞれDeploymentとServiceリソースを作成します。

`reconcileConfigMap`と同様に`CreateOrUpdate`を利用したリソースの作成も可能なのですが、
DeploymentやServiceリソースはフィールド数が多いこともあり、適切に差分を検出してリソースを更新することが面倒だったりします。

そこで今回は、[クライアントの使い方](./client.md)で紹介したApplyConfigurationを利用したServer-Side Apply方式でリソースを作成します。

[import:"reconcile-service"](../../codes/markdown-view/controllers/markdownview_controller.go)

### ステータスの更新

最後に、MarkdownViewリソースの状況をユーザーに知らせるためのステータスを更新します。

[import:"update-status"](../../codes/markdown-view/controllers/markdownview_controller.go)

ここでは、`reconcileDeployment`で作成したDeploymentリソースをチェックし、その状態に応じてMarkdownViewリソースの
ステータスを決定しています。

## 動作確認

Reconcile処理の実装が完了したら動作確認してみましょう。
[カスタムコントローラーの動作確認](../kubebuilder/kind.md)の手順通りにカスタムコントローラーをデプロイし、
サンプルのMarkdownViewリソースを適用します。

Deployment, Service, ConfigMapリソースが生成され、MarkdownViewリソースの状態がHealthyになっていることを確認しましょう。

```
$ kubectl get deployment,service,configmap
NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/viewer-markdownview-sample   1/1     1            1           177m

NAME                                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
service/viewer-markdownview-sample   ClusterIP   10.96.162.90   <none>        80/TCP    177m

NAME                                      DATA   AGE
configmap/markdowns-markdownview-sample   2      177m

$ kubectl get markdownview markdownview-sample
NAME                  REPLICAS   STATUS
markdownview-sample   1          Healthy
```

次にローカル環境から作成されたサービスにアクセスするため、Port Forwardをおこないます。

```
$ kubectl port-forward svc/viewer-markdownview-sample 3000:80
```

最後にブラウザで`http://localhost:3000`にアクセスしてください。
以下のようにレンダリングされたMarkdownが表示されれば成功です。

![index](./img/mdbook.png)
