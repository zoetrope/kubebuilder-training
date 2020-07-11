---
title: "Add an API"
draft: true
weight: 13
---

`kubebuilder create api`コマンドを利用すると、カスタムリソースやコントローラの実装の雛形を生成することができます。

以下のコマンドを実行して、Tenantを表現するためのカスタムリソースと、Tenantリソースを扱うコントローラを生成してみましょう。

```console
$ kubebuilder create api --group multitenancy --version v1 --kind Tenant `--namespaced=false`
Create Resource [y/n]
y
Create Controller [y/n]
y
$ make manifests
```

`--group`,`--version`, `--kind`オプションは、生成するカスタムリソースのGVKを指定します。

`--namespace`オプションでは、生成するカスタムリソースをnamespace-scopeとcluster-scopeのどちらにするか指定できます。
今回のTenantリソースはnamespaceなどのcluster-scopeのリソースを扱うため、cluster-scopeを指定しています。

Custom ResourceとControllerのソースコードを生成するかどうか聞かれるので、今回はどちらも`y`と回答します。
カスタムリソースではなく既存のリソースを扱うコントローラを実装する場合は、`Create Resource [y/n]`に`n`と回答します。


```
├── api
│   └── v1
│       ├── groupversion_info.go
│       ├── tenant_types.go
│       └── zz_generated.deepcopy.go
├── config
│   ├── crd
│   │   ├── bases
│   │   │   └── multitenancy.example.com_tenants.yaml
│   │   ├── kustomization.yaml
│   │   ├── kustomizeconfig.yaml
│   │   └── patches
│   │       ├── cainjection_in_tenants.yaml
│   │       └── webhook_in_tenants.yaml
│   ├── rbac
│   │   ├── role.yaml
│   │   ├── role_binding.yaml
│   │   ├── tenant_editor_role.yaml
│   │   └── tenant_viewer_role.yaml
│   └── samples
│      └── multitenancy_v1_tenant.yaml
├── controllers
│   ├── suite_test.go
│   └── tenant_controller.go
└── main.go
```
