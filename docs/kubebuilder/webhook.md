# Webhookの生成

Kubernetesには、Admission Webhookと呼ばれる拡張機能があります。
これは特定のリソースを作成・更新する際にWebhook APIを呼び出し、バリデーションやリソースの書き換えをおこなうための機能です。

`kubebuilder`コマンドでは、以下の3種類のオプションで生成するWebhookを指定できます。

- `--programmatic-validation`: リソースのバリデーションをおこなうためのWebhook
- `--defaulting`: リソースのフィールドにデフォルト値を設定するためのWebhook
- `--conversion`: カスタムリソースのバージョンアップ時にリソースの変換をおこなうためのWebhook

ここでは`--programmatic-validation`と`--defaulting`を指定して、MarkdownViewリソース用のWebhookを生成してみましょう。

注意: kindにはPodやDeploymentなどの既存のリソースを指定できません。

```console
$ kubebuilder create webhook --group view --version v1 --kind MarkdownView --programmatic-validation --defaulting
$ make manifests
```

以下のファイルが新たに追加されました。

```
├── api
│    └── v1
│        ├── markdownView_webhook.go
│        └── webhook_suite_test.go
└── config
     ├── certmanager
     │   ├── certificate.yaml
     │   ├── kustomization.yaml
     │   └── kustomizeconfig.yaml
     ├── default
     │   ├── manager_webhook_patch.yaml
     │   └── webhookcainjection_patch.yaml
     └── webhook
         ├── kustomization.yaml
         ├── kustomizeconfig.yaml
         ├── manifests.yaml
         └── service.yaml
```

## api/v1

`markdownview_webhook.go`がWebhook実装の雛形になります。
このファイルにWebhookの実装を追加していくことになります。

### config/certmanager

Admission Webhook機能を利用するためには証明書が必要となります。
[cert-manager][]を利用して証明書を発行するためのカスタムリソースが生成されています。

## config/webhook

`config/webhook`下は、Webhook機能を利用するために必要なマニフェストファイルです。
manifests.yamlファイルは`make manifests`ファイルで自動生成されるため、基本的に手動で編集する必要はありません。

## main.go

`main.go`には、以下のようなWebhookの初期化をおこなうためのコードが追加されています。

[import:"init-webhook",unindent="true"](../../codes/00_scaffold/main.go)

## kustomization.yamlの編集

Kubebuilderコマンドで生成した直後の状態では、`make manifests`コマンドでマニフェストを生成しても、Webhook機能が利用できるようにはなっていません。

[config/default/kustomization.yaml](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/markdown-view/config/default/kustomization.yaml)ファイルを編集する必要があります。

生成直後のkustomization.yamlは、`bases` の `../webhook` と `../certmanager`, `patchesStrategicMerge` の `manager_webhook_patch.yaml` と `webhookcainjection_patch.yaml`, `vars` がコメントアウトされていますが、これらのコメントを外します。

[import:"bases,enable-webhook,patches,enable-webhook-patch,vars"](../../codes/00_scaffold/config/default/kustomization.yaml)

[cert-manager]: https://github.com/jetstack/cert-manager
