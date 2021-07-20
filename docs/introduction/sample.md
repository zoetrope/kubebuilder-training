# 作成するカスタムコントローラ

本資料では、カスタムコントローラの例としてMarkdownViewコントローラを実装することとします。
MarkdownViewコントローラは、ユーザーが用意したMarkdownをレンダリングしてブラウザから閲覧できるようにサービスを提供するコントローラです。

MarkdownのレンダリングにはmdBookを利用することとします。

- https://rust-lang.github.io/mdBook/

MarkdownViewコントローラの主な処理の流れは次のようになります。

図を書く

- ユーザーはレンダリングしたいMarkdownを含むカスタムリソースを用意します。
- 指定したMarkdownをConfigMapリソースとして作成します。
- MarkdownをレンダリングするためのmdBookをDeploymentリソースとして作成します。
- レンダリングしたMarkdownを閲覧できるようにServiceリソースを作成します。

カスタムリソースは以下のように、Markdownの内容、レンダリングに利用するmdBookのコンテナイメージおよびレプリカ数を指定できるようにします。

[import](../../codes/markdown-viewer/config/samples/viewer_v1_markdownview.yaml)

ソースコードは以下にあるので参考にしてください。

- https://github.com/zoetrope/kubebuilder-training/tree/master/codes/markdown-viewer
