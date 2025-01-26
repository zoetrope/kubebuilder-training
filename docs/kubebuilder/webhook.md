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
```

APIを作成したときと同様に、以下のコマンドを実行してマニフェストファイルを生成しておきます。

```console
$ make manifests
```

以下のファイルが新たに追加されました。

```
├── api
│    └── v1
│        ├── markdownview_webhook.go
│        ├── markdownview_webhook_test.go
│        └── webhook_suite_test.go
└── config
     ├── certmanager
     │    ├── certificate.yaml
     │    ├── kustomization.yaml
     │    └── kustomizeconfig.yaml
     ├── crd
     │    └── patches
     │        ├── cainjection_in_markdownviews.yaml
     │        └── webhook_in_markdownviews.yaml
     ├── default
     │    ├── manager_webhook_patch.yaml
     │    └── webhookcainjection_patch.yaml
     └── webhook
         ├── kustomization.yaml
         ├── kustomizeconfig.yaml
         ├── manifests.yaml
         └── service.yaml
```

## api/v1

`markdownview_webhook.go`がWebhook実装の雛形になります。
このファイルにWebhookの実装を追加していくことになります。

## config

### certmanager

Admission Webhook機能を利用するためには証明書が必要となります。
[cert-manager][]を利用して証明書を発行するためのカスタムリソースが生成されています。

### crd/patches

このディレクトリにはConversion Webhook用のパッチファイルが格納されています。CRDのバージョンアップをおこなう際に利用します。

`cainjection_in_markdownviews.yaml`は、cert-managerのCA Injection機能を有効にするためのパッチファイルです。
また、`webhook_in_markdownviews.yaml`は、Conversion Webhookを有効にするためのパッチファイルです。

### webhook

`config/webhook`下は、Admission Webhook機能を利用するために必要なマニフェストファイルです。

`webhookcainjectoin_patch.yaml`は、cert-managerのCA Injection機能を有効にするためのパッチファイルです。
また、`manager_webhook_patch.yaml`は、Admission Webhook用の証明書をカスタムコントローラーから参照できるようにするためのパッチファイルです。
`manifests.yaml`ファイルは`make manifests`コマンドで自動生成されるため、手動で編集する必要はありません。

## cmd/main.go

`cmd/main.go`には、以下のようなWebhookの初期化をおこなうためのコードが追加されています。

[import:"init-webhook",unindent="true"](../../codes/00_scaffold/cmd/main.go)

## kustomization.yamlの編集

Kubebuilderコマンドで生成した直後の状態では、Webhook機能が利用できるようにはなっていません。
`config/default/kustomization.yaml`ファイルを編集する必要があります。

`config/default/kustomization.yaml`ファイルを開き、以下のように`resources`の`../certmanager`, `patches`の`webhookcainjection_patch.yaml`, `replacements`のコメントを外して有効化します。

[import](../../codes/00_scaffold/config/default/kustomization.yaml)

[cert-manager]: https://github.com/cert-manager/cert-manager
