# 応用テクニック

## 更新処理

なお、Annotationsなどの一部のフィールドはKubernetesの標準コントローラが値を設定する場合があります。
そのため、以下のように値を上書きしてしまうと、他のコントローラが設定した値が消えてしまいます。

```go
op, err := ctrl.CreateOrUpdate(ctx, r.Client, role, func() error {
	role.Annotations = map[string]string{
		"an1": "test",
	}
	return nil
}
```

そのような問題を避けるため、Annotationsを更新する場合は上書きではなく、以下のように追加しましょう。

```go
op, err := ctrl.CreateOrUpdate(ctx, r.Client, role, func() error {
	if role.Annotations == nil {
		role.Annotations = make(map[string]string)
	}
	role.Annotations["an1"] = "test"
	return nil
}
```

### イベントのフィルタリング

`WithEventFilter`では、`For`, `Owns`, `Watches`で監視対象としたリソースの変更イベントをまとめてフィルタリングすることができます。
後述しますが、`For` や `Owns` 個々に、より細かくフィルタリングすることもできます。

下記のような[predicate.Funcs](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/predicate?tab=doc#Funcs)を用意して、`WithEventFilter`関数で指定します。

[import:"pred",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

例えば`CreateFunc`でtrueを返し、`DeleteFunc`,`UpdateFunc`でfalseを返すようにすれば、リソースが作成されたときのみReconcileが呼び出されるようにできます。
また、引数でイベントの詳細な情報が渡ってくるので、それを利用してより複雑なフィルタリングをおこなうことも可能です。
なお、`GenericFunc`は後述の外部イベントのフィルタリングに利用します。

重要な注意点として、kube-apiserver に発行した CREATE/UPDATE/PATCH 操作が一対一でイベントにはなりません。
たとえば CREATE 直後に UPDATE すると、イベントとしては CreateFunc しか呼び出されないことがあります。
また、実際にはリソースが更新されていなくても、コントローラの起動直後は CREATE イベントが発火されます。
このような挙動をするため、イベントをフィルタリングして CREATE と UPDATE で異なる更新処理をおこなう実装はおすすめしません。

なお`WithEventFilter`を利用すると`For`や`Owns`,`Watches`で指定したすべての監視対象にフィルターが適用されますが、
下記のように`For`や`Owns`,`Watches`のオプションとして個別にフィルターを指定することも可能です。

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


### ステータスの更新

例えば`status.phase`フィールドで状態を保持し、その状態に応じて動作するコントローラを考えてみましょう。

最初にコントローラのReconcileが実行されたときには状態はAでした。

```yaml
status:
  phase: A
```

次にコントローラのReconcileが実行されたときには状態はCに変化しました。

```yaml
status:
  phase: C
```

このとき実際にはphaseはA->B->Cと変化したにも関わらず、Bに変化したときのイベントをコントローラが取りこぼしていると、正しく処理ができない可能性があります。

そこで上記のような状態の持たせ方はせずに、各状態のON/OFFをリストで表現すれば、Bに変化したことを取りこぼさずに必要な処理を実行させることが可能になります。

```yaml
status:
  conditions:
  - type: A
    status: True
  - type: B
    status: True
  - type: C
    status: False
```

[API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)

コントローラが扱うリソースに何も変更が加えられなかった場合は、ステータスを更新する必要もないでしょう。
そこで下記のように更新をおこなったかどうかをフラグで返すようにしておきます。

[import:"reconcile"](../../codes/tenant/controllers/tenant_controller.go)

上記の関数の戻り値に応じてステータスの更新をおこないます。
Conditionsの更新には、`meta.SetStatusCondition()`という関数が用意されています。
この関数を利用すると、同じタイプのConditionがすでに存在する場合は値を更新し、存在しない場合は追加してくれます。
また、Condition.Statusの値が変化したときだけ`LastTransitionTime`が現在の時刻で更新されます。

[import:"status",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

これにより、ユーザーはテナントリソースのステータスを確認することが可能になります。


## ディスカバリーベースのクライアント

client-goを利用してCRDを扱う場合、[k8s.io/client-go/dynamic](https://pkg.go.dev/k8s.io/client-go/dynamic?tab=doc)や[k8s.io/apimachinery/pkg/apis/meta/v1/unstructured](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured?tab=doc)による動的型クライアントを利用するか、[kubernetes/code-generator](https://github.com/kubernetes/code-generator)を利用してコード生成をおこなう必要がありました。

しかし、controller-runtimeのClientでは、引数に構造体を渡すだけで標準リソースでもカスタムリソースでもAPIを呼び分けてくれています。
このClientはどのように仕組みになっているのでしょうか。

まずは渡された構造体の型を Scheme に登録された情報から探します。そうすると GVK が得られます。

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

## unstructuredのキャッシュ

## Loggerの使い方

Reconcile内でlogger使えるよ

記事を参考にして。

オプションの変更について
