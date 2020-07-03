---
title: "Add an API"
draft: true
weight: 13
---

```console
$ kubebuilder create api --group webapp --version v1 --kind Guestbook
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
│       ├── guestbook_types.go
│       └── zz_generated.deepcopy.go
├── config
│   ├── crd
│   │   ├── bases
│   │   │   └── webapp.example.com_guestbooks.yaml
│   │   ├── kustomization.yaml
│   │   ├── kustomizeconfig.yaml
│   │   └── patches
│   │       ├── cainjection_in_guestbooks.yaml
│   │       └── webhook_in_guestbooks.yaml
│   ├── rbac
│   │   ├── guestbook_editor_role.yaml
│   │   ├── guestbook_viewer_role.yaml
│   │   ├── role.yaml
│   │   └── role_binding.yaml
│   └── samples
│        └── webapp_v1_guestbook.yaml
├── controllers
│   ├── guestbook_controller.go
│   └── suite_test.go
└── main.go
```
