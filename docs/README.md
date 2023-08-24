# つくって学ぶKubebuilder

本資料ではKuberbuilderを利用してカスタムコントローラー/オペレーターを開発する方法について学びます。

## Kubebuilderとは

Kubebuilderは、Kubernetesを拡張するためのカスタムコントローラー/オペレーターを開発するためのフレームワークです。

Kubernetesでは、標準で用意されているDeploymentやServiceなどのリソースを利用することで、簡単にアプリケーションのデプロイやサービスの提供ができるようになっています。
さらに標準リソースを利用するだけでなく、ユーザーが独自のカスタムリソースを定義してKubernetesを機能拡張することが可能になっています。
このカスタムリソースを扱うためのプログラムをカスタムコントローラーと呼びます。
また、カスタムコントローラーを利用して独自のソフトウェアのセットアップや運用を自動化するためのプログラムをオペレーターと呼びます。

カスタムコントローラーやオペレーターの実装例として、次のようなものがあります。

- [cert-manager](https://cert-manager.io/docs/)
- [MOCO](https://github.com/cybozu-go/moco)

cert-managerは、CertificateリソースやIssuerリソースなどのカスタムリソースを利用して、証明書の発行を自動化できるカスタムコントローラーです。
MOCOは、MySQLClusterリソースやBackupPolicyリソースを利用して、MySQLクラスターの構築や自動バックアップを管理するためのオペレーターです。

Kubebuilderでは、client-goを使いやすく抽象化したライブラリとマニフェストを自動生成するツールを提供することで、簡単にカスタムコントローラーやオペレーターを開発できます。

Kubebuilderは、下記のツールとライブラリから構成されています。

- [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
  - カスタムコントローラーのプロジェクトの雛形を生成するツール
- [controller-tools](https://github.com/kubernetes-sigs/controller-tools)
  - Goのソースコードからマニフェストを生成するツール
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
  - カスタムコントローラーを実装するためのフレームワーク・ライブラリ

本資料ではこれらのツールの利用してカスタムコントローラーを実装する方法を学んでいきます。

## 対応バージョン

* Kubebuilder: v3.11.1
* controller-tools: v0.12.0
* controller-runtime: v0.15.0

## 更新履歴

* 2020/07/30: 初版公開
* 2021/04/29: Kubebuilder v3対応
* 2021/07/25: サンプルをMarkdownViewコントローラーに変更、本文の全面見直し
* 2022/06/20: Kubebuilder v3.4.1対応
* 2022/07/18: Kubebuilder v3.5.0対応、サンプルコードの見直し
* 2023/08/24: Kubebuilder v3.11.1対応
