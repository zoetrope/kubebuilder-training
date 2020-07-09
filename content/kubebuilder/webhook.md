---
title: "Add a Webhook"
draft: true
weight: 14
---

```console
$ kubebuilder create webhook --group multitenancy --version v1 --kind Tenant --programmatic-validation --defaulting
$ make manifests
```

```
├── api
│   └── v1
│       ├── tenant_webhook.go
│       └── zz_generated.deepcopy.go
├── config
│   └── webhook
│       ├── kustomization.yaml
│       ├── kustomizeconfig.yaml
│       ├── manifests.yaml
│       └── service.yaml
└── main.go
```
