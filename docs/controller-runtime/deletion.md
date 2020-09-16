# リソースの削除

テナントコントローラは、テナントリソースが作成されたら、マニフェストに記述されている内容に基づいてnamespaceを作成します。
逆にテナントリソースが削除されたら、作成したnamespaceも削除しなければなりません。

しかし、Reconcileループによる削除処理は難しい問題です。
なぜなら親リソースが削除されたというイベントを取りこぼしてしまうと、その親リソースに関する情報は消えてしまい、どの子リソースを削除すべきか判断できなくなってしまうからです。
このようにイベントドリブンなトリガーで動く処理は、Kubernetesのコンセプトに反していると言えます。

そこでKubernetesでは、ownerReferenceによるガベージコレクションと、Finalizerという2種類のリソース削除の仕組みを提供しています。

## ownerReferenceによるガベージコレクション

1つめのリソース削除の仕組みはownerReferenceによるガベージコレクションの仕組みです。([参考](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/))
これは親リソースが削除されると、そのリソースの子のリソースもガベージコレクションにより自動的に削除されるという仕組みです。

例えば、以下のようにnamespaceリソースを作成する際に、親リソースとしてテナントリソースを指定します。
そのための関数としてcontroller-runtimeでは、[controllerutil.SetControllerReference](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#SetControllerReference)というユーティリティ関数を用意しています。

[import:"namespace,controller-reference",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

すると、作成されたnamespaceリソースには、下記のように`ownerReferences`フィールドが付与されています。

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: tenant-sample-test1
  ownerReferences:
  - apiVersion: multitenancy.example.com/v1
    blockOwnerDeletion: true
    controller: true
    kind: Tenant
    name: tenant-sample
    uid: c296e7b2-6c92-470b-b110-f62eb43b2bbe
spec:
  finalizers:
  - kubernetes
status:
  phase: Active
```

この状態で親のテナントリソースを削除すると、子のnamespaceリソースも自動的に削除されます。

なお、異なるnamespaceのリソースをownerにしたり、cluster-scopedリソースのownerにnamespace-scopedリソースを指定することはできません。
今回のテナントコントローラのようにNamespaceやClusterRoleなどのcluster-scopedリソースを扱う場合は、カスタムリソースもcluster-scopedにする必要があります。

また、`SetControllerReference`と似た関数で[controllerutil.SetOwnerReference](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#SetOwnerReference)もあります。

`SetControllerReference`は、`controller`フィールドと`blockOwnerDeletion`フィールドにtrueが指定されており、1つのリソースに1つのオーナーのみしか指定することができません。また子リソースが削除されるまで親リソースの削除がブロックされます。

一方の`SetOwnerReference`は1つのリソースに複数のオーナーを指定することができ、子リソースの削除はブロックされずバックグラウンドで実施されます。


## Finalizer

### Finalizerの仕組み

ownerReferenceとガベージコレクションにより、親リソースと一緒に子リソースを削除することができると説明しました。
しかし、この仕組だけでは削除できないケースもあります。直接の親ではないリソースを削除したいケースや、Kubernetesで管理していない外部のリソースなどを削除したいケースなどがあります。

例えばTopoLVMでは、LogicalVolumeというカスタムリソースを作成すると、ノード上にLVM(Logical Volume Manager)のLV(Logical Volume)を作成します。
Kubernetes上のLogicalVolumeカスタムリソースが削除されたら、それに合わせてノード上のLVも削除しなければなりません。

そのようなリソースの削除には、Finalizerという仕組みを利用することができます。

Finalizerの仕組みを利用するためには、まず親リソースの`finalizers`フィールドにFinalizerの名前を指定します。
なお、この名前はテナントコントローラが管理しているFinalizerであると識別できるように、他のコントローラと衝突しない名前にしておきましょう。

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  finalizers:
  - tenant.finalizers.multitenancy.example.com
# 以下省略
```

`finalizers`フィールドが付与されているリソースは、リソースを削除しようとしても削除されません。
代わりに、以下のように`deletionTimestamp`が付与されるだけです。

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  finalizers:
  - tenant.finalizers.multitenancy.example.com
  deletionTimestamp: "2020-07-24T15:23:54Z"
# 以下省略
```

カスタムコントローラは`deletionTimestamp`が付与されていることを発見すると、そのリソースに関連するリソースを削除し、その後に`finalizers`フィールドを削除します。
`finalizers`フィールドが空になると、Kubernetesがこのリソースを完全に削除します。

このような仕組みにより、コントローラが削除イベントを取りこぼしたとしても、テナントリソースが削除されるまでは何度もReconcileが呼び出されるため、子のリソースの情報が失われて削除できなくなるという問題を回避できます。
一方で、カスタムリソースよりも先にコントローラを削除してしまった場合は、いつまでたってもカスタムリソースが削除されないという問題が発生することになるので注意しましょう。

### Finalizerの実装方法

それではFinalizerを実装してみましょう。
controller-runtimeでは、Finalizerを扱うためのユーティリティ関数として[controllerutil.ContainsFinalizer](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#ContainsFinalizer)、[controllerutil.AddFinalizer](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#AddFinalizer)、[controllerutil.RemoveFinalizer](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#RemoveFinalizer)などを提供しているのでこれを利用しましょう。

[import:"finalizer",unindent:"true"](../../codes/tenant/controllers/tenant_controller.go)

`deletionTimestamp`が付与されていなければ、`finalizers`フィールドを追加します。

`deletionTimestamp`が付与されていた場合は、`finalizers`に自分で指定した名前が存在した場合はリソースの削除をおこない、その後`finalizers`フィールドをクリアします。
