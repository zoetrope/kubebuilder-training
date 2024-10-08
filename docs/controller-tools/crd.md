# CRDマニフェストの生成

コントローラーでカスタムリソースを扱うためには、そのリソースのCRD(Custom Resource Definition)を定義する必要があります。
CRDのマニフェストは複雑で、手書きで作成するにはかなりの手間がかかります。

そこでKubebuilderではcontroller-genというツールを提供しており、Goで記述したstructからCRDを生成できます。

まずは`kubebuilder create api`コマンドで生成された`api/v1/markdownview_types.go`を見てみましょう。

[import](../../codes/00_scaffold/api/v1/markdownview_types.go)

`MarkdownViewSpec`, `MarkdownViewStatus`, `MarkdownView`, `MarkdownViewList`という構造体が定義されており、`//+kubebuilder:`から始まるマーカーコメントがいくつか付与されています。
controller-genは、これらの構造体とマーカーを頼りにCRDの生成をおこないます。

`MarkdownView`がカスタムリソースの本体となる構造体です。`MarkdownViewList`は`MarkdownView`のリストを表す構造体です。これら2つの構造体は基本的に変更することはありません。
`MarkdownViewSpec`と`MarkdownViewStatus`は`MarkdownView`構造体を構成する要素です。この2つの構造体を書き換えてカスタムリソースを定義していきます。

一般的にカスタムリソースの`Spec`はユーザーが記述するもので、システムのあるべき状態をユーザーからコントローラーに伝える用途として利用されます。
一方の`Status`は、コントローラーが処理した結果をユーザーや他のシステムに伝える用途として利用されます。

なお、CRDを利用してKubernetes APIを拡張する際には、以下の規約に従うことが推奨されています。一度目を通しておくとよいでしょう。

- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)

## MarkdownViewSpec

さっそく`MarkdownViewSpec`を書き換えていきましょう。

[作成するカスタムコントローラ](../introduction/sample.md)において、MarkdownViewコントローラーが扱うカスタムリソースとして下記のようなマニフェストを例示しました。

[import](../../codes/20_manifests/config/samples/view_v1_markdownview.yaml)

上記のマニフェストを扱うための構造体を用意しましょう。

[import:"spec"](../../codes/20_manifests/api/v1/markdownview_types.go)

まず下記の3つのフィールドを定義します。

- `Markdowns`: 表示したいマークダウンファイルの一覧
- `Replicas`: Viewerのレプリカ数
- `ViewerImage`: Markdownの表示に利用するViewerのイメージ名

各フィールドの上に`// +kubebuilder`という文字列から始まるマーカーと呼ばれるコメントが記述されています。
これらのマーカーによって、生成されるCRDの内容を制御できます。

付与できるマーカーは`controller-gen crd -w`コマンドで確認できます。

### Required/Optional

`Markdowns`フィールドには`+kubebuiler:validation:Required`マーカーが付与されています。
これはこのフィールドが必須項目であることを示しており、ユーザーがマニフェストを記述する際にこの項目を省略できません。
一方の`Replicas`と`ViewerImage`には`+optional`が付与されており、この項目が省略可能であることを示しています。

マーカーを指定しなかった場合はデフォルトでRequiredなフィールドになります。

なお、ファイル内に下記のマーカーを配置すると、デフォルトの挙動をOptionalに変更できます。

```
// +kubebuilder:validation:Optional
```

`+optional`マーカーを付与しなくても、フィールドの後ろのJSONタグに`omitempty`を付与した場合は、自動的にOptionalなフィールドとなります。

```go
type SampleSpec struct {
	Value string `json:"value,omitempty"`
}
```

Optionalなフィールドは、以下のようにフィールドの型をポインタにできます。
これによりマニフェストで値を指定しなかった場合の挙動が異なります。
ポインタ型にした場合はnullが入り、実体にした場合はその型の初期値(intの場合は0)が入ります。

```go
type SampleSpec struct {
	// +optional
	Value1 int  `json:"value1"`
	// +optional
	Value2 *int `json:"value2"`
}
```

### Validation

Kubebuilderには`Required`以外にも様々なバリデーションが用意されています。
詳しくは`controller-gen crd -w`コマンドで確認してください。

- リストの最小要素数、最大要素数
- 文字列の最小長、最大長
- 数値の最小値、最大値
- 正規表現にマッチするかどうか
- リスト内の要素がユニークかどうか

## MarkdownViewStatus

次にMarkdownViewリソースの状態を表現するための`MarkdownViewStatus`を書き換えます。

[import:"status"](../../codes/20_manifests/api/v1/markdownview_types.go)

カスタムリソースの状態を表現するには、[`metav1.Condition`](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1)を利用することが一般的です。
今回のカスタムコントローラーでは、ConditionのTypeとして`Available`,`Degraded`の3つの状態を表現できるようにしました。

- Available: レンダリングされたMarkdownが閲覧可能な状態
- Degraded: Reconcile処理に失敗した状態

## MarkdownView

続いて`MarkdownView`構造体のマーカーを見てみましょう。

[import:"markdown-view"](../../codes/20_manifests/api/v1/markdownview_types.go)

Kubebuilderが生成した初期状態では、`+kubebuilder:object:root=true`と`+kubebuilder:subresource`の2つのマーカーが指定されています。
ここではさらに`+kubebuilder:printcolumn`を追加することとします。
  
`+kubebuilder:object:root=true`は、`MarkdownView`構造体がカスタムリソースのrootオブジェクトであることを表すマーカーです。

`+kubebuilder:subresource`と`+kubebuilder:printcolumn`マーカーについて、以降で解説します。

### subresource

`+kubebuilder:subresource:status`というマーカーを追加すると、`status`フィールドがサブリソースとして扱われるようになります。

Kubernetesでは、すべてのリソースはそれぞれ独立したAPIエンドポイントを持っており、APIサーバー経由でリソースの取得・作成・変更・削除をおこなうことができます。

サブリソースを有効にすると`status`フィールドがメインのリソースと独立したAPIエンドポイントを持つようになります。

これによりメインのリソース全体を取得・更新しなくても、`status`のみの取得や更新が可能になります。
ただし、あくまでもメインのリソースに属するサブのリソースなので、個別の作成や削除はできません。

ユーザーが`spec`フィールドを記述し、コントローラーが`status`フィールドを記述するという役割分担を明確にできるので、基本的には`status`はサブリソースにしておくのがよいでしょう。
なおKubebuilder v3では、`status`フィールドがサブリソースに指定するマーカーが最初から指定されるようになりました。

CRDでは任意のフィールドをサブリソースにはできず、`status`と`scale`の2つのフィールドのみに対応しています。

### printcolumn

`+kubebuilder:printcolumn`マーカーを付与すると、kubectlでカスタムリソースを取得したときに表示する項目を指定できます。

表示対象のフィールドはJSONPathで指定可能です。
例えば、`JSONPath=".spec.replicas"`と記述すると、kubectl getしたときに`.spec.replicas`の値が表示されます。

kubectlでMarkdownViewリソースを取得すると、下記のようにREPLICASやSTATUSの値が表示されていることが確認できます。

```
$ kubectl get markdownview
NAME                  REPLICAS   AVAILABLE
MarkdownView-sample   1          
```

## CRDマニフェストの生成

最後にGoで記述したstructからCRDを生成してみましょう。

以下のコマンドを実行してください。

```console
$ make manifests
```

すると、以下のようなCRDのマニフェストファイルが生成されます。

[import](../../codes/20_manifests/config/crd/bases/view.zoetrope.github.io_markdownviews.yaml)
