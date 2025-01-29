# カスタムコントローラーの動作確認

Kubebuilderコマンドで生成したプロジェクトをビルドし、[Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)環境で動かしてみましょう。

Kindとはローカル環境にKubernetesクラスターを構築するためのツールで、手軽にコントローラーのテストや動作確認をおこなうことができます。

## kindの立ち上げ

まずはkindコマンドを利用してKubernetesクラスターを作成します。

```console
$ kind create cluster
```

## cert-managerのインストール

Webhook用の証明書を発行するためにcert-managerが必要となります。
下記のコマンドを実行してcert-managerのデプロイをおこないます。([参考](https://cert-manager.io/docs/installation/kubectl/))

```console
$ kubectl apply --validate=false -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
```

cert-managerのPodが起動したことを確認しましょう。

```console
$ kubectl get pod -n cert-manager
NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager-7dd5854bb4-whlcn              1/1     Running   0          26s
cert-manager-cainjector-64c949654c-64wjk   1/1     Running   0          26s
cert-manager-webhook-6bdffc7c9d-hkr8h      1/1     Running   0          26s
```

## カスタムコントローラーのコンテナイメージの用意

以下のコマンドを実行して、カスタムコントローラーのコンテナイメージをビルドします。

```console
$ make docker-build
```

このコンテナイメージを利用するためには、ビルドしたコンテナイメージをDockerHubなどのコンテナレジストリに登録するか、kind環境にロードする必要があります。

ここでは下記のコマンドを利用してkind環境にコンテナイメージをロードしましょう。

```console
$ kind load docker-image controller:latest
```

なおコンテナイメージのタグ名に`latest`を指定した場合、ImagePullPolicyが`Always`になり、ロードしたコンテナイメージが利用されない場合があります。
([参考](https://kind.sigs.k8s.io/docs/user/quick-start/#loading-an-image-into-your-cluster))

そこで、`config/manager/manager.yaml`に`imagePullPolicy: IfNotPresent`を追加しておきます。

```diff
           - --leader-elect
           - --health-probe-bind-address=:8081
         image: controller:latest
+        imagePullPolicy: IfNotPresent
         name: manager
         securityContext:
           allowPrivilegeEscalation: false
```

## カスタムコントローラーの動作確認

次に以下のコマンドを実行して、作成したCRDをKubernetesクラスターに適用します。

```console
$ make install
```

続いて以下のコマンドを実行して、カスタムコントローラーをデプロイするためのマニフェストを適用します。

```console
$ make deploy
```

カスタムコントローラーのPodがRunningになったことを確認してください。

```console
$ kubectl get pod -n markdown-view-system
NAME                                                READY   STATUS    RESTARTS   AGE
markdown-view-controller-manager-5bc678bbf9-vb9r5   1/1     Running   0          30s
```

次にカスタムコントローラーのログを表示させておきます。

```console
$ kubectl logs -n markdown-view-system deployments/markdown-view-controller-manager -c manager -f
```

最後にサンプルのカスタムリソースを適用します。

```console
$ kubectl apply -f config/samples/view_v1_markdownview.yaml
```

以下のようにWebhookやReconcileのメッセージがカスタムコントローラーのログに表示されていれば成功です。

```console
2024-08-11T07:32:54Z    INFO    controller-runtime.builder      Registering a mutating webhook  {"GVK": "view.zoetrope.github.io/v1, Kind=MarkdownView", "path": "/mutate-view-zoetrope-github-io-v1-markdownview"}
2024-08-11T07:32:54Z    INFO    controller-runtime.webhook      Registering webhook     {"path": "/mutate-view-zoetrope-github-io-v1-markdownview"}
2024-08-11T07:32:54Z    INFO    controller-runtime.builder      Registering a validating webhook        {"GVK": "view.zoetrope.github.io/v1, Kind=MarkdownView", "path": "/validate-view-zoetrope-github-io-v1-markdownview"}
2024-08-11T07:32:54Z    INFO    controller-runtime.webhook      Registering webhook     {"path": "/validate-view-zoetrope-github-io-v1-markdownview"}
2024-08-11T07:32:54Z    INFO    setup   starting manager
2024-08-11T07:32:54Z    INFO    controller-runtime.metrics      Starting metrics server
2024-08-11T07:32:54Z    INFO    setup   disabling http/2
2024-08-11T07:32:54Z    INFO    starting server {"name": "health probe", "addr": "[::]:8081"}
2024-08-11T07:32:54Z    INFO    controller-runtime.webhook      Starting webhook server
2024-08-11T07:32:54Z    INFO    setup   disabling http/2
2024-08-11T07:32:54Z    INFO    controller-runtime.certwatcher  Updated current TLS certificate
I0811 07:32:54.398221       1 leaderelection.go:250] attempting to acquire leader lease markdown-view-system/3ca5b296.zoetrope.github.io...
2024-08-11T07:32:54Z    INFO    controller-runtime.webhook      Serving webhook server  {"host": "", "port": 9443}
2024-08-11T07:32:54Z    INFO    controller-runtime.certwatcher  Starting certificate watcher
I0811 07:32:54.408485       1 leaderelection.go:260] successfully acquired lease markdown-view-system/3ca5b296.zoetrope.github.io
2024-08-11T07:32:54Z    DEBUG   events  markdown-view-controller-manager-7b7bf8bc56-pm7tp_693ac946-5132-4674-8770-81b2dcdb8f19 became leader       {"type": "Normal", "object": {"kind":"Lease","namespace":"markdown-view-system","name":"3ca5b296.zoetrope.github.io","uid":"e0d1dda2-7e64-40b1-9c67-f8b056465798","apiVersion":"coordination.k8s.io/v1","resourceVersion":"1411"}, "reason": "LeaderElection"}
2024-08-11T07:32:54Z    INFO    Starting EventSource    {"controller": "markdownview", "controllerGroup": "view.zoetrope.github.io", "controllerKind": "MarkdownView", "source": "kind source: *v1.MarkdownView"}
2024-08-11T07:32:54Z    INFO    Starting Controller     {"controller": "markdownview", "controllerGroup": "view.zoetrope.github.io", "controllerKind": "MarkdownView"}
2024-08-11T07:32:54Z    INFO    Starting workers        {"controller": "markdownview", "controllerGroup": "view.zoetrope.github.io", "controllerKind": "MarkdownView", "worker count": 1}
2024-08-11T07:32:54Z    INFO    controller-runtime.metrics      Serving metrics server  {"bindAddress": ":8443", "secure": true}
2024-08-11T07:33:58Z    INFO    markdownview-resource   default {"name": "markdownview-sample"}
2024-08-11T07:33:58Z    INFO    markdownview-resource   validate create {"name": "markdownview-sample"}
```

## 開発時の流れ

開発時には、カスタムコントローラーの実装を書き換えて何度も動作確認をおこなうことになります。
次のような手順で、効率よく開発を進めることができます。

- コントローラーの実装が変わった場合は、下記のコマンドでコンテナイメージをビルドし、kind環境にロードし直します。
```
$ make docker-build
$ kind load docker-image controller:latest
```

- CRDに変更がある場合は下記のコマンドを実行します。ただし、互換性のない変更をおこなった場合はこのコマンドに失敗するため、事前に`make uninstall`を実行してください。
```
$ make install
```

- CRD以外のマニフェストファイルに変更がある場合は下記のコマンドを実行します。ただし、互換性のない変更をおこなった場合はこのコマンドに失敗するため、事前に`make undeploy`を実行してください。
```
$ make deploy
```

- 次のコマンドでカスタムコントローラーを再起動します。
```
$ kubectl rollout restart -n markdown-view-system deployment markdown-view-controller-manager
```

## Tiltによる効率的な開発

前述したようにカスタムコントローラーの開発時は、ソースコードやマニフェストを変更するたびに複数のmakeコマンドを何度も実行する必要があり、
非常に面倒です。

[Tilt](https://tilt.dev)を利用すると、ソースコードやマニフェストの変更を監視し、
コンテナイメージの再ビルド、Kubernetesクラスタへのマニフェストの適用、Podの再起動などを自動的におこなってくれます。

興味のある方は下記の記事をご覧ください。

- [Tiltでカスタムコントローラーの開発を効率化しよう](https://zenn.dev/zoetro/articles/fba4c77a7fa3fb)

なお、本書のサンプルプログラムではTiltが利用できるようにセットアップしてあります。
詳細は以下のコードをご覧ください。

- https://github.com/zoetrope/kubebuilder-training/tree/main/codes/10_tilt

まず、以下のページを参考に[aqua](https://aquaproj.github.io)をインストールします。

- https://aquaproj.github.io/docs/install/

次にaquaを使って、各種ツールをインストールします。

```console
$ aqua i
```

続いて以下のコマンドを実行して、Kubernetesクラスタとコンテナレジストリを立ち上げ、cert-managerをデプロイします。

```console
$ make start
```

最後にtiltを立ち上げて、ブラウザで http://localhost:10350 にアクセスします。

```console
$ tilt up
```

正常に動作していれば、ソースコードやマニフェストの変更に応じて、kind上のリソースが自動的に更新されるはずです。

終了時には以下のコマンドを実行してください。

```console
$ make stop
```
