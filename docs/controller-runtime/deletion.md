# リソースの削除

ここではKubernetesにおけるリソースの削除処理について解説します。

実はコントローラーにおいて削除処理は難しい問題です。
例えばMarkdownViewリソースが削除されたら、そのMarkdownViewに紐付く形で作成されたConfigMap, Deployment, Serviceリソースも一緒に削除しなければなりません。
しかし、もしMarkdownViewが削除されたというイベントを取りこぼしてしまうと、そのリソースに関する情報は消えてしまい、関連するどのリソースを削除すべきか判断できなくなってしまうからです。

そこでKubernetesでは、ownerReferenceによるガベージコレクションと、Finalizerというリソース削除の仕組みを提供しています。

## ownerReferenceによるガベージコレクション

1つめのリソース削除の仕組みはownerReferenceによるガベージコレクションです。([参考](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/))
これは親リソースが削除されると、そのリソースの子リソースもガベージコレクションにより自動的に削除されるという仕組みです。

Kubernetesではリソースの親子関係を表すために`.metadata.ownerReferences`フィールドを利用します。

controller-runtimeが提供している[controllerutil.SetControllerReference](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#SetControllerReference)
関数を利用することで、指定したリソースにownerReferenceを設定することができます。

先ほど作成した、`reconcileConfigMap`関数で`controllerutil.SetControllerReference`を利用してみましょう。

[import:"reconcile-configmap",unindent:"true"](../../codes/50_completed/controllers/markdownview_controller.go)

この関数を利用すると、ConfigMapリソースに以下のような`.metadata.ownerReferences`が付与され、このリソースに親リソースの情報が設定されます。

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: "2021-07-25T09:35:43Z"
  name: markdowns-markdownview-sample
  namespace: default
  ownerReferences:
  - apiVersion: view.zoetrope.github.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: MarkdownView
    name: markdownview-sample
    uid: 8e8701a6-fa67-4ab8-8e0c-29c21ae6e1ec
  resourceVersion: "17582"
  uid: 8803226f-7d8f-4632-b3eb-e47dc36eabf3
data:
  ・・省略・・
```

この状態で親のMarkdownViewリソースを削除すると、子のConfigMapリソースも自動的に削除されます。

なお、異なるnamespaceのリソースをownerにしたり、cluster-scopedリソースのownerにnamespace-scopedリソースを指定することはできません。

また、`SetControllerReference`と似た関数で[controllerutil.SetOwnerReference](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#SetOwnerReference)もあります。
`SetControllerReference`は、1つのリソースに1つのオーナーのみしか指定できず、`controller`フィールドと`blockOwnerDeletion`フィールドにtrueが指定されているため子リソースが削除されるまで親リソースの削除がブロックされます。
一方の`SetOwnerReference`は1つのリソースに複数のオーナーを指定でき、子リソースの削除はブロックされません。

`controllerutil.SetControllerReference`は、Server-Side Applyで利用するApplyConfiguration型には対応していません。
そこで、以下のような補助関数を用意しましょう。

[import:"controller-reference",unindent:"true"](../../codes/50_completed/controllers/markdownview_controller.go)

Server-Side Applyでガベージコレクションを利用する際は、この補助関数を利用してApplyConfiguration型を作成するときに
ownerReferenceを設定します。

[import:"service-apply-configuration",unindent:"true"](../../codes/50_completed/controllers/markdownview_controller.go)

## Finalizer

### Finalizerの仕組み

ownerReferenceとガベージコレクションにより、親リソースと一緒に子リソースを削除できると説明しました。
しかし、この仕組だけでは削除できないケースもあります。
例えば、親リソースと異なるnamespaceやスコープの子リソースを削除したい場合や、Kubernetesで管理していない外部のリソースを削除したい場合
などは、ガベージコレクション機能は利用できません。

例えばTopoLVMでは、LogicalVolumeというカスタムリソースを作成すると、ノード上にLVM(Logical Volume Manager)のLV(Logical Volume)を作成します。
Kubernetes上のLogicalVolumeカスタムリソースが削除されたら、それに合わせてノード上のLVも削除しなければなりません。

そのようなリソースの削除には、Finalizerという仕組みを利用できます。

Finalizerの仕組みを利用するためには、まず親リソースの`finalizers`フィールドにFinalizerの名前を指定します。
なお、この名前はMarkdownViewコントローラーが管理しているFinalizerであると識別できるように、他のコントローラーと衝突しない名前にしておきましょう。

```yaml
apiVersion: view.zoetrope.github.io/v1
kind: MarkdownView
metadata:
  finalizers:
  - markdownview.finalizers.view.zoetrope.github.io
# 以下省略
```

`finalizers`フィールドが付与されているリソースは、リソースを削除しようとしても削除されません。
代わりに、以下のように`deletionTimestamp`が付与されるだけです。

```yaml
apiVersion: view.zoetrope.github.io/v1
kind: MarkdownView
metadata:
  finalizers:
    - markdownview.finalizers.view.zoetrope.github.io
  deletionTimestamp: "2021-07-24T15:23:54Z"
# 以下省略
```

カスタムコントローラーは`deletionTimestamp`が付与されていることを発見すると、そのリソースに関連するリソースを削除し、その後に`finalizers`フィールドを削除します。
`finalizers`フィールドが空になると、Kubernetesがこのリソースを完全に削除します。

このような仕組みにより、コントローラーが削除イベントを取りこぼしたとしても、対象のリソースが削除されるまでは何度もReconcileが呼び出されるため、子のリソースの情報が失われて削除できなくなるという問題を回避できます。
一方で、カスタムリソースよりも先にコントローラーを削除してしまった場合は、いつまでたってもカスタムリソースが削除されないという問題が発生することになるので注意しましょう。

### Finalizerの実装方法

それではFinalizerを実装してみましょう。
controller-runtimeでは、Finalizerを扱うためのユーティリティ関数として[controllerutil.ContainsFinalizer](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#ContainsFinalizer)、[controllerutil.AddFinalizer](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#AddFinalizer)、[controllerutil.RemoveFinalizer](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/controller/controllerutil?tab=doc#RemoveFinalizer)などを提供しているのでこれを利用しましょう。

以下のように、Finalizersフィールドを利用して、独自のリソース削除処理を実装できます。

```go
finalizerName := "markdwonview.finalizers.view.zoetrope.github.io"
if !mdView.ObjectMeta.DeletionTimestamp.IsZero() {
    // deletionTimestampがゼロではないということはリソースの削除が開始されたということ

    // finalizersに上記で指定した名前が存在した場合は削除処理を実施する
    if controllerutil.ContainsFinalizer(&mdView, finalizerName) {
        // ここで外部リソースを削除する
        deleteExternalResources()

        // finalizersフィールドをクリアしてリソースを削除できるようにする
        controllerutil.RemoveFinalizer(&mdView, finalizerName)
        err = r.Update(ctx, &mdView)
        if err != nil {
            return ctrl.Result{}, err
        }
    }
    return ctrl.Result{}, nil
}

// deletionTimestampが付与されていなければ、finalizersフィールドを追加します。
if !controllerutil.ContainsFinalizer(&mdView, finalizerName) {
    controllerutil.AddFinalizer(&mdView, finalizerName)
    err = r.Update(ctx, &mdView)
    if err != nil {
        return ctrl.Result{}, err
    }
}
```

