# カスタムコントローラーのリリース

カスタムコントローラーを開発したら、それをリリースする必要があります。

Kubebuilderが生成したプロジェクトでは、`make docker-push`でコンテナイメージをpushしたり、`make build-installer`でカスタムコントローラーをインストールするためのマニフェストを生成することができます。
しかし、リリースするための手順が十分に提供されているわけではありません。

GoReleaserによるコンテナイメージのリリース方法や、Chart ReleaserによるHelm Chartのリリース方法を以下の記事にまとめましたので、参考にしてみてください。

- [Kubernetes カスタムコントローラー楽々メンテナンス](https://zenn.dev/zoetro/articles/kubernetes-controller-maintenance)
