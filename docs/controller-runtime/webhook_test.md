# Webhookのテスト(Envtest)

ここではEnvtestを利用したWebhookのテストの書き方を学びます。

## テスト環境のセットアップ

Webhookもコントローラーのテストと同じくEnvtestを利用できます。
Kubebuilderによってテストを実行するためのコードが以下のように生成されています。

[import, title="api/v1/webhook_suite_test.go"](../../codes/40_reconcile/api/v1/webhook_suite_test.go)

基本的にはコントローラーのテストコードと似ていますが、`envtest.Environment`を作成する際に、Webhook用のマニフェストのパスを指定したり、
`ctrl.NewManager`を呼び出す際に`Host`,`Port`,`CertDir`のパラメータをtestEnvのパラメータで上書きする必要があります。

## Webhookのテスト

Webhookのテストコードを書いてみましょう。

[import, title="api/v1/markdownview_webhook_test.go"](../../codes/40_reconcile/api/v1/markdownview_webhook_test.go)

MutatingWebhookのテストでは、入力となるマニフェストファイル(input.yaml)を利用してリソースを作成し、
作成されたリソースが期待値となるマニフェストファイル(output.yaml)の内容と一致することを確認しています。

ValidatingWebhookのテストでは、Validなマニフェストファイル(valid.yaml)を利用してリソースが作成できることと、
Invalidなマニフェストファイル(empty-markdowns.yaml, invalid-replicas.yaml, without-summary.yaml)を利用してリソースの作成に失敗することをテストしています。

最後に、`make test`でテストに通ることを確認しましょう。

```console
$ make test
KUBEBUILDER_ASSETS="~/markdown-view/bin/k8s/1.30.0-linux-amd64" go test $(go list ./... | grep -v /e2e) -coverprofile cover.out
        github.com/zoetrope/markdown-view/cmd           coverage: 0.0% of statements
        github.com/zoetrope/markdown-view/test/utils            coverage: 0.0% of statements
ok      github.com/zoetrope/markdown-view/api/v1        5.573s  coverage: 51.6% of statements
ok      github.com/zoetrope/markdown-view/internal/controller   6.136s  coverage: 69.6% of statements
```
