# つくって学ぶKubebuilder

本資料ではKuberbuilderを利用してカスタムコントローラ/オペレータを開発する方法について学びます。

## Kubebuilderとは

Kubebuilderは、Kubernetesを拡張するためのカスタムコントローラ/オペレータを開発するためのフレームワークです。

Kubernetesでは、PodやDeploymentなどのリソースを用意することで、
カスタムリソースと呼ばれる
独自のリソースを定義することができます。
このカスタムリソースの状態をチェックし、必要に応じて処理をおこなうプログラムをカスタムコントローラと呼びます。

また、カスタムコントローラを利用して、独自のソフトウェアのセットアップや運用を自動化するためのプログラムをオペレータと呼びます。

カスタムコントローラとは、
オペレータとは、

例として

- [cert-manager](https://cert-manager.io/docs/)
- [Sealed Secrets]()
- [MOCO]()
- [Rook-Cpeh](https://rook.io)

OperatorHub

Kubernetes向けのカスタムコントローラは、Kubernetesが公式に提供している[kubernetes/client-go](https://github.com/kubernetes/client-go)を利用するだけでも開発できます。
しかしそのためにはKubernetesの設計コンセプトを正しく理解する必要があり、たくさんのマニフェストを記述する必要もあるため、それほど簡単ではありません。

そこでKubebuilderでは、client-goを使いやすく抽象化したライブラリと、マニフェストを自動生成するツール群を提供することで、簡単にカスタムコントローラを開発できるようになっています。

Kubebuilderは、下記のツールとライブラリから構成されています。

- [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
  - カスタムコントローラのプロジェクトの雛形を生成するツール
- [controller-tools](https://github.com/kubernetes-sigs/controller-tools)
  - Goのソースコードからマニフェストを生成するツール
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
  - カスタムコントローラを実装するためのライブラリ

本資料ではカスタムコントローラをつくりながらこれらのツールの使い方を学んでいきます。

## 対応バージョン

* kubebuilder: v3.1.0
* controller-tools: v0.6.1
* controller-runtime: v0.9.2

## 更新履歴

* 2020/07/30: 初版公開
* 2021/04/29: Kubebuilder v3対応
* 2021/07/10: 
