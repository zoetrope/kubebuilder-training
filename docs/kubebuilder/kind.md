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
$ kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.16.1/cert-manager.yaml
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
2020-07-03T09:57:11.980Z        DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/mutate-multitenancy-example-com-v1-tenant", "UID": "1bc1074e-a16d-4fe7-a302-6be2b6ded099", "kind": "multitenancy.example.com/v1, Kind=tenant", "resource": {"group":"multitenancy.example.com","version":"v1","resource":"tenant"}}
2020-07-03T09:57:11.981Z        INFO    tenant-resource      default {"name": "tenant-sample"}
2020-07-03T09:57:11.981Z        DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/mutate-multitenancy-example-com-v1-tenant", "UID": "1bc1074e-a16d-4fe7-a302-6be2b6ded099", "allowed": true, "result": {}, "resultError": "got runtime.Object without object metadata: &Status{ListMeta:ListMeta{SelfLink:,ResourceVersion:,Continue:,RemainingItemCount:nil,},Status:,Message:,Reason:,Details:nil,Code:200,}"}
2020-07-03T09:57:11.982Z        DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/validate-multitenancy-example-com-v1-tenant", "UID": "b352235b-e49c-4653-a059-10692137ea1f", "kind": "multitenancy.example.com/v1, Kind=tenant", "resource": {"group":"multitenancy.example.com","version":"v1","resource":"tenant"}}
2020-07-03T09:57:11.982Z        INFO    tenant-resource      validate create {"name": "tenant-sample"}
2020-07-03T09:57:11.982Z        DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/validate-multitenancy-example-com-v1-tenant", "UID": "b352235b-e49c-4653-a059-10692137ea1f", "allowed": true, "result": {}, "resultError": "got runtime.Object without object metadata: &Status{ListMeta:ListMeta{SelfLink:,ResourceVersion:,Continue:,RemainingItemCount:nil,},Status:,Message:,Reason:,Details:nil,Code:200,}"}
2020-07-03T09:57:11.986Z        DEBUG   controller-runtime.controller   Successfully Reconciled {"controller": "tenant", "request": "tenant-sample"}
```
