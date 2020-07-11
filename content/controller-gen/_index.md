---
title: "controller-gen"
draft: false
weight: 20
---

controller-tools
controller-genとは

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
