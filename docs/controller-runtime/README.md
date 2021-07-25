# controller-runtime

カスタムコントローラーを開発するためには、Kubernetesが標準で提供している[client-go](https://github.com/kubernetes/client-go), [apimachinery](https://github.com/kubernetes/apimachinery), [api](https://github.com/kubernetes/api)などのパッケージを利用することになります。

[controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)は、これらのパッケージを抽象化・隠蔽し、より簡単にカスタムコントローラーを実装可能にしたライブラリです。

抽象化・隠蔽しているとは言っても、Kubernetesのコンセプトに準拠する形で実装されています。
必要があればオプションを指定することにより、`client-go`や`apimachinery`が提供している機能のほとんどを利用できます。
controller-runtimeの設計コンセプトについて知りたい方は[KubeBuilder Design Principles](https://github.com/kubernetes-sigs/kubebuilder/blob/master/DESIGN.md#controller-runtime)を参照してください。

controller-runtimeが提供する代表的なコンポーネントには以下のものがあります。

- [manager.Manager](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#Manager)
  - 複数のコントローラーをまとめて管理するためのコンポーネント。
  - リーダー選出やメトリクスサーバーとしての機能など、カスタムコントローラーを実装するために必要な数多くの機能を提供します。
- [client.Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client?tab=doc#Client)
  - Kubernetesのkube-apiserverとやり取りするためのクライアント。
  - 監視対象のリソースをインメモリにキャッシュする機能などを持ち、カスタムリソースも型安全に扱うことが可能なクライアントとなっている。
- [reconcile.Reconciler](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile?tab=doc#Reconciler)
  - カスタムコントローラーが実装すべきインタフェース。

以降のページではこれらの機能を詳細に解説していきます。
