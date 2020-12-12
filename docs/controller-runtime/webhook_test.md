# Webhookのテスト

## テスト環境のセットアップ

Kubebuilder v3から、WebHookのテストをセットアップするコードが生成されるようになりました。
[api/v1/webhook_suite_test.go](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/tenant/api/v1/webhook_suite_test.go)を見てみましょう。

[import, title="api/v1/webhook_suite_test.go"](../../codes/tenant/api/v1/webhook_suite_test.go)

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
