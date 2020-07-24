# カスタムコントローラの基礎

## Declarative

Kubernetesの最も基礎となる考え方Declarative(宣言型)

例えば、以下のようなYAMLフォーマットで記述されたマニフェストファイルを用意する。

kube-controller-managerと呼ばれるプログラムは、このマニフェストファイルの記述内容に従って、コンテナを立ち上げるようにKubeletに命令を出します。

立ち上げるだけではなく、マニフェストに記述された内容と、実際に起動しているPodの状態を常に比較し、もし差分が発生すればその差分がなくなるような動きをします。
どういうことかというと、例えば、ユーザーが以下のようにマニフェストを書き換えたとします。
するとkube-controller-managerは実際に起動しているコンテナとイメージのバージョンが異なることを検知し、コンテナを新しいイメージで立ち上げ直します。

逆に、何らかの問題が発生してコンテナが削除されてしまった場合、期待した状態と異なっているため、kube-controller-managerは再度コンテナを立ち上げる処理をおこないます。

常に、ユーザーが期待する状態と、実際の状態を比較して、その差分を埋めるような処理を実行する。

## CRDとCR

APIとResourceの関係
res


上述したように、KubernetesにはPodを始めとしてたくさんの標準リソースが用意されています。
Kubernetes上で複雑なシステムを構築したい場合、標準リソースだけでは
そこで、Kubernetesの利用者が自由に新しいリソースを定義するための仕組みが用意されています。

CRD(Custom Resource Definition)

コントローラでカスタムリソースを扱うためには、そのリソースのCRD(Custom Resource Definition)を定義する必要があります。このCRDはOpenAPI v3.0の形式で書く

- [CRDの例](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/tenant/config/crd/bases/multitenancy.example.com_tenants.yaml)

CRの例

## カスタムコントローラ

* コントローラ、オペレータ

あるリソースの状態をチェックして、それをあるべき姿に持っていこうとするプログラムのことをコントローラと呼びます。
先に紹介したkube-controller-managerは、PodコントローラやServiceコントローラなど、標準リソース用のコントローラの集合から構成されています。
一方で、ユーザーが定義したカスタムリソースを対象としたコントローラのことをカスタムコントローラと呼びます。

* Reconciliation loop

* 冪等

冪等ではない例。
例えば、先程のPodのコントローラ。
Podを作成しなければならない。
Reconcileが呼び出されるたびに新しいコンテナを立ち上げてしまう
1度作成すると、2回目に呼び出されたときに処理に失敗してエラーを返してしまう

* Edge-Driven Trigger, Level-Driven Trigger

Edge
* 状態の変化が発生した時点で処理が呼び出される

Level
* 状態を定期的にチェックして、特定の条件に入ったときにトリガーされる

`Phase`フィールドを用意して現在の状態のみを格納するのではなく、`Conditions`フィールドで各状態を
判断できるようにしておくことが推奨されています。

[API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties)

```yaml
status:
  phase: A
```

```yaml
status:
  phase: C
```

```yaml
status:
  conditions:
  - type: A
    status: True
  - type: B
    status: True
  - type: C
    status: False
```
