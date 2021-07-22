# RBACマニフェストの生成

KubernetesではRBAC(Role-based access control)によりリソースへのアクセス権を制御することができます。
カスタムコントローラにおいても、利用するリソースにのみアクセスできるように適切に権限を設定する必要があります。

controller-genでは、Goのソースコード中に埋め込まれたマーカーを元にRBACのマニフェストを生成することができます。

まずはKubebuilderによって生成されたマーカーを見てみましょう。

```go
//+kubebuilder:rbac:groups=viewer.zoetrope.github.io,resources=markdownviews,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=viewer.zoetrope.github.io,resources=markdownviews/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=viewer.zoetrope.github.io,resources=markdownviews/finalizers,verbs=update
```

- `groups`: 権限を与えたいリソースのAPIグループを指定します。
- `resources`: 権限を与えたいリソースの種類を指定します。
- `verb`: どのような権限を与えるのかを指定します。コントローラがおこなう操作に応じた権限を指定します。

MarkdownViewリソースと、そのサブリソースである`status`と`finalizer`に権限が付与されています。
なお、サブリソースはlistやcreate,delete操作をおこなえないので`get;update;patch`の権限のみが付与されています。

これらに加えて、MarkdownViewコントローラが作成するConfigMap, Deployment, Serviceリソースを操作する権限のマーカーを追加しましょう。

[import:"rbac"](../../codes/markdown-viewer/controllers/markdownview_controller.go)

なお、controller-runtimeの提供するClientは、Getでリソースを取得した場合も裏でListやWatchを呼び出しています。
そのためgetしかしない場合でも、get, list, watchを許可しておきましょう。

`make manifests`を実行すると以下のように`config/rbac/role.yaml`が更新されます。

[import](../../codes/markdown-viewer/config/rbac/role.yaml)
