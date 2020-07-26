# APIの雛形作成

`kubebuilder create api`コマンドを利用すると、カスタムリソースやカスタムコントローラの実装の雛形を生成することができます。

以下のコマンドを実行して、Tenantを表現するためのカスタムリソースと、テナントリソースを扱うカスタムコントローラを生成してみましょう。

```console
$ kubebuilder create api --group multitenancy --version v1 --kind Tenant `--namespaced=false`
Create Resource [y/n]
y
Create Controller [y/n]
y
$ make manifests
```

`--group`,`--version`, `--kind`オプションは、生成するカスタムリソースのGVKを指定します。
- `--kind`: 作成するリソースの名前を指定します。
- `--group`: テナントリソースが属するグループ名を指定します。
- `--version`: 適切なバージョンを指定します。今後仕様が変わる可能性がありそうなら`v1alpha1`や`v1beta1`を指定し、安定版のリソースを作成するのであれば`v1`を指定します。

`--namespace`オプションでは、生成するカスタムリソースをnamespace-scopedとcluster-scopedのどちらにするか指定できます。
今回のテナントリソースはnamespaceなどのcluster-scopedのリソースを扱うため、cluster-scopedを指定しています。

カスタムリソースとコントローラのソースコードを生成するかどうか聞かれるので、今回はどちらも`y`と回答します。

コマンドの実行に成功すると、下記のようなファイルが新たに生成されます。

```
├── api
│    └── v1
│        ├── groupversion_info.go
│        ├── tenant_types.go
│        └── zz_generated.deepcopy.go
├── config
│    ├── crd
│    │    ├── bases
│    │    │    └── multitenancy.example.com_tenants.yaml
│    │    ├── kustomization.yaml
│    │    ├── kustomizeconfig.yaml
│    │    └── patches
│    │        ├── cainjection_in_tenants.yaml
│    │        └── webhook_in_tenants.yaml
│    ├── rbac
│    │    ├── role.yaml
│    │    ├── role_binding.yaml
│    │    ├── tenant_editor_role.yaml
│    │    └── tenant_viewer_role.yaml
│    └── samples
│        └── multitenancy_v1_tenant.yaml
├── controllers
│    ├── suite_test.go
│    └── tenant_controller.go
└── main.go
```

それぞれのファイルの内容をみていきましょう。

## api/v1

`tenant_types.go`は、テナントリソースをGo言語のstructで表現したものです。
今後、テナントリソースの定義をおこなう場合にはこのファイルを編集していくことになります。

`groupversion_info.go`は初期生成後に編集する必要はありません。
`zz_generated.deepcopy.go`は`tenant_types.go`の内容から自動生成されるファイルなので編集する必要はありません。

## controllers

`tenant_controller.go`は、カスタムコントローラのメインロジックになります。
今後、カスタムコントローラの処理は基本的にこのファイルに書いていくことになります。

`suite_test.go`はテストコードです。詳細は[コントローラのテスト](../controller-runtime/controller_test.md)で解説します。

## main.go

`main.go`には、下記のようなコントローラの初期化処理が追加されています。

```go
if err = (&controllers.TenantReconciler{
	Client: mgr.GetClient(),
	Log:    ctrl.Log.WithName("controllers").WithName("Tenant"),
	Scheme: mgr.GetScheme(),
}).SetupWithManager(mgr); err != nil {
	setupLog.Error(err, "unable to create controller", "controller", "Tenant")
	os.Exit(1)
}
```

## config

configディレクトリ下には、いくつかのファイルが追加されています。

### crd

crdディレクトリにはCRD(Custom Resource Definition)のマニフェストが追加されています。

これらのマニフェストは`api/v1/tenant_types.go`から自動生成されるものなので、基本的に編集する必要はありません。
ただし、Conversion Webhookを利用したい場合は、`cainjection_in_tenants.yaml`と`webhook_in_tenants.yaml`のパッチを利用するように`kustomization.yaml`を書き換えてください。

### rbac

`role.yaml`と`role_binding.yaml`は、テナントリソースを扱うための権限が追加されています。

`tenant_editor_role.yaml`と`tenant_viewer_role.yaml`は、テナントリソースの編集・読み取りの権限です。
必要に応じて利用しましょう。

### samples

カスタムリソースのサンプルマニフェストです。
テストで利用したり、ユーザー向けに提供できるように記述しておきましょう。
