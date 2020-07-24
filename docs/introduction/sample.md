# 作成するカスタムコントローラ

本資料ではカスタムコントローラの実装例として、Kubernetesでマルチテナンシーを実現するためのテナントコントローラを実装します。

Kubernetesでは、namespaceという仕組みにより各種リソースの分離をおこなうことが可能になっています。
しかし複数のチームがKubernetesを利用するユースケースでは、namespaceだけでは機能が不足する場合があります。
例えば1つのチームが複数のnamespaceを利用したいと思っても、namespaceの作成には強い権限が必要となるため、チームのメンバーがnamespaceを自由に追加することはできません。

そこでテナントという仕組みを考えてみましょう。
- テナントは複数のnamespaceから構成される
- テナントの管理者に指定されたユーザーはテナントに新しいnamespaceを自由に追加・削除できる

次にこのテナントを表現するためのカスタムリソースを考えてみます。
下記のようにテナントを構成するnamespace名の一覧と、namespace名のプリフィックス、管理者を指定できるようにしましょう。

```yaml
apiVersion: multitenancy.example.com/v1
kind: Tenant
metadata:
  name: sample
spec:
  namespaces:
    - test1
    - test2
  namespacePrefix: sample-
  admin:
    kind: User
    name: test
    namespace: default
    apiGroup: rbac.authorization.k8s.io
```

本資料でこれから開発するカスタムコントローラは、上記のカスタムリソースを読み取り、実際のnamespaceを作成したり管理者権限の設定をおこなうことになります。

ソースコードは以下にありますので参考にしてください。

- https://github.com/zoetrope/kubebuilder-training/tree/master/codes/tenant
