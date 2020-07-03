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

## 

```console
$ make docker build
$ kind load docker-image controller:latest
```

```console
$ make install
```
