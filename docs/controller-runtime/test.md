# テスト

Kubebuilderが生成したプロジェクトには、以下のようにいくつかのテストコードが含まれています。

```console
.
├── api
│    └── v1
│        ├── markdownview_webhook_test.go       # Envtestを利用したWebhookのテストコード
│        └── webhook_suite_test.go              # Envtestをセットアップするためのコード
├── internal
│    └── controller
│        ├── markdownview_controller_test.go    # Envtestを利用したカスタムコントローラーのテストコード
│        └── suite_test.go                      # Envtestをセットアップするためのコード
└── test
     ├── e2e
     │    ├── e2e_suite_test.go                 # E2Eテストをセットアップするためのコード
     │    └── e2e_test.go                       # E2Eテストコード
     └── utils
         └── utils.go                           # E2Eテストで利用するユーティリティ関数
```

1つは、Envtestと呼ばれるツールを利用したカスタムコントローラーやWebhookの簡易的なテストです。
もう1つは、本物のKubernetesクラスターを利用したE2E(End-to-End)テストです。

なおEnvtestとE2Eテストのコードでは、[Ginkgo](https://onsi.github.io/ginkgo/)/[Gomega](https://onsi.github.io/gomega/)というテストフレームワークを利用しています。
以下のBookにGinkgo/Gomegaを利用したカスタムコントローラーやWebhookのテストの記述方法をまとめていますので、参考にしてください。

- [Ginkgo/GomegaによるKubernetes Operatorのテスト手法](https://zenn.dev/zoetro/books/testing-kubernetes-operator)

## Envtest

[Envtest](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest?tab=doc)はetcdとkube-apiserverのみを立ち上げてテスト用の環境を構築し、カスタムコントローラーやWebhookの簡易的なテストを実行するためのツールです。

Envtestでは、etcdとkube-apiserverのみを立ち上げており、controller-managerやschedulerは動いていません。
そのため、DeploymentやCronJobリソースを作成しても、Podは作成されないので注意してください。
一方、Kubernetesクラスターを立ち上げる必要がないため、高速にテスト環境を構築できるという利点があります。

さらに、controller-runtimeは[Envtest Binaries Manager](https://github.com/kubernetes-sigs/controller-runtime/tree/master/tools/setup-envtest)というツールを提供しています。
このツールを利用することで、Envtestで利用するetcdやkube-apiserverの任意のバージョンのバイナリを簡単にセットアップできます。

`make test`コマンドにより、Envtest環境をセットアップし、テストを実行することができます。

カスタムコントローラーに関するテストの記述方法は[コントローラーのテスト(Envtest)](controller_test.md)、Webhookに関するテストの記述方法は[Webhookのテスト(Envtest)](webhook_test.md)を参照してください。

## E2Eテスト

E2E(End-to-End)テストは、本物のKubernetesクラスターを利用してテストを実行します。
Envtestとは異なり、テストを実行するためのKubernetesクラスターを立ち上げる必要があるため、テスト環境の構築には時間がかかります。
一方で、本物のKubernetesクラスターを利用するため、実際の運用環境に近い状況でテストを実行できるという利点があります。

E2Eテストを実行するためには、まず以下のコマンドを実行してKubernetesクラスターを立ち上げます。

```console
$ kind create cluster
```

次に、以下のコマンドを実行してE2Eテストを実行します。

```console
$ make test-e2e
```
