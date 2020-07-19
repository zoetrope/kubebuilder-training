# RBACの生成

コントローラが利用するリソースへのアクセス権限を適切に設定する必要があります。
controller-genでは、goのソースコード中に埋め込まれたマーカーを元にRBACのマニフェストを生成することができます。

[import:"rbac"](../../codes/tenant/controllers/tenant_controller.go)

まずは、Tenantリソースのアクセス許可
また、statusをサブリソース化した場合は、個別に権限を追加する必要があります。

前述したように、controller-runtimeの適用しているClientは、Getでリソースを取得した場合にも裏でListやWatchを呼び出しています。
そのためgetしかしない場合でも、get, list, watchを許可しておきましょう。
