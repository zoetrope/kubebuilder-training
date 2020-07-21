# Kubebuilder

まずは`kubebuilder`コマンドの利用方法を紹介します。
`kubebuilder`コマンドは、カスタムコントローラのプロジェクトの雛形を自動生成するためのツールです。
ソースコードだけでなく、MakefileやDockerfile、各種マニフェストなど数多くのファイルを生成します。

`kubebuilder`コマンドのヘルプを表示してみましょう。

```console
$ kubebuilder -h
Usage:
  kubebuilder [command]

Available Commands:
  create      Scaffold a Kubernetes API or webhook.
  edit        This command will edit the project configuration
  help        Help about any command
  init        Initialize a new project
  version     Print the kubebuilder version
```

`kubebuilder`には、プロジェクトの新規作成をおこなう`init`サブコマンド、新しいAPIやWebhookの生成をおこなう`create`サブコマンド、生成したプロジェクトの構成を変更する`edit`サブコマンドがあります。

以降では、`init`サブコマンドと`create`サブコマンドの使い方を紹介します。
