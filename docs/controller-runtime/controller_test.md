# コントローラのテスト

controller-runtimeは[envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest?tab=doc)というパッケージを提供しており、コントローラやWebhookの簡易的なテストを実施することが可能です。

envtestはetcdとkube-apiserverを立ち上げてテスト用の環境を構築します。
また環境変数`USE_EXISTING_CLUSTER`を指定すれば、既存のKubernetesクラスタを利用したテストをおこなうことも可能です。

envtestでは、etcdとkube-apiserverのみを立ち上げており、controller-managerやschedulerは動いていません。
そのため、DeploymentやCronJobリソースを作成しても、Podは作成されないので注意してください。

なおcontroller-genが生成するテストコードでは、[Ginkgo](https://github.com/onsi/ginkgo)というテストフレームワークを利用しています。
このフレームワークの利用方法については[Ginkgoのドキュメント](https://onsi.github.io/ginkgo/)を御覧ください。

## Envtest Binaries Manager

controller-runtimeは、[Envtest Binaries Manager](https://github.com/kubernetes-sigs/controller-runtime/tree/master/tools/setup-envtest)
というツールを提供しています。
このツールを利用することで、Envtestで利用するetcdやkube-apiserverの任意のバージョンのバイナリをセットアップすることができます。

Kubebuilder v3.1時点では、Envtest Binaries Managerが利用されるようになっていないので、Makefileを書き換えておきましょう。

まず、Envtest Binaries Managerをインストールするためのターゲットを追加します。

[import:"setup-envtest"](../../codes/markdown-viewer/Makefile)

testターゲットは以下のように書き換えます。

[import:"test"](../../codes/markdown-viewer/Makefile)

## テスト環境のセットアップ

controller-genによって自動生成された[controllers/suite_test.go](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/markdown-viewer/controllers/suite_test.go)を見てみましょう。

[import, title="controllers/suite_test.go"](../../codes/markdown-viewer/controllers/suite_test.go)

まず`envtest.Environment`でテスト用の環境設定をおこないます。
ここでは、`CRDDirectoryPaths`で適用するCRDのマニフェストのパスを指定しています。

`testEnv.Start()`を呼び出すとetcdとkube-apiserverが起動します。
あとはコントローラのメイン関数と同様に初期化処理をおこなうだけです。

テスト終了時にはetcdとkube-apiserverを終了するように`testEnv.Stop()`を呼び出します。

## コントローラのテスト

それでは実際のテストを書いていきましょう。

まずは各テストの実行前と実行後に呼び出される`BeforeEach`と`AfterEach`関数を実装します。

[import:"setup",unindent:"true"](../../codes/markdown-viewer/controllers/markdownview_controller_test.go)

`BeforeEach`では、テストで利用したリソースをすべて削除します。 (なお、Serviceリソースは`DeleteAllOf`をサポートしていないため、1つずつ削除しています。)
その後、MarkdownViewReconcilerを作成し、Reconciliation Loop処理を開始します。

`AfterEach`では、`BeforeEach`で起動したReconciliation Loop処理を停止します。

次にテストケースを記述します。

[import:"test",unindent:"true"](../../codes/markdown-viewer/controllers/markdownview_controller_test.go)

これらのテストケースでは`k8sClient`を利用してKubernetesクラスタにMarkdownViewリソースを作成し、
その後に期待するリソースが作成されていることを確認しています。
Reconcile処理はテストコードとは非同期に動くため、Eventually関数を利用してリソースが作成できるまで待つようにしています。

最後のテストではStatusが更新されることを確認しています。
本来はここでStatusがHealthyになることをテストすべきですが、Envtestではcontroller-managerが存在しないためDeploymentがReadyにならず、
MarkdownViewのStatusもHealthyになりません。
Envtestは実際のKubernetesクラスターとは異なるということを意識してテストを書くようにしましょう。

テストが書けたら、`make test`でテストを実行してみましょう。
