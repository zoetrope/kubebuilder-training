# Kubebuilder

`kubebuilder`コマンドは、カスタムコントローラーのプロジェクトの雛形を自動生成するためのツールです。
ソースコードだけでなく、MakefileやDockerfile、各種マニフェストなど数多くのファイルを生成します。

`kubebuilder`コマンドのヘルプを表示してみましょう。

```console
$ kubebuilder -h
CLI tool for building Kubernetes extensions and tools.

Usage:
  kubebuilder [flags]
  kubebuilder [command]

Examples:
The first step is to initialize your project:
    kubebuilder init [--plugins=<PLUGIN KEYS> [--project-version=<PROJECT VERSION>]]

<PLUGIN KEYS> is a comma-separated list of plugin keys from the following table
and <PROJECT VERSION> a supported project version for these plugins.

                             Plugin keys | Supported project versions
-----------------------------------------+----------------------------
               base.go.kubebuilder.io/v4 |                          3
 deploy-image.go.kubebuilder.io/v1-alpha |                          3
                    go.kubebuilder.io/v4 |                          3
         grafana.kubebuilder.io/v1-alpha |                          3
      kustomize.common.kubebuilder.io/v2 |                          3

For more specific help for the init command of a certain plugins and project version
configuration please run:
    kubebuilder init --help --plugins=<PLUGIN KEYS> [--project-version=<PROJECT VERSION>]

Default plugin keys: "go.kubebuilder.io/v4"
Default project version: "3"


Available Commands:
  alpha       Alpha-stage subcommands
  completion  Load completions for the specified shell
  create      Scaffold a Kubernetes API or webhook
  edit        Update the project configuration
  help        Help about any command
  init        Initialize a new project
  version     Print the kubebuilder version

Flags:
  -h, --help                     help for kubebuilder
      --plugins strings          plugin keys to be used for this subcommand execution
      --project-version string   project version (default "3")

Use "kubebuilder [command] --help" for more information about a command.
```

`kubebuilder`には、プロジェクトの新規作成をおこなう`init`サブコマンド、新しいAPIやWebhookの生成をおこなう`create`サブコマンド、生成したプロジェクトの設定を変更する`edit`サブコマンドがあります。

本資料では、`init`サブコマンドと`create`サブコマンドの使い方を紹介します。
