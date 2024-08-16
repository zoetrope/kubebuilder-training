# Webhookマニフェストの生成

AdmissionWebhookを利用するためには、`MutatingWebhookConfiguration`や`ValidatingWebhookConfiguration`などのマニフェストを用意する必要があります。
controller-genは`// +kubebuilder:webhook`マーカーの記述に基づいてマニフェストを生成できます。

まずはMutating Webhookのマーカーを見てみましょう。

[import:"webhook-defaulter"](../../codes/20_manifests/api/v1/markdownview_webhook.go)

同様にValidating Webhookのマーカーを確認します。

[import:"webhook-validator"](../../codes/20_manifests/api/v1/markdownview_webhook.go)

- `path`: Webhookのパスを指定します。これはcontroller-runtimeが自動的に生成するパスなので基本的には変更せずに利用します。
- `mutating`: Webhookで値を書き換えるかどうかを指定します。Defaulterでは`true`, Validatorでは`false`を指定します。
- `failurePolicy`: Webhook APIの呼び出しに失敗したときの挙動を指定します。`fail`を指定するとWebhookが呼び出せない場合はリソースの作成もできません。`ignore`を指定するとWebhookが呼び出せなくてもリソースが作成できてしまいます。
- `sideEffects`: Webhook APIの呼び出しに副作用があるかどうかを指定します。これはAPIサーバーをdry-runモードで呼び出したときの挙動に影響します。副作用がない場合は`None`, ある場合は`Some`を指定します。
- `groups`,`versions`,`resource`: Webhookの対象となるリソースのGVKを指定します。
- `verbs`: Webhookの対象となるリソースの操作を指定できます。`create`, `update`, `delete`などを指定できます。
- `name`: Webhookの名前を指定します。ドットで区切られた3つ以上のセグメントを持つドメイン名でなければなりません。
- `admissionReviewVersions`: WebhookがサポートするAdmissionReviewのバージョンを指定します。Kubernetes 1.16以降の環境でしか動作させないのであれば`v1`のみで問題ありません。1.15以前の環境で動作させたい場合は`v1beta1`も指定しましょう。

`make manifests`を実行すると、マーカーの内容に基づいて以下のようなマニフェストファイルが生成されます。

[import](../../codes/20_manifests/config/webhook/manifests.yaml)
