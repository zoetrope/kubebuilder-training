---
title: "Add a Webhook"
draft: true
weight: 14
---

```console
$ kubebuilder create webhook --group webapp --version v1 --kind Guestbook --programmatic-validation --defaulting
$ make manifests
```

```
├── api
│   └── v1
│       ├── groupversion_info.go
│       ├── guestbook_types.go
│       ├── guestbook_webhook.go
│       └── zz_generated.deepcopy.go
├── config
│   └── webhook
│       ├── kustomization.yaml
│       ├── kustomizeconfig.yaml
│       ├── manifests.yaml
│       └── service.yaml
└── main.go
```
