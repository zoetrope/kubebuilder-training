# プロジェクトの雛形作成

それではさっそく`kubebuilder init`コマンドを利用して、プロジェクトの雛形を生成しましょう。

```console
$ mkdir markdown-view
$ cd markdown-view
$ kubebuilder init --domain zoetrope.github.io --repo github.com/zoetrope/markdown-view
```

`--domain`で指定した名前は、これから作成するカスタムリソースのグループ名に使われます。
他の人の作ったカスタムリソースと衝突しないように、あなたが保持するドメインなどを利用してユニークな名前を指定してください。

`--repo`にはgo modulesのmodule名を指定します。
GitHubにリポジトリを作る場合は`github.com/<user_name>/<product_name>`を指定します。

コマンドの実行に成功すると、下記のようなファイルが生成されます。

```
.
├── cmd
│    └── main.go
├── config
│    ├── default
│    │    ├── kustomization.yaml
│    │    ├── manager_metrics_patch.yaml
│    │    └── metrics_service.yaml
│    ├── manager
│    │    ├── kustomization.yaml
│    │    └── manager.yaml
│    ├── prometheus
│    │    ├── kustomization.yaml
│    │    └── monitor.yaml
│    └── rbac
│        ├── kustomization.yaml
│        ├── leader_election_role_binding.yaml
│        ├── leader_election_role.yaml
│        ├── metrics_auth_role_binding.yaml
│        ├── metrics_auth_role.yaml
│        ├── metrics_reader_role.yaml
│        ├── role_binding.yaml
│        ├── role.yaml
│        └── service_account.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── hack
│    └── boilerplate.go.txt
├── Makefile
├── PROJECT
├── README.md
└── test
    ├── e2e
    │    ├── e2e_suite_test.go
    │    └── e2e_test.go
    └── utils
        └── utils.go
```

それでは生成されたファイルをそれぞれ見ていきましょう。

## Makefile

コード生成やコントローラーのビルドなどをおこなうためのMakefileです。

`make help`でターゲットの一覧を確認してみましょう。

```console
$ make help

Usage:
  make <target>

General
  help             Display this help.

Development
  manifests        Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
  generate         Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
  fmt              Run go fmt against code.
  vet              Run go vet against code.
  test             Run tests.
  lint             Run golangci-lint linter
  lint-fix         Run golangci-lint linter and perform fixes

Build
  build            Build manager binary.
  run              Run a controller from your host.
  docker-build     Build docker image with the manager.
  docker-push      Push docker image with the manager.
  docker-buildx    Build and push docker image for the manager for cross-platform support
  build-installer  Generate a consolidated YAML with CRDs and deployment.

Deployment
  install          Install CRDs into the K8s cluster specified in ~/.kube/config.
  uninstall        Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
  deploy           Deploy controller to the K8s cluster specified in ~/.kube/config.
  undeploy         Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.

Dependencies
  kustomize        Download kustomize locally if necessary.
  controller-gen   Download controller-gen locally if necessary.
  envtest          Download setup-envtest locally if necessary.
  golangci-lint    Download golangci-lint locally if necessary.
```

## PROJECT

ドメイン名やリポジトリのURLや生成したAPIの情報などが記述されています。
基本的にこのファイルを編集することはあまりないでしょう。

## hack/boilerplate.go.txt

自動生成されるソースコードの先頭に挿入されるボイラープレートです。

デフォルトではApache 2 Licenseの文面が記述されているので、必要に応じて書き換えてください。

## cmd/main.go

これから作成するカスタムコントローラーのエントリーポイントとなるソースコードです。

ソースコード中に`//+kubebuilder:scaffold:imports`, `//+kubebuilder:scaffold:scheme`, `//+kubebuilder:scaffold:builder`などのコメントが記述されています。
Kubebuilderはこれらのコメントを目印にソースコードの自動生成をおこなうので、決して削除しないように注意してください。

## config

configディレクトリ配下には、カスタムコントローラーをKubernetesクラスターにデプロイするためのマニフェストが生成されます。

実装する機能によっては必要のないマニフェストも含まれているので、適切に取捨選択してください。

### default

マニフェストをまとめて利用するための設定が記述されています。

`manager_metrics_patch.yaml`は、カスタムコントローラーのメトリクスを有効にするためのパッチです。

`metrics_service.yaml`は、カスタムコントローラーのメトリクスにアクセスするためのサービス定義です。

利用するマニフェストに応じて、`kustomization.yaml`を編集してください。

### manager

カスタムコントローラーのDeploymentリソースのマニフェストです。
カスタムコントローラーのコマンドラインオプションの変更をおこなった場合など、必要に応じて書き換えてください。

### prometheus

カスタムリソースのメトリクスを収集するための設定を記述したマニフェストです。
Prometheus Operatorを利用している場合、このマニフェストを適用するとPrometheusが自動的にカスタムコントローラーのメトリクスを収集してくれるようになります。

### rbac

各種権限を設定するためのマニフェストです。

`leader_election_role.yaml`と`leader_election_role_binding.yaml`は、リーダーエレクション機能を利用するために必要な権限です。

`metrics_auth_`から始まるファイルは、メトリクスエンドポイントへのアクセスを制限するためのマニフェストです。

`role.yaml`と`role_binding.yaml`は、コントローラーが各種リソースにアクセスするための権限を設定するマニフェストです。
この2つのファイルは自動生成されるものなので、手動で編集しないように注意してください。
`service_account`は、カスタムコントローラーのサービスアカウントを定義するマニフェストで、`role.yaml`で定義した権限が割り当てられます。

## test

E2Eテストをおこなうためのファイルが格納されています。
