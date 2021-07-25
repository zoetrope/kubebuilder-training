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

[import](../../codes/markdown-view/config/samples/view_v1_markdownview.yaml)

ソースコードは以下にあるので参考にしてください。

- https://github.com/zoetrope/kubebuilder-training/tree/master/codes/markdown-view
