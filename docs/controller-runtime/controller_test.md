# コントローラーのテスト

controller-runtimeは[envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest?tab=doc)というパッケージを提供しており、
コントローラーやWebhookの簡易的なテストを実施できます。

envtestはetcdとkube-apiserverを立ち上げてテスト用の環境を構築します。
また環境変数`USE_EXISTING_CLUSTER`を指定すれば、既存のKubernetesクラスターを利用したテストをおこなうことも可能です。

Envtestでは、etcdとkube-apiserverのみを立ち上げており、controller-managerやschedulerは動いていません。
そのため、DeploymentやCronJobリソースを作成しても、Podは作成されないので注意してください。

controller-runtimeは、[Envtest Binaries Manager](https://github.com/kubernetes-sigs/controller-runtime/tree/master/tools/setup-envtest)
というツールを提供しています。
このツールを利用することで、Envtestで利用するetcdやkube-apiserverの任意のバージョンのバイナリをセットアップできます。

なおcontroller-genが生成するテストコードでは、[Ginkgo](https://github.com/onsi/ginkgo)というテストフレームワークを利用しています。
このフレームワークの利用方法については[Ginkgoのドキュメント](https://onsi.github.io/ginkgo/)を御覧ください。

## テスト環境のセットアップ

controller-genによって自動生成された`controllers/suite_test.go`を見てみましょう。

[import, title="controllers/suite_test.go"](../../codes/40_reconcile/controllers/suite_test.go)

まず`envtest.Environment`でテスト用の環境設定をおこないます。
ここでは、`CRDDirectoryPaths`で適用するCRDのマニフェストのパスを指定しています。

`testEnv.Start()`を呼び出すとetcdとkube-apiserverが起動します。
あとはコントローラーのメイン関数と同様に初期化処理をおこなうだけです。

テスト終了時にはetcdとkube-apiserverを終了するように`testEnv.Stop()`を呼び出します。

## コントローラーのテスト

それでは実際のテストを書いていきましょう。

[import](../../codes/40_reconcile/controllers/markdownview_controller_test.go)

まずは各テストの実行前と実行後に呼び出される`BeforeEach`と`AfterEach`を実装します。

`BeforeEach`では、テストで利用したリソースをすべて削除します。 (なお、Serviceリソースは`DeleteAllOf`をサポートしていないため、1つずつ削除しています。)
その後、MarkdownViewReconcilerを作成し、Reconciliation Loop処理を開始します。

`AfterEach`では、`BeforeEach`で起動したReconciliation Loop処理を停止します。

次に`It`を利用してテストケースを記述します。

これらのテストケースでは`k8sClient`を利用してKubernetesクラスターにMarkdownViewリソースを作成し、
その後に期待するリソースが作成されていることを確認しています。
Reconcile処理はテストコードとは非同期に動くため、Eventually関数を利用してリソースが作成できるまで待つようにしています。

なお、`newMarkdownView`はテスト用のMarkdownViewリソースを作成するための補助関数です。

最後のテストではStatusが更新されることを確認しています。
本来はここでStatusがHealthyになることをテストすべきでしょう。
しかし、Envtestではcontroller-managerが存在しないためDeploymentがReadyにならず、MarkdownViewのStatusもHealthyになることはありません。
よってここではStatusが何かしら更新されればOKというテストにしています。
Envtestは実際のKubernetesクラスターとは異なるということを意識してテストを書くようにしましょう。

テストが書けたら、`make test`で実行してみましょう。
テストに成功すると以下のようにokと表示されます。

```console
?       github.com/zoetrope/markdown-view       [no test files]
ok      github.com/zoetrope/markdown-view/api/v1        6.957s  coverage: 51.6% of statements
ok      github.com/zoetrope/markdown-view/controllers   8.319s  coverage: 85.3% of statements
```
