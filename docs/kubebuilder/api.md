# APIの雛形作成

`kubebuilder create api`コマンドを利用すると、カスタムリソースとカスタムコントローラーの実装の雛形を生成できます。

以下のコマンドを実行して、MarkdownViewを表現するためのカスタムリソースと、MarkdownViewを扱うカスタムコントローラーを生成してみましょう。

カスタムリソースとコントローラーのソースコードを生成するかどうか聞かれるので、今回はどちらも`y`と回答します。

```console
$ kubebuilder create api --group view --version v1 --kind MarkdownView
INFO Create Resource [y/n]
y
INFO Create Controller [y/n]
y
```

`--group`,`--version`, `--kind`オプションは、生成するカスタムリソースのGVKを指定します。
- `--kind`: 作成するリソースの名前を指定します。
- `--group`: リソースが属するグループ名を指定します。
- `--version`: 適切なバージョンを指定します。今後仕様が変わる可能性がありそうなら`v1alpha1`や`v1beta1`を指定し、安定版のリソースを作成するのであれば`v1`を指定します。

また、以下のコマンドを実行することでCRDやRBACなどのマニフェストを自動生成することができます。

```console
$ make manifests
```

コマンドの実行に成功すると、新たに下記のファイルが生成されます。

```
.
├── api
│    └── v1
│        ├── groupversion_info.go
│        ├── markdownview_types.go
│        └── zz_generated.deepcopy.go
├── config
│    ├── crd
│    │    ├── bases
│    │    │    └── view.zoetrope.github.io_markdownviews.yaml
│    │    ├── kustomization.yaml
│    │    └── kustomizeconfig.yaml
│    ├── rbac
│    │    ├── markdownview_editor_role.yaml
│    │    └── markdownview_viewer_role.yaml
│    └── samples
│        ├── kustomization.yaml
│        └── view_v1_markdownview.yaml
└── internal
     └── controller
         ├── markdownview_controller.go
         ├── markdownview_controller_test.go
         └── suite_test.go
```

それぞれのファイルの内容をみていきましょう。

## api/v1

`markdownview_types.go`は、MarkdownViewリソースをGo言語のstructで表現したものです。
今後、MarkdownViewリソースの定義を変更する場合にはこのファイルを編集していくことになります。

`groupversion_info.go`は初期生成後に編集する必要はありません。
`zz_generated.deepcopy.go`は`markdownview_types.go`の内容から自動生成されるファイルなので編集する必要はありません。

## cmd/main.go

`cmd/main.go`には、下記のようなコントローラーの初期化処理が追加されています。

[import:"init-reconciler",unindent="true"](../../codes/00_scaffold/cmd/main.go)

## config

configディレクトリ下には、いくつかのファイルが追加されています。

### crd

crdディレクトリにはCRD(Custom Resource Definition)のマニフェストが追加されています。
これらのマニフェストは`api/v1/markdownView_types.go`から自動生成されるものなので、基本的に手動で編集する必要はありません。

### rbac

`role.yaml`には、MarkdownViewリソースを扱うための権限が追加されています。

また、MarkdownViewリソースを扱うためのRoleを定義したマニフェストとして、`markdownview_editor_role.yaml`と`markdownview_viewer_role.yaml`が追加されています。

### samples

カスタムリソースのサンプルマニフェストです。
テストで利用したり、ユーザー向けに提供できるように記述しておきましょう。

## internal/controller

`markdownview_controller.go`は、カスタムコントローラーのメインロジックになります。
今後、カスタムコントローラーの処理は基本的にこのファイルに書いていくことになります。

`suite_test.go`はテストコードです。詳細は[コントローラのテスト](../controller-runtime/controller_test.md)で解説します。

