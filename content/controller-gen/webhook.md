---
title: "Webhook"
draft: false
weight: 23
---

[api/v1/tenant_webhook.go](https://github.com/zoetrope/kubebuilder-training/blob/master/static/codes/tenant/api/v1/tenant_webhook.go)

```go
// +kubebuilder:webhook:path=/mutate-multitenancy-example-com-v1-tenant,mutating=true,failurePolicy=fail,groups=multitenancy.example.com,resources=tenants,verbs=create,versions=v1,name=mtenant.kb.io

var _ webhook.Defaulter = &Tenant{}
```

```go
// +kubebuilder:webhook:verbs=update,path=/validate-multitenancy-example-com-v1-tenant,mutating=false,failurePolicy=fail,groups=multitenancy.example.com,resources=tenants,versions=v1,name=vtenant.kb.io

var _ webhook.Validator = &Tenant{}
```
