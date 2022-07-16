# RBACマニフェストの生成

KubernetesではRBAC(Role-based access control)によりリソースへのアクセス権を制御できます。
カスタムコントローラーにおいても、利用するリソースにのみアクセスできるように適切な権限を設定する必要があります。

controller-genでは、Goのソースコード中に埋め込まれたマーカーを元にRBACのマニフェストを生成できます。

まずはKubebuilderによって生成されたマーカーを見てみましょう。

[import:"rbac"](../../codes/00_scaffold/controllers/markdownview_controller.go)

- `groups`: 権限を与えたいリソースのAPIグループを指定します。
- `resources`: 権限を与えたいリソースの種類を指定します。
- `verb`: どのような権限を与えるのかを指定します。コントローラーがおこなう操作に応じた権限を指定します。

MarkdownViewリソースと、そのサブリソースである`status`と`finalizer`に権限が付与されています。
なお、サブリソースはlistやcreate,delete操作をおこなえないので`get;update;patch`の権限のみが付与されています。

これらに加えてMarkdownViewコントローラーが作成するConfigMap, Deployment, Service, Eventリソースを操作する権限のマーカーを追加しましょう。

[import:"rbac"](../../codes/20_manifests/controllers/markdownview_controller.go)

なお、controller-runtimeの提供するClientは、Getでリソースを取得した場合も裏でListやWatchを呼び出しています。
そのためgetしかしない場合でも、get, list, watchを許可しておきましょう。

`make manifests`を実行すると以下のように`config/rbac/role.yaml`が更新されます。

[import](../../codes/20_manifests/config/rbac/role.yaml)
