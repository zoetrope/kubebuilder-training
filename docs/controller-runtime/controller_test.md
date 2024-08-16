# コントローラーのテスト(Envtest)

ここではEnvtestを利用したカスタムコントローラーのテストの書き方を学びます。

## テスト環境のセットアップ

まず、自動生成された`internal/controller/suite_test.go`を見てみましょう。

[import, title="internal/controller/suite_test.go"](../../codes/40_reconcile/internal/controller/suite_test.go)

最初に`envtest.Environment`でテスト用の環境設定をおこなっています。
`CRDDirectoryPaths`で適用するCRDのマニフェストのパスを指定し、BinaryAssetsDirectoryでEnvtestのバイナリファイルのディレクトリを指定しています。

`testEnv.Start()`を呼び出すとetcdとkube-apiserverが起動します。
あとはカスタムコントローラーのmain関数と同様にscheme初期化処理をおこない、最後にEnvtestのapiserverに接続するためのクライアントを作成しています。

テスト終了時にはetcdとkube-apiserverを終了するように、AfterSuiteで`testEnv.Stop()`を呼び出します。

## コントローラーのテスト

それでは実際のテストを書いていきましょう。

[import](../../codes/40_reconcile/internal/controller/markdownview_controller_test.go)

まずは各テストの実行前と実行後に呼び出される`BeforeEach`と`AfterEach`を実装します。
`BeforeEach`では、テスト用のNamespaceとMarkdownViewリソースを作成しています。
`AfterEach`では、テストで利用したリソースを削除していmます。 (なお、Serviceリソースは`DeleteAllOf`をサポートしていないため、1つずつ削除しています。)

次に`It`を利用してテストケースを記述します。
このテストケースでは`k8sClient`を利用してKubernetesクラスターにMarkdownViewリソースを作成し、その後に期待するリソースが作成されていることを確認しています。

最後のテストではStatusが更新されることを確認しています。
本来はここでStatusがHealthyになることをテストすべきでしょう。
しかし、Envtestではcontroller-managerが存在しないためDeploymentがReadyにならず、MarkdownViewのStatusもHealthyになることはありません。
よってここではStatusが何かしら更新されればOKというテストにしています。
Envtestは実際のKubernetesクラスターとは異なるということを意識してテストを書くようにしましょう。

テストが書けたら、`make test`で実行してみましょう。
テストに成功すると以下のようにokと表示されます。

```console
$ make test
KUBEBUILDER_ASSETS="~/markdown-view/bin/k8s/1.30.0-linux-amd64" go test $(go list ./... | grep -v /e2e) -coverprofile cover.out
        github.com/zoetrope/markdown-view/cmd           coverage: 0.0% of statements
        github.com/zoetrope/markdown-view/test/utils            coverage: 0.0% of statements
ok      github.com/zoetrope/markdown-view/api/v1        5.573s  coverage: 51.6% of statements
ok      github.com/zoetrope/markdown-view/internal/controller   6.136s  coverage: 69.6% of statements
```
