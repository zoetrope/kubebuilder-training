# Kindで動かしてみよう

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
$ kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.1.0/cert-manager.yaml
```

## コントローラのコンテナイメージの用意

コンテナイメージのタグ名に`latest`を利用すると、`imagePullPolicy: Always`になってしまうため、kind上にロードしたコンテナイメージが利用されない問題があります。([参考](https://kind.sigs.k8s.io/docs/user/quick-start/#loading-an-image-into-your-cluster))
そこでコンテナイメージをビルドする前に、`Makefile`を編集してイメージのタグ名を変更しておきます。

```diff
- IMG ?= controller:latest
+ IMG ?= controller:v1
```

コンテナイメージをビルドします。

```console
$ make docker-build
```

このコンテナイメージを利用するためには、ビルドしたコンテナイメージをDockerHubなどのコンテナレジストリに登録するか、kind環境にロードする必要があります。
ここでは下記のコマンドを利用してkind環境にコンテナイメージをロードしましょう。

```console
$ kind load docker-image controller:v1
```


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
$ kubectl get pod -n tenant-system
NAME                                         READY   STATUS    RESTARTS   AGE
tenant-controller-manager-6dd494cc9c-vwbzq   1/1     Running   0          1m
```

次にコントローラのログを表示させておきます。

```console
$ kubectl logs -n tenant-system tenant-controller-manager-6dd494cc9c-vwbzq -c manager -f
```

サンプルのカスタムリソースを適用します。

```console
$ kubectl apply -f config/samples/multitenancy_v1_tenant.yaml
```

以下のようにWebhookやReconcileのメッセージがコントローラのログに表示されていれば成功です。

```console
2020-12-12T08:40:34.435Z        INFO    controller-runtime.metrics      metrics server is starting to listen    {"addr": "127.0.0.1:8080"}
2020-12-12T08:40:34.435Z        INFO    controller-runtime.builder      Registering a mutating webhook  {"GVK": "multitenancy.example.com/v1, Kind=Tenant", "path": "/mutate-multitenancy-example-com-v1-tenant"}
2020-12-12T08:40:34.435Z        INFO    controller-runtime.webhook      registering webhook     {"path": "/mutate-multitenancy-example-com-v1-tenant"}
2020-12-12T08:40:34.435Z        INFO    controller-runtime.builder      Registering a validating webhook        {"GVK": "multitenancy.example.com/v1, Kind=Tenant", "path": "/validate-multitenancy-example-com-v1-tenant"}
2020-12-12T08:40:34.436Z        INFO    controller-runtime.webhook      registering webhook     {"path": "/validate-multitenancy-example-com-v1-tenant"}
2020-12-12T08:40:34.436Z        INFO    setup   starting manager
I1212 08:40:34.436378       1 leaderelection.go:243] attempting to acquire leader lease  tenant-system/27475f02.example.com...
2020-12-12T08:40:34.436Z        INFO    controller-runtime.manager      starting metrics server {"path": "/metrics"}
2020-12-12T08:40:34.536Z        INFO    controller-runtime.webhook.webhooks     starting webhook server
2020-12-12T08:40:34.536Z        INFO    controller-runtime.certwatcher  Updated current TLS certificate
2020-12-12T08:40:34.536Z        INFO    controller-runtime.webhook      serving webhook server  {"host": "", "port": 9443}
2020-12-12T08:40:34.537Z        INFO    controller-runtime.certwatcher  Starting certificate watcher
2020-12-12T08:40:44.991Z        DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/mutate-multitenancy-example-com-v1-tenant", "UID": "cc5f97e6-dd4c-460e-8790-1d8163ba2f94", "kind": "multit
enancy.example.com/v1, Kind=Tenant", "resource": {"group":"multitenancy.example.com","version":"v1","resource":"tenants"}}
2020-12-12T08:40:44.991Z        INFO    tenant-resource default {"name": "tenant-sample"}
2020-12-12T08:40:44.992Z        DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/mutate-multitenancy-example-com-v1-tenant", "code": 200, "reason": "", "UID": "cc5f97e6-dd4c-460e-8790-1d8163ba2f9
4", "allowed": true}
2020-12-12T08:40:44.997Z        DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/validate-multitenancy-example-com-v1-tenant", "UID": "09cb9d18-9d22-45b9-bdca-75065db20c6c", "kind": "mult
itenancy.example.com/v1, Kind=Tenant", "resource": {"group":"multitenancy.example.com","version":"v1","resource":"tenants"}}
2020-12-12T08:40:44.997Z        INFO    tenant-resource validate update {"name": "tenant-sample"}
2020-12-12T08:40:44.997Z        DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/validate-multitenancy-example-com-v1-tenant", "code": 200, "reason": "", "UID": "09cb9d18-9d22-45b9-bdca-75065db20
c6c", "allowed": true}
```
