# Webhookの実装

Kubernetesでは、リソースの作成・更新・削除をおこなう直前にWebhookで任意の処理を実行するとことができます。
Mutating Webhookではリソースの値を書き換えることができ、Validating Webhookでは値の検証をおこなうことができます。

controller-runtimeでは、Mutating Webhookを実装するためのDefaulterとValidating Webhookを実装するためのValidatorが用意されています。

## Defaulterの実装

まずはDefaulterの実装です。
Defaultメソッドでは、MarkdownViewリソースの値を書き換えることができます。

[import:"head,webhook-defaulter,default"](../../codes/40_reconcile/api/v1/markdownview_webhook.go)

ここでは`r.Spec.ViewerImage`が空だった場合に、デフォルトのコンテナイメージを指定しています。

## Validatorの実装

次にValidatorの実装です。
ValidateCreate, ValidateUpdate, ValidateDeleteは、それぞれリソースの作成・更新・削除のタイミングで呼び出される関数です。
これらの関数の中でMarkdownViewリソースの内容をチェックし、エラーを返すことでリソースの操作を失敗させることができます。

[import:"head,webhook-validator,validate"](../../codes/40_reconcile/api/v1/markdownview_webhook.go)

今回はValidateCreateとValidateUpdateで同じバリデーションをおこなうことにしましょう。
`.Spec.Replicas`の値が1から5の範囲にない場合と、`.Spec.Markdowns`に`SUMMARY.md`が含まれない場合はエラーとします。

なお、ValidationWebhookを実装する際には`"k8s.io/apimachinery/pkg/util/validation/field"`パッケージが役立ちます。
このパッケージを利用してエラーの原因や問題のあるフィールドを指定することで、バリデーションエラー時のメッセージがわかりやすいものになります。

## 動作確認

それでは、Webhookの動作確認をしてみましょう。

Webhookの実装をおこなったカスタムコントローラーをKubernetesクラスターにデプロイし、下記のような`ViewerImage`を指定していないマニフェストを適用します。

```yaml
apiVersion: view.zoetrope.github.io/v1
kind: MarkdownView
metadata:
  name: markdownview-sample
spec:
  markdowns:
    SUMMARY.md: |
      # Summary

      - [Page1](page1.md)
    page1.md: |
      # Page 1

      一ページ目のコンテンツです。
  replicas: 1
```

作成されたリソースを確認して、`ViewerImage`にデフォルトのコンテナイメージ名が入っていれば成功です。

```
$ kubectl get markdownview markdownview-sample -o jsonpath="{.spec.viewerImage}"
peaceiris/mdbook:latest
```

続いてバリデーションWebhookの動作も確認してみましょう。

先ほど作成したリソースを編集して`replicas`を大きな値にしたり、`markdowns`に`SUMMARY.md`を含めないようにしたりしてみましょう。
以下のようなエラーが発生すれば成功です。

```
$ kubectl edit markdownview markdownview-sample

markdownviews.view.zoetrope.github.io "markdownview-sample" was not valid:
 * spec.replicas: Invalid value: 10: replicas must be in the range of 1 to 5.
 * spec.markdowns: Required value: markdowns must have SUMMARY.md.
```
