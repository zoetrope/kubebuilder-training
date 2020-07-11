---
title: "Test"
draft: true
weight: 15
---

testenv


kubebuilderをセットアップすると、kube-apiserver

```console
$ ls /usr/local/kubebuilder/bin/
etcd  kube-apiserver  kubebuilder  kubectl  kustomize
```

## テスト環境のセットアップ

{{% code file="/static/codes/tenant/controllers/suite_test.go" language="go" %}}

## コントローラのテスト

{{% code file="/static/codes/tenant/controllers/tenant_controller_test.go" language="go" %}}

## テストの実行

make test
