# Kubebuilder

`kubebuilder`コマンドは、カスタムコントローラのプロジェクトの雛形を自動生成するためのツールです。
ソースコードだけでなく、MakefileやDockerfile、各種マニフェストなど数多くのファイルを生成します。

`kubebuilder`コマンドのヘルプを表示してみましょう。

```console
$ kubebuilder -h

(中略)

Available Commands:
  completion  Load completions for the specified shell
  create      Scaffold a Kubernetes API or webhook
  edit        This command will edit the project configuration
  help        Help about any command
  init        Initialize a new project
  version     Print the kubebuilder version

Flags:
  -h, --help   help for kubebuilder

Use "kubebuilder [command] --help" for more information about a command.
```

`kubebuilder`には、プロジェクトの新規作成をおこなう`init`サブコマンド、新しいAPIやWebhookの生成をおこなう`create`サブコマンド、生成したプロジェクトの設定を変更する`edit`サブコマンドがあります。

本資料では、`init`サブコマンドと`create`サブコマンドの使い方を紹介します。
