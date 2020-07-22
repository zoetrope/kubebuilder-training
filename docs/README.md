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

