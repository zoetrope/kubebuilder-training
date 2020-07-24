# コントローラのテスト

controller-runtimeでは[envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest?tab=doc)パッケージを提供しており、コントローラやWebhookの簡易的なテストを実施することが可能です。

envtestはetcdとkube-apiserverを立ち上げてテスト用の環境を構築します。
環境変数`USE_EXISTING_CLUSTER`を指定すれば、既存のKubernetesクラスタを利用したテストをおこなうことも可能です。

なおcontroller-genが生成するテストコードでは、[Ginkgo](https://github.com/onsi/ginkgo)というテストフレームワークを利用しています。
このフレームワークの利用方法については[Ginkgonoのドキュメント](https://onsi.github.io/ginkgo/)を御覧ください。

## テスト環境のセットアップ

controller-genによって自動生成された[controllers/suite_test.go](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/tenant/controllers/suite_test.go)を見てみましょう。

[import, title="controllers/suite_test.go"](../../codes/tenant/controllers/suite_test.go)

まず`envtest.Environment`でテスト用の環境設定をおこないます。
ここでは、`CRDDirectoryPaths`で適用するCRDのマニフェストのパスを指定しています。

`testEnv.Start()`を呼び出すとetcdとkube-apiserverが起動します。
あとはコントローラのメイン関数と同様に初期化処理をおこなうだけです。

テスト終了時にはetcdとkube-apiserverを終了するように`testEnv.Stop()`を呼び出します。

## コントローラのテスト

それでは実際のテストを書いていきましょう。

[import, title="controllers/tenant_controller_test.go"](../../codes/tenant/controllers/tenant_controller_test.go)

このテストではまずテナントのオブジェクトを用意して、`k8sClient`を利用してKubernetesクラスタにリソースの作成をおこなっています。
次にテナントリソースを取得して、ステータスがReadyになったことを確認しています。
最後に指定した通りのnamespaceリソースが作成されていることを確認しています。

## テストの実行

`make test`や`go test`でテストを実行してみましょう。

```console
$ go test -v ./controllers/
=== RUN   TestAPIs
Running Suite: Controller Suite
===============================
Random Seed: 1595578997
Will run 1 of 1 specs

Ran 1 of 1 Specs in 5.511 seconds
SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestAPIs (5.51s)
PASS
ok      github.com/zoetrope/kubebuilder-training/codes/controllers      5.528s
```
