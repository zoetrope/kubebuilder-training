# Summary

* [つくって学ぶKubebuilder](README.md)
  * [インストール](introduction/installation.md)
  * [カスタムコントローラーの基礎](introduction/basics.md)
  * [MarkdownViewコントローラー](introduction/sample.md)
  * [参考情報](introduction/references.md)
* [Kubebuilder](kubebuilder/README.md)
  * [プロジェクトの雛形作成](kubebuilder/new-project.md)
  * [APIの雛形作成](kubebuilder/api.md)
  * [Webhookの雛形作成](kubebuilder/webhook.md)
  * [カスタムコントローラーの動作確認](kubebuilder/kind.md)
  <!-- * [手軽な動作確認](kubebuilder/debug.md) -->
* [controller-tools](controller-tools/README.md)
  * [CRDマニフェストの生成](controller-tools/crd.md)
  <!-- * [CRDマニフェストの生成(応用編)](controller-tools/advanced_crd.md) -->
  * [RBACマニフェストの生成](controller-tools/rbac.md)
  * [Webhookマニフェストの生成](controller-tools/webhook.md)
* [controller-runtime](controller-runtime/README.md)
  * [クライアントの使い方](controller-runtime/client.md)
  <!-- * [Server Side Apply](controller-runtime/ssa.md) -->
  * [Reconcileの実装](controller-runtime/reconcile.md)
  * [コントローラーのテスト](controller-runtime/controller_test.md)
  * [Webhookの実装](controller-runtime/webhook.md)
  * [Webhookのテスト](controller-runtime/webhook_test.md)
  * [リソースの削除](controller-runtime/deletion.md)
  * [Manager](controller-runtime/manager.md)
  * [モニタリング](controller-runtime/monitoring.md)
  <!-- * [応用テクニック](controller-runtime/advanced.md) -->
  <!-- * [CRDのバージョニング](controller-runtime/versioning.md) -->
