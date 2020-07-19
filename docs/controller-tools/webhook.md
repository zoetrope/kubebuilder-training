# Webhookマニフェストの生成

AdmissionWebhookを利用するためには、`MutatingWebhookConfiguration`や`ValidatingWebhookConfiguration`などのマニフェストを用意する必要があります。

[import:"webhook-defaulter"](../../codes/tenant/api/v1/tenant_webhook.go)

pathは、
failurePolicyは、WebhookのAPIに接続できない場合など呼び出しに失敗したときの挙動を指定します。
verbsには、create, update, delete

[import:"webhook-validator"](../../codes/tenant/api/v1/tenant_webhook.go)
