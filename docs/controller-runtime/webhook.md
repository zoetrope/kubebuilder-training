# Webhookの実装

[Webhookマニフェストの生成](../controller-tools/webhook.md)で解説したように、テナントコントローラではリソースの作成時にデフォルト値を設定するためのWebhookと、リソースの更新時にバリデーションするためのWebhookを作成します。

これらのWebhookの実装は非常に簡単で、controller-genで生成された関数に必要な処理を書いていくだけです。

## デフォルト値設定のWebhook

まずはデフォルト値を設定するWebhookの実装です。

[import:"default"](../../codes/tenant/api/v1/tenant_webhook.go)

`namespacePrefix`フィールドが空だった場合は、テナントの名前に`-`を連結した文字列を`namespacePrefix`として利用します。

## バリデーションのWebhook

次にバリデーションWebhookの実装です。

[import:"validate"](../../codes/tenant/api/v1/tenant_webhook.go)

今回はCreateとDelete時のバリデーションをおこなわないため、`ValidateUpdate`関数のみを実装します。
更新前のリソースが引数で渡ってくるので、`namespacePrefix`フィールドが変更されていればバリデーションエラーとします。

このようなバリデーションを実装することで、途中で`namespacePrefix`を変更できないようにすることが可能です。

## 動作確認

Webhookの動作確認をしてみましょう。

Webhookの実装をおこなったカスタムコントローラをKubernetesクラスタにデプロイし、下記のような`namespacePrefix`を指定していないマニフェストを適用します。

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  name: sample
spec:
  namespaces:
    - test1
  admin:
    kind: ServiceAccount
    name: default
    namespace: default
```

作成されたリソースを確認して、`namespacePrefix`に"sample-"という文字列が入っていれば成功です。

```
$ kubectl get tenant sample
NAME     ADMIN     PREFIX    READY
sample   default   sample-   True
```

続いてバリデーションWebhookの動作も確認してみましょう。

先ほど作成したリソースをeditして`namespacePrefix`を別の名前に変更しようとしたときに、下記のようなエラーが発生すれば成功です。

```
$ kubectl edit tenant sample
error: tenants.multitenancy.example.com "sample" could not be patched: admission webhook "vtenant.kb.io" denied the request: spec.namespacePrefix field should not be changed
You can run `kubectl replace -f /tmp/kubectl-edit-bwuei.yaml` to try this update again.
```
