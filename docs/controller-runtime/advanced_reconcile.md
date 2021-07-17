# Reconcile (応用編)

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
