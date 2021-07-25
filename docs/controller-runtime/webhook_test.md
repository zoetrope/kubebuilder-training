# Webhookのテスト

## テスト環境のセットアップ

Webhookもコントローラーのテストと同じくEnvtestを利用できます。
Kubebuilderによってテストを実行するためのコードが以下のように生成されています。

[import, title="api/v1/webhook_suite_test.go"](../../codes/markdown-view/api/v1/webhook_suite_test.go)

基本的にはコントローラーのテストコードと似ていますが、`envtest.Environment`を作成する際に、Webhook用のマニフェストのパスを指定したり、
`ctrl.NewManager`を呼び出す際に`Host`,`Port`,`CertDir`のパラメータをtestEnvのパラメータで上書きする必要があります。

## Webhookのテスト

Webhookのテストコードを書いてみましょう。

[import, title="api/v1/markdownview_webhook_test.go"](../../codes/markdown-view/api/v1/markdownview_webhook_test.go)

MutatingWebhookのテストでは、入力となるマニフェストファイル(before.yaml)を利用してリソースを作成し、
作成されたリソースが期待値となるマニフェストファイル(after.yaml)の内容と一致することを確認しています。

ValidatingWebhookのテストでは、Validなマニフェストファイル(valid.yaml)を利用してリソースが作成できることと、
Invalidなマニフェストファイル(empty-markdowns.yaml, invalid-replicas.yaml, without-summary.yaml)を利用してリソースの作成に失敗することをテストしています。

最後に、`make test`でテストに通ることを確認しましょう。

