# つくって学ぶKubebuilder

本資料ではKubebuilderを利用して、Kubernetesを拡張するカスタムコントローラの実装方法を学びます。

## Kubebuilderとは

なぜKubebuilderを使うのか。
Kubernetesはカスタムコントローラを実装するための機能を標準で提供しています。

たくさんのマニフェストを用意し、Kubernetesの設計コンセプトを理解してたくさんのコードを書くのは大変。

- [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
  - カスタムコントローラを実装するためのプロジェクトの雛形を生成するためのツール
- [controller-tools](https://github.com/kubernetes-sigs/controller-tools)
  - Goのソースコードからマニフェストを生成するツール
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
  - カスタムコントローラを実装するためのライブラリ

## 参考資料

本資料ではKubebuilderの実践的な内容を解説しています。
基本的な内容を知りたい場合は、下記の資料を参考にしてください。

- [The Kubebuilder Book](https://book.kubebuilder.io/)
    - Kubebuilderの公式ドキュメントです。
- [実践入門Kubernetesカスタムコントローラへの道](https://nextpublishing.jp/book/11389.html)
    - カスタムコントローラを作成するための知識を幅広くに分かりやすく解説している書籍です。
    - client-go, Kubebuilder, Operator SDKを利用したコントローラの実装方法が解説されています。
- [Programming Kubernetes](https://learning.oreilly.com/library/view/programming-kubernetes/9781492047094/)
    - client-goやカスタムリソースなど、コントローラを開発する上で必要なKubernetesの構成要素を詳細に解説している書籍です。

## 参考実装

Kubebuilderの活用事例として、我々が実装しているOSSを紹介します。

- [TopoLVM](https://github.com/topolvm/topolvm)
    - LVMを利用したDynamic Provisioning可能なCSIプラグイン実装
- [Contour Plus](https://github.com/cybozu-go/contour-plus)
    - Ingress Controller [Contour](https://github.com/projectcontour/contour)を機能拡張するためのコントローラ
- [neco-admission](https://github.com/cybozu/neco-containers/tree/master/admission)
    - カスタムポリシーを適用するためのAdmission WebHook実装
- [local-pv-provisioner](https://github.com/cybozu/neco-containers/tree/master/local-pv-provisioner)
    - 指定した条件にマッチしたデバイスから自動的にlocal Persistent Volumeリソースを作成するコントローラ
- [MOCO](https://github.com/cybozu-go/moco)
    - MySQLクラスタの構築を自動化するオペレータ(実装中)
