# コントローラのテスト

testenv


kubebuilderをセットアップすると、kube-apiserver

```console
$ ls /usr/local/kubebuilder/bin/
etcd  kube-apiserver  kubebuilder  kubectl  kustomize
```

## テスト環境のセットアップ

[import, title="controllers/suite_test.go"](../../codes/tenant/controllers/suite_test.go)

## コントローラのテスト

[import, title="controllers/tenant_controller_test.go"](../../codes/tenant/controllers/tenant_controller_test.go)

## テストの実行

make test
