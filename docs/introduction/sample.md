# 作成するカスタムコントローラ

TODO: 全面書き換え

本資料ではカスタムコントローラの実装例として、Markdownをレンダリングするサービスをデプロイするコントローラを実装します。


[import](../../codes/markdown-viewer/config/samples/viewer_v1_markdownview.yaml)


本資料でこれから開発するカスタムコントローラは、上記のカスタムリソースを読み取り、namespaceを作成したり管理者権限の設定をおこなうことになります。

ソースコードは以下にあるので参考にしてください。

- https://github.com/zoetrope/kubebuilder-training/tree/master/codes/markdown-viewer
