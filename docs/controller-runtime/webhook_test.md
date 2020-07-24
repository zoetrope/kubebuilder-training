# Webhookのテスト

controller-runtime v0.6.0では、[envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest?tab=doc)でwebhookのテストもサポートするようになりました。

## テスト環境のセットアップ

コントローラ用のテスト環境を用意するコードはcontroller-genが自動生成してくれますが、Webhook用のコードはまだ生成してくれません(controller-gen 0.3.0の場合)。

そこで[api/v1/suite_test.go](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/tenant/api/v1/suite_test.go)のようなコードを自分で用意する必要があります。

[import, title="api/v1/suite_test.go"](../../codes/tenant/api/v1/suite_test.go)

基本的にはコントローラのテストコードと似ていますが、`envtest.Environment`を作成する際に、Webhook用のマニフェストのパスを指定したり、`ctrl.NewManager`を呼び出す際に`Host`,`Port`,`CertDir`のパラメータをtestEnvのパラメータで上書きする必要があります。

## Webhookのテスト

コントローラと同様にWebhookのテストも書いてみましょう。

[import, title="api/v1/tenant_webhook_test.go"](../../codes/tenant/api/v1/tenant_webhook_test.go)

このテストではまずテナントのオブジェクトを用意して、`k8sClient`を利用してKubernetesクラスタにリソースの作成をおこなっています。
その後、defaultingのWebhookにより、`namespacePrefix`の値が正しく設定されていることを確認しています。

## テストの実行

`make test`や`go test`でテストを実行してみましょう。

```
$ go test -v ./api/v1/
=== RUN   TestWebhooks
Running Suite: Webhook Suite
============================
Random Seed: 1595580068
Will run 1 of 1 specs

Ran 1 of 1 Specs in 5.101 seconds
SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped
--- PASS: TestWebhooks (5.10s)
PASS
ok      github.com/zoetrope/kubebuilder-training/codes/api/v1   5.191s
```
