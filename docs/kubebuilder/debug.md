# 手軽な動作確認

## make runによる実行

コントローラの開発中は何度もプログラムをビルドして実行し直すことになりますが、コンテナイメージのビルドに時間がかかるため非効率になりがちです。
また、コンテナとして動作しているプログラムはデバッグ実行がしにくいという問題もあります。

Kubebuilderによって生成されたMakefileには`make run`というターゲットが用意されており、コントローラをローカルのプロセスとして動作させることが可能です。

ただし、Kubernetesクラスタ上で動作させる場合といくつか違いがあるので、注意して利用してください。

* リーダーエレクション機能を利用する場合は、[Options](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager?tab=doc#Options)の`LeaderElectionNamespace`を指定する必要があります。
* 証明書やAPIのアクセス経路の問題があり、そのままではWebhookは動きません。
* コントローラがAPIサーバにアクセスするときの権限が、クラスタ内で動作させた場合と異なります(例えば、`$HOME/.kube/config`の設定を利用すると、kubectlと同じ権限を持つようになります)。
* [Downward API](https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/)が利用できません。

## Telepresenceによる実行

`make run`による実行では上記のようにいくつかの問題があります。
特にWebhookを手軽に動作させることができない点が不便です。

そこで、Telepresenceというツールを利用してコントローラをローカルのプロセスとして実行する方法を紹介します。

まずは下記のページを参考にしてTelepresenceをインストールしてください。

* [Installing Telepresence](https://www.telepresence.io/reference/install)

Telepresenceを利用する場合、ConfigMapやSecretをボリュームとしてアクセスする際のパスがPodとして実行する場合と異なります。
Kubebuilderで生成したカスタムコントローラでは、Webhookの証明書のパスを下記のように設定し、`NewManager`するときのOptionとして指定してあげましょう。

[import:"telepresence",unindent="true"](../../codes/tenant/main.go)

次に[Kindで動かしてみよう](./kind.md)の手順通りにコントローラをデプロイします。

最後に下記のコマンドで、Kubernetes上で動いているコントローラをローカルで動いているプロセスと置き換えます。

```console
telepresence --namespace tenant-system --swap-deployment tenant-controller-manager:manager --run make run
```

なお、kubebuilderが生成したコントローラのマニフェスト(`config/manager/manager.yaml`)には以下のように非常に小さなサイズのResourcesが指定されています。
これが原因でTelepresenceのProxyがOOM Killerに終了させられることがあるので、メモリサイズを増やしておくとよいでしょう。

```yaml
resources:
  limits:
    cpu: 100m
    memory: 30Mi
  requests:
    cpu: 100m
    memory: 20Mi
```
