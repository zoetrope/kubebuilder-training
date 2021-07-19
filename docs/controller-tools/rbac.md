# RBACマニフェストの生成

KubernetesではRBAC(Role-based access control)によりリソースへのアクセス権を制御することができます。
カスタムコントローラにおいても、利用するリソースにのみアクセスできるように適切に権限を設定する必要があります。

controller-genでは、Goのソースコード中に埋め込まれたマーカーを元にRBACのマニフェストを生成することができます。

MarkdownViewerコントローラに付与したマーカーを見てみましょう。

[import:"rbac"](../../codes/markdown-viewer/controllers/markdownview_controller.go)

まずは、MarkdownViewリソースに対して`get;list;watch;create;update;patch;delete`の権限を与えます。
サブリソースである`status`と`finalizer`にも、個別に権限を追加する必要があります。
サブリソースはcreateやdelete操作をおこなえないので`get;update;patch`の権限を与えます。

また、MarkdownViewコントローラが管理するConfigMap, Deployment, Serviceを操作する権限も追加します。

なお、controller-runtimeの提供するClientは、Getでリソースを取得した場合も裏でListやWatchを呼び出しています。
そのためgetしかしない場合でも、get, list, watchを許可しておきましょう。

`make manifests`を実行すると以下のように`config/rbac/role.yaml`が更新されます。

[import](../../codes/markdown-viewer/config/rbac/role.yaml)
