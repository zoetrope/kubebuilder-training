# カスタムコントローラの動作確認

kubebuilderコマンドで生成したプロジェクトをビルドし、[Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)環境で動かしてみましょう。

Kindとはローカル環境にKubernetesクラスタを構築するためのツールで、手軽にコントローラのテストや動作確認をおこなうことができます。

## kindの立ち上げ

まずはkindコマンドを利用してKubernetesクラスタを作成します。

```console
$ kind create cluster
```

## cert-managerのインストール

Webhook用の証明書を発行するためにcert-managerが必要となります。
下記のコマンドを実行してcert-managerのデプロイをおこないます。([参考](https://cert-manager.io/docs/installation/kubernetes/))

```console
$ kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
```

cert-managerのPodが起動したことを確認しましょう。

```console
$ kubectl get pod -n cert-manager
NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager-7dd5854bb4-whlcn              1/1     Running   0          26s
cert-manager-cainjector-64c949654c-64wjk   1/1     Running   0          26s
cert-manager-webhook-6bdffc7c9d-hkr8h      1/1     Running   0          26s
```

## コントローラのコンテナイメージの用意

コンテナイメージをビルドします。

```console
$ make docker-build
```

このコンテナイメージを利用するためには、ビルドしたコンテナイメージをDockerHubなどのコンテナレジストリに登録するか、kind環境にロードする必要があります。

ここでは下記のコマンドを利用してkind環境にコンテナイメージをロードしましょう。

```console
$ kind load docker-image controller:latest
```

なおkind環境では、コンテナイメージのタグ名に`latest`を利用するとロードしたコンテナイメージが利用されない問題があります。([参考](https://kind.sigs.k8s.io/docs/user/quick-start/#loading-an-image-into-your-cluster))

そこで、[config/manager/manager.yaml](https://github.com/zoetrope/kubebuilder-training/blob/master/codes/markdown-viewer/config/manager/manager.yaml)に`imagePullPolicy: IfNotPresent`を追加しておきます。

## コントローラの動作確認

次にCRDをKubernetesクラスタに適用します。

```console
$ make install
```

続いて各種マニフェストを適用します。

```console
$ make deploy
```

コントローラのPodがRunningになったことを確認してください。

```console
$ kubectl get pod -n markdown-viewer-system
NAME                                                  READY   STATUS    RESTARTS   AGE
markdown-viewer-controller-manager-5bc678bbf9-vb9r5   2/2     Running   0          30s
```

次にコントローラのログを表示させておきます。

```console
$ kubectl logs -n markdown-viewer-system markdown-viewer-controller-manager-5bc678bbf9-vb9r5 -c manager -f
```

サンプルのカスタムリソースを適用します。

```console
$ kubectl apply -f config/samples/viewer_v1_markdownview.yaml
```

以下のようにWebhookやReconcileのメッセージがコントローラのログに表示されていれば成功です。

```console
2021-07-10T09:29:49.311Z        INFO    controller-runtime.metrics      metrics server is starting to listen     {"addr": "127.0.0.1:8080"}
2021-07-10T09:29:49.311Z        INFO    controller-runtime.builder      Registering a mutating webhook   {"GVK": "viewer.zoetrope.github.io/v1, Kind=MarkdownView", "path": "/mutate-viewer-zoetrope-github-io-v1-markdownview"}
2021-07-10T09:29:49.311Z        INFO    controller-runtime.webhook      registering webhook      {"path": "/mutate-viewer-zoetrope-github-io-v1-markdownview"}
2021-07-10T09:29:49.311Z        INFO    controller-runtime.builder      Registering a validating webhook {"GVK": "viewer.zoetrope.github.io/v1, Kind=MarkdownView", "path": "/validate-viewer-zoetrope-github-io-v1-markdownview"}
2021-07-10T09:29:49.311Z        INFO    controller-runtime.webhook      registering webhook      {"path": "/validate-viewer-zoetrope-github-io-v1-markdownview"}
2021-07-10T09:29:49.311Z        INFO    setup   starting manager
I0710 09:29:49.312373       1 leaderelection.go:243] attempting to acquire leader lease markdown-viewer-system/c124e721.zoetrope.github.io...
2021-07-10T09:29:49.312Z        INFO    controller-runtime.manager      starting metrics server  {"path": "/metrics"}
2021-07-10T09:29:49.312Z        INFO    controller-runtime.webhook.webhooks     starting webhook server
2021-07-10T09:29:49.312Z        INFO    controller-runtime.certwatcher  Updated current TLS certificate
2021-07-10T09:29:49.312Z        INFO    controller-runtime.webhook      serving webhook server   {"host": "", "port": 9443}
2021-07-10T09:29:49.312Z        INFO    controller-runtime.certwatcher  Starting certificate watcher
I0710 09:29:49.409787       1 leaderelection.go:253] successfully acquired lease markdown-viewer-system/c124e721.zoetrope.github.io
2021-07-10T09:29:49.409Z        DEBUG   controller-runtime.manager.events       Normal  {"object": {"kind":"ConfigMap","namespace":"markdown-viewer-system","name":"c124e721.zoetrope.github.io","uid":"b48865ea-3d05-47bd-be4f-4d03a14b7a36","apiVersion":"v1","resourceVersion":"1982"}, "reason": "LeaderElection", "message": "markdown-viewer-controller-manager-5bc678bbf9-vb9r5_d64b0043-4a95-432e-9c76-3001247a87ac became leader"}
2021-07-10T09:29:49.409Z        DEBUG   controller-runtime.manager.events       Normal  {"object": {"kind":"Lease","namespace":"markdown-viewer-system","name":"c124e721.zoetrope.github.io","uid":"3ef3dcde-abbb-440b-9052-1c85ed01d67d","apiVersion":"coordination.k8s.io/v1","resourceVersion":"1983"}, "reason": "LeaderElection", "message": "markdown-viewer-controller-manager-5bc678bbf9-vb9r5_d64b0043-4a95-432e-9c76-3001247a87ac became leader"}
2021-07-10T09:29:49.410Z        INFO    controller-runtime.manager.controller.markdownview       Starting EventSource    {"reconciler group": "viewer.zoetrope.github.io", "reconciler kind": "MarkdownView", "source": "kind source: /, Kind="}
2021-07-10T09:29:49.410Z        INFO    controller-runtime.manager.controller.markdownview       Starting Controller     {"reconciler group": "viewer.zoetrope.github.io", "reconciler kind": "MarkdownView"}
2021-07-10T09:29:49.511Z        INFO    controller-runtime.manager.controller.markdownview       Starting workers        {"reconciler group": "viewer.zoetrope.github.io", "reconciler kind": "MarkdownView", "worker count": 1}
2021-07-10T09:33:53.622Z        DEBUG   controller-runtime.webhook.webhooks     received request {"webhook": "/mutate-viewer-zoetrope-github-io-v1-markdownview", "UID": "20fe30b5-6d45-4592-ae4b-ee5048e054d1", "kind": "viewer.zoetrope.github.io/v1, Kind=MarkdownView", "resource": {"group":"viewer.zoetrope.github.io","version":"v1","resource":"markdownviews"}}
2021-07-10T09:33:53.623Z        INFO    markdownview-resource   default {"name": "markdownview-sample"}
2021-07-10T09:33:53.623Z        DEBUG   controller-runtime.webhook.webhooks     wrote response   {"webhook": "/mutate-viewer-zoetrope-github-io-v1-markdownview", "code": 200, "reason": "", "UID": "20fe30b5-6d45-4592-ae4b-ee5048e054d1", "allowed": true}
2021-07-10T09:33:53.626Z        DEBUG   controller-runtime.webhook.webhooks     received request {"webhook": "/validate-viewer-zoetrope-github-io-v1-markdownview", "UID": "904fc35e-4415-4a90-af96-52cbe1cef1b7", "kind": "viewer.zoetrope.github.io/v1, Kind=MarkdownView", "resource": {"group":"viewer.zoetrope.github.io","version":"v1","resource":"markdownviews"}}
2021-07-10T09:33:53.626Z        INFO    markdownview-resource   validate create {"name": "markdownview-sample"}
2021-07-10T09:33:53.626Z        DEBUG   controller-runtime.webhook.webhooks     wrote response   {"webhook": "/validate-viewer-zoetrope-github-io-v1-markdownview", "code": 200, "reason": "", "UID": "904fc35e-4415-4a90-af96-52cbe1cef1b7", "allowed": true}
```

## 開発時の流れ

開発時には、カスタムコントローラの実装を書き換えて何度も動作確認をおこなうことになります。
次のような手順で、効率よく開発を進めることができます。

- コントローラの実装が変わった場合は、下記のコマンドでコンテナイメージをビルドし、kind環境にロードし直します。
```
$ make docker-build
$ kind load docker-image controller:latest
```

- CRDに変更がある場合は下記のコマンドを実行します。ただし、互換性のない変更をおこなった場合はこのコマンドに失敗するため、事前に`make uninstall`を実行してください。
```
make install
```

- CRD以外のマニフェストファイルに変更がある場合は下記のコマンドを実行します。ただし、互換性のない変更をおこなった場合はこのコマンドに失敗するため、事前に`make undeploy`を実行してください。
```
make deploy
```

- 次のコマンドでカスタムコントローラを再起動します。
```
$ kubectl rollout restart -n markdown-viewer-system deployment markdown-viewer-controller-manager
```
