---
title: "Client"
draft: true
weight: 31
---

## Clientの作成

```go
import (
	multitenancyv1 "github.com/zoetrope/kubebuilder-training/static/codes/api/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme   = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = multitenancyv1.AddToScheme(scheme)
}

func main() {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		return
	}

	client := mgr.GetClient()
	reader := mgr.GetAPIReader()
}
```

## Get

## List

```go
const namespaceOwnerKey = ".metadata.ownerReference.controller"

func selectOwnedNamespaces(obj runtime.Object) []string {
	namespace := obj.(*corev1.Namespace)
	owner := metav1.GetControllerOf(namespace)
	if owner == nil {
		return nil
	}
	if owner.APIVersion != multitenancyv1.GroupVersion.String() || owner.Kind != "Tenant" {
		return nil
	}
	return []string{owner.Name}
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(&corev1.Namespace{}, namespaceOwnerKey, selectOwnedNamespaces)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&multitenancyv1.Tenant{}).
		Owns(&corev1.Namespace{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}
```

```go
	var namespaces corev1.NamespaceList
	err := r.List(ctx, &namespaces, client.MatchingFields(map[string]string{namespaceOwnerKey: tenant.Name}))
	if err != nil {
		log.Error(err, "unable to fetch namespaces")
		return false, err
	}
```

## Create/Update

```go
		target := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		log.Info("creating the new namespace", "name", name)
		err = r.Create(ctx, &target, &client.CreateOptions{})
		if err != nil {
			log.Error(err, "unable to create the namespace", "name", name)
			return updated, err
		}
```

## CreateOrUpdate

## Patch

## Status.Update/Patch

## Delete/DeleteOfAll

