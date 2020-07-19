# controller-tools

Kubebuilderでは、カスタムコントローラの開発を補助するためのツール群として[controller-tools](https://github.com/kubernetes-sigs/controller-tools)を提供しています。

controller-toolsには下記のツールが含まれていますが、本資料ではcontroller-genのみを取り扱います。

- controller-gen
- type-scaffold
- helpgen

##  controller-genとは

- crd
- schemapatch
- webhook
- rbac
- object

`// +kubebuilder:`というマーカーを目印にしてコード生成をおこないます。

利用可能なマーカーは下記のコマンドで確認することができます。

```console
$ controller-gen crd -w
$ controller-gen webhook -w
```
