# Webhookの生成

AdmissionWebhookを利用するためには、`MutatingWebhookConfiguration`や`ValidatingWebhookConfiguration`などののマニフェストを用意する必要があります。

[import:"webhook-defaulter"](../../codes/tenant/api/v1/tenant_webhook.go)

pathは、
failurePlicyは、WebhookのAPIに接続できない場合など呼び出しに失敗したときの挙動を指定します。
verbsには、create, update, delete

[import:"webhook-validator"](../../codes/tenant/api/v1/tenant_webhook.go)
