# ガベージコレクション

https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/


[import:"namespace,controller-reference"](../../codes/tenant/controllers/tenant_controller.go)

外部リソースや、ownerReferenceを付与できない、つまりオーナーにはなれないリソース
TopoLVMでは、LVとか

そのようなリソースの削除には、Finalizerという仕組みを利用することができます。

[import:"finalizer"](../../codes/tenant/controllers/tenant_controller.go)

Finalizerフィールドが付与されたリソースは、リソースの削除時に即座に削除されることはなく、
まず`ObjectMeta.DeletionTimestamp`が付与されます。

そこで、`ObjectMeta.DeletionTimestamp`が付与されている場合に、外部リソースの削除をおこない、
それからFinalizerを削除します。

controllerutilにはFinalizerを扱うためのユーティリティ関数が用意されています。

この処理は定石。
