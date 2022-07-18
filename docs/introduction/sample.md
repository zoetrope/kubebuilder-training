# MarkdownViewコントローラー

本資料では、カスタムコントローラーの例としてMarkdownViewコントローラーを実装することとします。
MarkdownViewコントローラーは、ユーザーが用意したMarkdownをレンダリングしてブラウザから閲覧できるようにサービスを提供するコントローラーです。

MarkdownのレンダリングにはmdBookを利用することとします。

- https://rust-lang.github.io/mdBook/

MarkdownViewコントローラーの主な処理の流れは次のようになります。

![MarkdownView Controller](./img/markdownview_controller.png)

- ユーザーはMarkdownViewカスタムリソースを作成します。
- MarkdownViewコントローラーは、作成されたMarkdownViewリソースの内容に応じて必要な各リソースを作成します。
  - カスタムリソースに記述されたMarkdownをConfigMapリソースとして作成します。
  - MarkdownをレンダリングするためのmdBookをDeploymentリソースとして作成します。
  - mdBookにアクセスするためのServiceリソースを作成します。
- ユーザーは、作成されたサービスを経由して、レンダリングされたMarkdownを閲覧できます。

MarkdownViewカスタムリソースには、以下のようにMarkdownの内容とレンダリングに利用するmdBookのコンテナイメージおよびレプリカ数を指定できるようにします。

[import](../../codes/50_completed/config/samples/view_v1_markdownview.yaml)

ソースコードは以下にあるので参考にしてください。

- https://github.com/zoetrope/kubebuilder-training/tree/master/codes

ディレクトリ構成は以下の通りです。

```
codes
├── 00_scaffold:  Kubebuilderで生成したコード
├── 10_tilt:      Tiltを利用した開発環境のセットアップを追加
├── 20_manifests: CRD, RBAC, Webhook用のマニフェストを生成
├── 30_client:    クライアントライブラリの利用例を追加
├── 40_reconcile: Reconcile処理、およびWebhookを実装
└── 50_completed: Finalizer, Recorder, モニタリングのコードを追加
```
