# つくって学ぶKubebuilder

本資料ではKuberbuilderを利用してカスタムコントローラ/オペレータを開発する方法について学びます。

## Kubebuilderとは

Kubebuilderは、Kubernetesを拡張するためのカスタムコントローラ/オペレータを開発するためのフレームワークです。

Kubernetesでは、標準で用意されているDeploymentやServiceなどのリソースを利用することで、簡単にアプリケーションのデプロイやサービスの提供ができるようになっています。
さらに標準リソースを利用するだけでなく、ユーザーが独自のカスタムリソースを定義してKubernetesを機能拡張することが可能になっています。
このカスタムリソースを扱うためのプログラムをカスタムコントローラと呼びます。
また、カスタムコントローラを利用して独自のソフトウェアのセットアップや運用を自動化するためのプログラムをオペレータと呼びます。

カスタムコントローラやオペレータの実装例として、次のようなものがあります。

- [cert-manager](https://cert-manager.io/docs/)
- [MOCO](https://github.com/cybozu-go/moco)

cert-managerは、CertificateリソースやIssuerリソースなどのカスタムリソースを利用して、証明書の発行を自動化することができるカスタムコントローラです。
MOCOは、MySQLClusterリソースやBackupPolicyリソースを利用して、MySQLクラスタの構築や自動バックアップを管理するためのオペレータです。

Kubebuilderでは、client-goを使いやすく抽象化したライブラリとマニフェストを自動生成するツールを提供することで、簡単にカスタムコントローラやオペレータを開発することができます。

Kubebuilderは、下記のツールとライブラリから構成されています。

- [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)
  - カスタムコントローラのプロジェクトの雛形を生成するツール
- [controller-tools](https://github.com/kubernetes-sigs/controller-tools)
  - Goのソースコードからマニフェストを生成するツール
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
  - カスタムコントローラを実装するためのフレームワーク・ライブラリ

本資料ではこれらのツールの利用してカスタムコントローラを実装する方法を学んでいきます。

## 対応バージョン

* kubebuilder: v3.1.0
* controller-tools: v0.6.1
* controller-runtime: v0.9.2

## 更新履歴

* 2020/07/30: 初版公開
* 2021/04/29: Kubebuilder v3対応
* 2021/07/10: 
