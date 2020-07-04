---
title: "Running"
draft: true
weight: 15
---

## kindの立ち上げ

```console
$ kind create cluster
```

## cert-managerのインストール

https://cert-manager.io/docs/installation/kubernetes/

```console
$ kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.15.1/cert-manager.yaml
```

## コントローラマネージャの動作確認

コンテナイメージをビルドして、kind環境にロードします。

```console
$ make docker build
$ kind load docker-image controller:latest
```

CRDをKubernetesクラスタに適用します。

```console
$ make install
```

次に`config/manager/manager.yaml`に`imagePullPolicy: IfNotPresent`を追加します。
これはコンテナレジストリからコンテナイメージを取得するのではなく、ローカルのイメージを利用するために必要な設定です。

次にコントローラマネージャのマニフェストを適用します。

```console
$ make deploy
```

コントローラマネージャのPodがRunningになったことを確認してください。

```console
$ kubectl get pod -n sample-system -l control-plane=controller-manager
NAME                                         READY   STATUS    RESTARTS   AGE
sample-controller-manager-6dd494cc9c-vwbzq   1/1     Running   0          1m
```

次にコントローラマネージャのログを表示させておきましょう。

```console
$ kubectl logs -n sample-system -l control-plane=controller-manager -f
```

サンプルのカスタムリソースを適用します。

```console
$ kubectl apply -f config/samples/webapp_v1_guestbook.yaml
```

Podのログに、以下のようにWebhookやReconcileのメッセージが表示されていれば成功です。

```consle
2020-07-03T09:57:11.980Z        DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/mutate-webapp-example-com-v1-guestbook", "UID": "1bc1074e-a16d-4fe7-a302-6be2b6ded099", "kind": "webapp.example.com/v1, Kind=Guestbook", "resource": {"group":"webapp.example.com","version":"v1","resource":"guestbooks"}}
2020-07-03T09:57:11.981Z        INFO    guestbook-resource      default {"name": "guestbook-sample"}
2020-07-03T09:57:11.981Z        DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/mutate-webapp-example-com-v1-guestbook", "UID": "1bc1074e-a16d-4fe7-a302-6be2b6ded099", "allowed": true, "result": {}, "resultError": "got runtime.Object without object metadata: &Status{ListMeta:ListMeta{SelfLink:,ResourceVersion:,Continue:,RemainingItemCount:nil,},Status:,Message:,Reason:,Details:nil,Code:200,}"}
2020-07-03T09:57:11.982Z        DEBUG   controller-runtime.webhook.webhooks     received request        {"webhook": "/validate-webapp-example-com-v1-guestbook", "UID": "b352235b-e49c-4653-a059-10692137ea1f", "kind": "webapp.example.com/v1, Kind=Guestbook", "resource": {"group":"webapp.example.com","version":"v1","resource":"guestbooks"}}
2020-07-03T09:57:11.982Z        INFO    guestbook-resource      validate create {"name": "guestbook-sample"}
2020-07-03T09:57:11.982Z        DEBUG   controller-runtime.webhook.webhooks     wrote response  {"webhook": "/validate-webapp-example-com-v1-guestbook", "UID": "b352235b-e49c-4653-a059-10692137ea1f", "allowed": true, "result": {}, "resultError": "got runtime.Object without object metadata: &Status{ListMeta:ListMeta{SelfLink:,ResourceVersion:,Continue:,RemainingItemCount:nil,},Status:,Message:,Reason:,Details:nil,Code:200,}"}
2020-07-03T09:57:11.986Z        DEBUG   controller-runtime.controller   Successfully Reconciled {"controller": "guestbook", "request": "default/guestbook-sample"}
```
