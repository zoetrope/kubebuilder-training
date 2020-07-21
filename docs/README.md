# つくって学ぶKubebuilder

本資料ではKubebuilderを利用して、Kubernetesを拡張するカスタムコントローラの実装方法を学びます。

## Kubebuilderとは

Kubebuilderは、Kubernetesを拡張するためのカスタムコントローラや、Admission Webhookなどを開発するためのフレームワークです。

Kubernetes向けのカスタムコントローラは、Kubernetesが公式に提供している[kubernetes/client-go](https://github.com/kubernetes/client-go)を利用するだけでも開発することができます。
しかしそのためにはKubernetesの設計コンセプトを正しく理解する必要があり、たくさんのマニフェストを記述する必要もあるため、それほど簡単ではありません。

そこでKubebuilderでは、client-goを使いやすく抽象化したライブラリとマニフェストを自動生成するツール群を提供することで、簡単にカスタムコントローラを開発できるようになっています。

Kubebuilderは、下記のツールとライブラリから構成されています。

- [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
  - カスタムコントローラを実装するためのプロジェクトの雛形を生成するためのツール
- [controller-tools](https://github.com/kubernetes-sigs/controller-tools)
  - Goのソースコードからマニフェストを生成するツール
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
  - カスタムコントローラを実装するためのライブラリ

本資料ではカスタムコントローラをつくりながらこれらのツールの使い方を学んでいきます。

## 参考資料

本資料では基本的な部分を飛ばしていたりもするので、より詳しく知りたい場合は下記の資料を参考にしてください。

- [The Kubebuilder Book](https://book.kubebuilder.io/)
  - Kubebuilderの公式ドキュメントです。
- [実践入門Kubernetesカスタムコントローラへの道](https://nextpublishing.jp/book/11389.html)
  - カスタムコントローラを作成するための知識を幅広くかつ分かりやすく解説している書籍です。
  - client-go, Kubebuilder, Operator SDKを利用したコントローラの実装方法が解説されています。
- [Programming Kubernetes](https://learning.oreilly.com/library/view/programming-kubernetes/9781492047094/)
  - client-goやカスタムリソースなど、コントローラを開発する上で必要なKubernetesの構成要素を詳細に解説している書籍です。

## 参考実装

本資料で紹介しているテクニックは下記のプロジェクトで実際に使われているものを参考にしています。
興味があればぜひコードリーディングしてみてください。

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
- [Coil](https://github.com/cybozu-go/coil)
  - CNIプラグイン
  - v2(実装中)ではKubebuilderを利用して実装を全面的に書き換えています。
