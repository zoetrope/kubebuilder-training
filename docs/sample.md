# 作成するコントローラ

本資料ではカスタムコントローラの実装例として、Kubernetesでマルチテナンシーを実現するための
テナントコントローラを実装します。

テナントコントローラは以下のような機能を持ちます。

- 各テナントは複数のnamespaceから構成される
- テナントリソースを作成するとテナントに属するnamespaceが作成される
- namespaceの名前にはprefixを指定することができる
- namespaceのprefixは途中で変更することができない
- テナントには管理者ユーザーを指定することができる
- 管理者ユーザーはテナントにnamespaceを追加・削除することができる

マルチテナンシーを実現するためには機能不足ですが、Kubebuilderを学ぶためには十分な要素が含まれています。

ソースコードは以下にあります。

- https://github.com/zoetrope/kubebuilder-training/tree/master/codes/tenant
