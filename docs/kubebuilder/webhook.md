# Webhookの生成

Kubernetesには、Admission Webhookと呼ばれる拡張機能があります。
これは特定のリソースを作成・更新する際にWebhook APIを呼び出し、バリデーションやリソースの書き換えをおこなうための機能です。

`kubebuilder`コマンドでは、以下の3種類のオプションで生成するWebhookを指定することができます。

- `--programmatic-validation`: リソースのバリデーションをおこなうためのWebhook
- `--defaulting`: リソースのフィールドにデフォルト値を設定するためのWebhook
- `--conversion`: カスタムリソースのバージョンアップ時にリソースの変換をおこなうためのWebhook

ここでは`--programmatic-validation`と`--defaulting`を指定して、Tenantリソース用のWebhookを生成してみましょう。

```console
$ kubebuilder create webhook --group multitenancy --version v1 --kind Tenant --programmatic-validation --defaulting
$ make manifests
```

以下のファイルが新たに追加されました。

```
├── api
│    └── v1
│        ├── tenant_webhook.go
│        └── webhook_suite_test.go
├── config
│    ├── certmanager
│    │   ├── certificate.yaml
│    │   ├── kustomization.yaml
│    │   └── kus
│    ├── crd
│    │   └── patches
│    │       ├── cainjection_in_tenants.yaml
│    │       └── webhook_in_tenants.yaml
│    ├── default
│    │   ├── manager_webhook_patch.yaml
│    │   └── webhookcainjection_patch.yaml
│    └── webhook
│        ├── kustomization.yaml
│        ├── kustomizeconfig.yaml
│        ├── manifests.yaml
│        └── service.yaml
├── default
│    ├── manager_webhook_patch.yaml
│    └── webhookcainjection_patch.yaml
└── main.go
```

## api/v1

`tenant_webhook.go`がWebhook実装の雛形になります。
このファイルにWebhookの実装を追加していくことになります。

### config/certmanager

Admission Webhook機能を提供するためには証明書が必要となります。
certmanagerディレクトリ下のマニフェストを適用すると、[cert-manager][]を利用して証明書を発行することができます。

## config/webhook

`config/webhook`下のファイルは、Webhook機能を利用するために必要なマニフェストになります。
基本的に編集する必要はありません。

## main.go

`main.go`には、以下のようなWebhookの初期化をおこなうためのコードが追加されています。

```go
if err = (&multitenancyv1.Tenant{}).SetupWebhookWithManager(mgr); err != nil {
	setupLog.Error(err, "unable to create webhook", "webhook", "Tenant")
	os.Exit(1)
}
```

## kustomization.yamlの編集

kubebuilderコマンドで生成した直後の状態では、`make manifests`コマンドでマニフェストを生成しても、Webhook機能が利用できるようにはなっていません。

[config/default/kustomization.yaml](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/tenant/config/default/kustomization.yaml)ファイルを編集する必要があります。

生成直後は`bases`の`../webhook`と`../certmanager`、`patchesStrategicMerge`の`manager_webhook_patch.yaml`と`webhookcainjection_patch.yaml`、`vars`がコメントアウトされていますが、これらのコメントを外します。

[import:"bases,enable-webhook,patches,enable-webhook-patch,vars"](../../codes/tenant/config/default/kustomization.yaml)

[cert-manager]: https://github.com/jetstack/cert-manager
