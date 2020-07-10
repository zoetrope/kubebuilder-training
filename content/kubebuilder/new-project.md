---
title: "Create a new project"
draft: true
weight: 12
---

それではさっそく`kubebuilder`コマンドを利用して、プロジェクトの雛形を生成しましょう。

```console
$ mkdir tenant
$ cd tenant
$ kubebuilder init --repo example.com/tenant --domain example.com
```

`--domain`で指定した名前はCRDのグループ名に使われます。
あなたの所属する組織が保持するドメインなどを利用して、ユニークでvalidな名前を指定してください。

コマンドの実行に成功すると、下記のようなファイルが生成されます。

```
├── Dockerfile
├── Makefile
├── PROJECT
├── bin
│   └── manager
├── config
│   ├── certmanager
│   │   ├── certificate.yaml
│   │   ├── kustomization.yaml
│   │   └── kustomizeconfig.yaml
│   ├── default
│   │   ├── kustomization.yaml
│   │   ├── manager_auth_proxy_patch.yaml
│   │   ├── manager_webhook_patch.yaml
│   │   └── webhookcainjection_patch.yaml
│   ├── manager
│   │   ├── kustomization.yaml
│   │   └── manager.yaml
│   ├── prometheus
│   │   ├── kustomization.yaml
│   │   └── monitor.yaml
│   ├── rbac
│   │   ├── auth_proxy_client_clusterrole.yaml
│   │   ├── auth_proxy_role.yaml
│   │   ├── auth_proxy_role_binding.yaml
│   │   ├── auth_proxy_service.yaml
│   │   ├── kustomization.yaml
│   │   ├── leader_election_role.yaml
│   │   ├── leader_election_role_binding.yaml
│   │   └── role_binding.yaml
│   └── webhook
│       ├── kustomization.yaml
│       ├── kustomizeconfig.yaml
│       └── service.yaml
├── go.mod
├── go.sum
├── hack
│   └── boilerplate.go.txt
└── main.go
```

## Makefile

コード生成やコントローラのビルドなどをおこなうためのMakefileです。

よく利用するターゲットとしては以下のものがあります。

| target       | 処理内容                             |
| -----        | -----                            |
| manifests    | goのソースコードからCRDやRBAC等のマニフェストを生成する |
| generate     | DeepCopy関数などを生成する                |
| docker-build | Dockerイメージのビルドをおこなう              |
| install      | KubernetesクラスタにCRDを適用する          |
| deploy       | Kubernetesクラスタにコントローラを適用する       |
| manager      | コントローラのビルド                       |
| run          | コントローラをローカル環境で実行する               |
| test         | テストを実行する                         |

Kubebuilder v2.3.1では、controller-gen v0.2.5を利用するようになっていますが、
Webhookのマニフェスト生成部分で問題があるため、以下のようにMakefile内の
controller-genのバージョンを最新にあげておくことを推奨します。

```diff
-	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
+	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
```

## PROJECT

ドメイン名やリポジトリのURLや生成したAPIの情報などが記述されています。

## main.go

あなたがこれから作成するコントローラのエントリーポイントとなるソースコードです。

ソースコード中に`// +kubebuilder:scaffold:imports`, `// +kubebuilder:scaffold:scheme`, `// +kubebuilder:scaffold:builder`などのコメントが記述されています。
Kubebuilderはこれらのコメントを目印にソースコードの自動生成をおこなうので、決して削除しないように注意してください。

## hack/boilerplate.go.txt

自動生成されるソースコードの先頭に挿入されるボイラープレートです。

デフォルトではApache 2 Licenseの文面が記述されているので、必要に応じて書き換えてください。

## config

configディレクトリ配下には、コントローラをKubernetesクラスタにデプロイするためのマニフェストが生成されます。

実装する機能によっては必要のないマニフェストも含まれているので、適切に取捨選択してください。

### manager

コントローラのDeploymentリソースのマニフェストです。
コントローラのコマンドラインオプションの変更などをおおこなった場合など、必要に応じて書き換えてください。

### rbac

各種権限を設定するためのマニフェストです。

`auth_proxy_`から始まる4つのファイルは、[kube-auth-proxy][]用のマニフェストです。
kube-auth-proxyを利用するとメトリクスエンドポイントへのアクセスをRBACで制限することができます。

`leader_election_role.yaml`と`leader_election_role_binding.yaml`は、リーダーエレクション機能
を利用するために必要な権限です。

`role.yaml`と`role_binding.yaml`は、コントローラが各種リソースにアクセスするための
権限を設定するマニフェストです。
この2つのファイルは基本的に自動生成されるものなので、開発者が編集する必要はありません。

必要のないファイルを削除した場合は、`kustomization.yaml`も編集してください。

### prometheus

Prometheus Operator用のカスタムリソースのマニフェストです。
Prometheus Operatorを利用している場合、このマニフェストを適用するとPrometheusが自動的に
コントローラのメトリクスを収集してくれるようになります。

Prometheus Operatorを利用していない場合は不要なので削除してしまいましょう。

### webhook

Admission Webhook機能を提供するためのマニフェストです。
Admission Webhook機能を利用しない場合は、ディレクトリごと消してしまってもよいでしょう。

### certmanager

Admission Webhook機能を提供するためには証明書が必要となります。
certmanagerディレクトリ下のマニフェストを適用すると、[cert-manager][]を利用して証明書を発行することができます。
Admission Webhook機能を利用しない場合や、cert-managerを利用せずに自前で証明書を用意する場合は
必要ないマニフェストなので、ディレクトリごと消してしまってもよいでしょう。

### default

上記のマニフェストをまとめて利用するための設定が記述されています。

`manager_auth_proxy_patch.yaml`は、[kube-auth-proxy][]を利用するために必要なパッチです。
kube-auth-proxyを利用しない場合は削除しても問題ありません。

`manager_webhook-patch.yaml`と`webhookcainjection_patch.yaml`は、Admission Webhook機能を利用する場合に
必要となるパッチファイルです。

利用するマニフェストに応じて、`kustomization.yaml`を編集してください。

[cert-manager]: https://github.com/jetstack/cert-manager
[kube-auth-proxy]: https://github.com/brancz/kube-rbac-proxy
