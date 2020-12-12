# コントローラ実装入門

controller-runtimeは非常にたくさんの機能を提供しています。
ここではまずcontroller-runtimeの基本的な機能に触れ、簡単なコントローラの実装方法を学んでいきます。
より詳細を知りたい方は次ページ以降をご覧ください。

## Reconcileの実装

まずはKubebuilderによって生成された`controllers/tenant_controller.go`を開いてみましょう。
以下のような`Reconcile`という関数がみつかると思います。
カスタムコントローラの実装は、この`Reconcile`関数の中に書いていくことになります。

```go
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("tenant", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}
```

この`Reconcile`関数は、Tenantカスタムリソースを作成したり、更新・削除をおこなったタイミングで呼び出されます。
Kubernetesクラスタの状態をTenantリソースで指定された内容と一致させるための処理を`Reconcile`関数に実装していきます。

## リソースの取得

`Reconcile`関数の引数である[reconcile.Request](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Request)には、このカスタムコントローラの管理対象であるTenantリソースのNamespaceとNameが含まれています。
この情報を利用して、APIサーバーからテナントリソースの情報を取得してみましょう。

```go
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tenant", req.NamespacedName)

	var tenant multitenancyv1.Tenant
	err := r.Get(ctx, req.NamespacedName, &tenant)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("got tenant", "tenant", tenant)

	return ctrl.Result{}, nil
}
```

`make run`などで実行してみると、取得したTenantリソースの内容がログに出力されていることがわかります。

## リソースの作成

つぎに、Tenantリソースに記述された内容に応じてリソースの作成をおこなってみましょう。

前述したように`Reconcile`関数は冪等、すなわち何度呼び出されても同じ結果になるように実装しなければなりません。
そこでcontroller-runtimeには、リソースが存在しなければ作成し、存在すれば更新する`CreateOrUpdate()`という便利な関数が用意されています。

この関数を利用して、TenantリソースのNamespacesフィールドに指定された名前のNamespaceリソースを作成してみましょう。
なお、`CreateOrUpdate`関数の第4引数にはリソースの更新処理をおこなうための関数を指定します。

```go
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tenant", req.NamespacedName)

	var tenant multitenancyv1.Tenant
	err := r.Get(ctx, req.NamespacedName, &tenant)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("got tenant", "tenant", tenant)

	for _, name := range tenant.Spec.Namespaces {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		_, err = ctrl.CreateOrUpdate(ctx, r, ns, func() error {
			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
```

下記のようなTenantリソースのマニフェストをKubernetesクラスタに適用し、上記のカスタムコントローラを実行してみましょう。
`sample1`, `sample2`という名前のNamespaceが作成されていることが確認できるでしょう。

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  name: tenant-sample
spec:
  namespaces:
    - sample1
    - sample2
```

## リソースの削除

作成したリソースは1つ1つ`Delete`関数を呼び出して削除することもできますが、Kubernetesにはリソースをガベージコレクションするための機能が用意されています。
詳細は[リソースの削除](./deletion.md)で解説しますが、ガベージコレクション機能を利用すると、ある親リソースが削除されるとそのリソースを親に持つ子リソースが自動的に削除されます。

controller-runtimeでは、リソースに親リソースを指定するために[controllerutil.SetControllerReference](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#SetControllerReference)という関数が用意されています。

下記のように、`CreateOrUpdate`の第4引数に指定した関数内で`controllerutil.SetControllerReference`を呼んでみます。

```go
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tenant", req.NamespacedName)

	var tenant multitenancyv1.Tenant
	err := r.Get(ctx, req.NamespacedName, &tenant)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("got tenant", "tenant", tenant)

	for _, name := range tenant.Spec.Namespaces {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		_, err = ctrl.CreateOrUpdate(ctx, r, ns, func() error {
			return ctrl.SetControllerReference(&tenant, ns, r.Scheme)
		})
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
```

この状態で先ほど作成した`tenant-sample`というリソースを削除してみましょう。
すると、そのリソースを親に持つ`sample1`, `sample2`というNamespaceリソースも自動的に削除されることが確認できます。

## ステータスの更新

カスタムコントローラが何らかの処理をおこなったら、その結果を利用者に知らせることは重要です。
そのような情報を伝えるためにカスタムリソースのステータスを利用します。

先ほど利用した`CreateOrUpdate`関数は、リソースの作成や更新をおこなったのか、それとも何もおこなわなかったのかを戻り値で返します。
これを利用して、リソースの作成や更新処理がおこなわれた場合にステータスを更新してみましょう。

```go
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tenant", req.NamespacedName)

	var tenant multitenancyv1.Tenant
	err := r.Get(ctx, req.NamespacedName, &tenant)
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("got tenant", "tenant", tenant)

	updated := false
	for _, name := range tenant.Spec.Namespaces {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		op, err := ctrl.CreateOrUpdate(ctx, r, ns, func() error {
			return ctrl.SetControllerReference(&tenant, ns, r.Scheme)
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		if op != controllerutil.OperationResultNone {
			updated = true
		}
	}
	if updated {
		tenant.Status = multitenancyv1.TenantStatus{
			Conditions: []multitenancyv1.TenantCondition{
				{
					Type:               multitenancyv1.ConditionReady,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now()),
				},
			},
		}
		err := r.Status().Update(ctx, &tenant)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
```

再度Tenantリソースを作成してみましょう。
成功すると、Tenantリソースのステータスが更新されていることが確認できます。

以上がカスタムコントローラ実装の基本となります。
Kubernetesやcontroller-runtimeには、カスタムコントローラを実装するために必要な機能が他にもたくさん用意されています。
以降のページではそれらの詳細について解説していきます。
