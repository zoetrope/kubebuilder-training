# つくって学ぶKubebuilder

本資料ではKubernetesを拡張するカスタムコントローラの実装を通して、Kubebuilderの利用方法やKubernetesプログラミングについて学びます。

## Kubebuilderとは

Kubebuilderは、Kubernetesを拡張するためのカスタムコントローラや、Admission Webhookなどを開発するためのフレームワークです。

Kubernetes向けのカスタムコントローラは、Kubernetesが公式に提供している[kubernetes/client-go](https://github.com/kubernetes/client-go)を利用するだけでも開発することができます。
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

* kubebuilder: v3.0.0
* controller-tools: v0.5.0
* controller-runtime: v0.8.3

## 更新履歴

* 2020/07/30: 初版公開
* 2021/04/29: Kubebuilder v3対応
