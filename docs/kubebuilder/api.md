# APIの雛形作成

`kubebuilder create api`コマンドを利用すると、カスタムリソースとカスタムコントローラーの実装の雛形を生成できます。

以下のコマンドを実行して、MarkdownViewを表現するためのカスタムリソースと、MarkdownViewを扱うカスタムコントローラーを生成してみましょう。

カスタムリソースとコントローラーのソースコードを生成するかどうか聞かれるので、今回はどちらも`y`と回答します。

```console
$ kubebuilder create api --group view --version v1 --kind MarkdownView
Create Resource [y/n]
y
Create Controller [y/n]
y
$ make manifests
```

`--group`,`--version`, `--kind`オプションは、生成するカスタムリソースのGVKを指定します。
- `--kind`: 作成するリソースの名前を指定します。
- `--group`: リソースが属するグループ名を指定します。
- `--version`: 適切なバージョンを指定します。今後仕様が変わる可能性がありそうなら`v1alpha1`や`v1beta1`を指定し、安定版のリソースを作成するのであれば`v1`を指定します。

コマンドの実行に成功すると、下記のようなファイルが新たに生成されます。

```
├── api
│    └── v1
│        ├── groupversion_info.go
│        ├── markdownview_types.go
│        └── zz_generated.deepcopy.go
├── config
│    ├── crd
│    │    ├── bases
│    │    │   └── view.zoetrope.github.io_markdownviews.yaml
│    │    ├── kustomization.yaml
│    │    ├── kustomizeconfig.yaml
│    │    └── patches
│    │        ├── cainjection_in_markdownviews.yaml
│    │        └── webhook_in_markdownviews.yaml
│    ├── rbac
│    │    ├── role.yaml
│    │    ├── markdownview_editor_role.yaml
│    │    └── markdownview_viewer_role.yaml
│    └── samples
│        └── view_v1_markdownview.yaml
└── controllers
     ├── markdownview_controller.go
     └── suite_test.go
```

それぞれのファイルの内容をみていきましょう。

## api/v1

`markdownview_types.go`は、MarkdownViewリソースをGo言語のstructで表現したものです。
今後、MarkdownViewリソースの定義を変更する場合にはこのファイルを編集していくことになります。

`groupversion_info.go`は初期生成後に編集する必要はありません。
`zz_generated.deepcopy.go`は`markdownview_types.go`の内容から自動生成されるファイルなので編集する必要はありません。

## controllers

`markdownview_controller.go`は、カスタムコントローラーのメインロジックになります。
今後、カスタムコントローラーの処理は基本的にこのファイルに書いていくことになります。

`suite_test.go`はテストコードです。詳細は[コントローラのテスト](../controller-runtime/controller_test.md)で解説します。

## main.go

`main.go`には、下記のようなコントローラーの初期化処理が追加されています。

```go
if err = (&controllers.MarkdownViewReconciler{
	Client: mgr.GetClient(),
	Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
	setupLog.Error(err, "unable to create controller", "controller", "MarkdownView")
	os.Exit(1)
}
```

## config

configディレクトリ下には、いくつかのファイルが追加されています。

### crd

crdディレクトリにはCRD(Custom Resource Definition)のマニフェストが追加されています。

これらのマニフェストは`api/v1/markdownView_types.go`から自動生成されるものなので、基本的に手動で編集する必要はありません。
ただし、Conversion Webhookを利用したい場合は、`cainjection_in_markdownViews.yaml`と`webhook_in_markdownViews.yaml`のパッチを利用するように`kustomization.yaml`を書き換えてください。

### rbac

`role.yaml`には、MarkdownViewリソースを扱うための権限が追加されています。

`markdownview_editor_role.yaml`と`markdownview_viewer_role.yaml`は、MarkdownViewリソースの編集・読み取りの権限です。
必要に応じて利用しましょう。

### samples

カスタムリソースのサンプルマニフェストです。
テストで利用したり、ユーザー向けに提供できるように記述しておきましょう。
