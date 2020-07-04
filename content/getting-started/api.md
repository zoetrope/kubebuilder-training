---
title: "Add an API"
draft: true
weight: 13
---

```console
$ kubebuilder create api --group multitenancy --version v1 --kind Tenant
Create Resource [y/n]
y
Create Controller [y/n]
y
$ make manifests
```



```
├── api
│   └── v1
│       ├── groupversion_info.go
│       ├── tenant_types.go
│       └── zz_generated.deepcopy.go
├── config
│   ├── crd
│   │   ├── bases
│   │   │   └── multitenancy.example.com_tenants.yaml
│   │   ├── kustomization.yaml
│   │   ├── kustomizeconfig.yaml
│   │   └── patches
│   │       ├── cainjection_in_tenants.yaml
│   │       └── webhook_in_tenants.yaml
│   ├── rbac
│   │   ├── role.yaml
│   │   ├── role_binding.yaml
│   │   ├── tenant_editor_role.yaml
│   │   └── tenant_viewer_role.yaml
│   └── samples
│      └── multitenancy_v1_tenant.yaml
├── controllers
│   ├── suite_test.go
│   └── tenant_controller.go
└── main.go
```
