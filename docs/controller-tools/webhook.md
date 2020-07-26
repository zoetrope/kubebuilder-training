# Webhookマニフェストの生成

AdmissionWebhookを利用するためには、`MutatingWebhookConfiguration`や`ValidatingWebhookConfiguration`などのマニフェストを用意する必要があります。
controller-genでは`// +kubebuilder:webhook`マーカーの記述に基づいてマニフェストを生成することができます。

まずはデフォルト値を設定するWebhookのマーカーを見てみましょう。

[import:"webhook-defaulter"](../../codes/tenant/api/v1/tenant_webhook.go)

`groups`,`versions`,`resource`には、Webhookの対象となるリソースのGVKを指定します。
`path`はWebhookのパスを指定しますが、これはcontroller-runtimeが自動的に生成するパスなので基本的には変更せずに利用します。
`mutating`にはMutatingWebhookかどうかを指定します。
`failurePolicy`は、WebhookのAPIに接続できない場合など呼び出しに失敗したときの挙動を指定します。
`verbs`はリソースに対してどの操作をおこなったときにWebhookを呼び出すかを指定できます。

今回はテナントリソースが作成されたときだけデフォルト値を設定するように、`verbs`をcreateのみに変更しました。

次にバリデーションWebhookのマーカーを見てみましょう。

[import:"webhook-validator"](../../codes/tenant/api/v1/tenant_webhook.go)

今回はテナントリソースが更新されたときだけバリデーションをおこなうように、`verbs`をupdateのみに変更しました。

`make manifests`を実行すると[config/webhook/manifests.yaml](../../codes/tenant/config/webhook/manifests.yaml)が更新されます。
