package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	multitenancyv1 "github.com/zoetrope/kubebuilder-training/static/codes/api/v1"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=multitenancy.example.com,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=multitenancy.example.com,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

func (r *TenantReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tenant", req.NamespacedName)

	// your logic here

	var tenant multitenancyv1.Tenant
	err := r.Get(ctx, req.NamespacedName, &tenant)
	if err != nil {
		log.Error(err, "unable to get tenant")
		return ctrl.Result{}, err
	}
	err = r.reconcileNamespaces(ctx, log, tenant)
	if err != nil {
		log.Error(err, "unable to reconcile")
		return ctrl.Result{}, err
	}

	err = r.reconcileRoleBindings(ctx, log, tenant)
	if err != nil {
		log.Error(err, "unable to reconcile")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

const namespaceOwnerKey = ".metadata.ownerReference.controller"

func predicate(obj runtime.Object) []string {
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

func (r *TenantReconciler) reconcileNamespaces(ctx context.Context, log logr.Logger, tenant multitenancyv1.Tenant) error {
	var namespaces corev1.NamespaceList
	err := r.List(ctx, &namespaces, client.MatchingFields(map[string]string{namespaceOwnerKey: tenant.Name}))
	if err != nil {
		log.Error(err, "unable to fetch namespaces")
		return err
	}
	namespaceNames := make(map[string]corev1.Namespace)
	for _, ns := range namespaces.Items {
		namespaceNames[ns.Name] = ns
	}

	for _, ns := range tenant.Spec.Namespaces {
		name := tenant.Spec.NamespacePrefix + ns
		if _, ok := namespaceNames[name]; ok {
			delete(namespaceNames, name)
			continue
		}
		target := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		err = ctrl.SetControllerReference(&tenant, &target, r.Scheme)
		if err != nil {
			log.Error(err, "unable to set owner reference", "name", name)
			return err
		}
		log.Info("creating the new namespace", "name", name)
		err = r.Create(ctx, &target, &client.CreateOptions{})
		if err != nil {
			log.Error(err, "unable to create the namespace", "name", name)
			return err
		}
		delete(namespaceNames, name)
	}

	for _, ns := range namespaceNames {
		log.Info("deleting the new namespace", "name", ns.Name)
		err = r.Delete(ctx, &ns, &client.DeleteOptions{})
		if err != nil {
			log.Error(err, "unable to delete the namespace", "name", ns.Name)
			return err
		}
	}

	return nil
}

func (r *TenantReconciler) reconcileRoleBindings(ctx context.Context, log logr.Logger, tenant multitenancyv1.Tenant) error {
	for _, ns := range tenant.Spec.Namespaces {
		name := tenant.Spec.NamespacePrefix + ns

		rb := &rbacv1.RoleBinding{}
		rb.SetNamespace(name)
		rb.SetName(name + "-admin")

		op, err := ctrl.CreateOrUpdate(ctx, r.Client, rb, func() error {
			rb.RoleRef = rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "admin",
			}
			rb.Subjects = []rbacv1.Subject{tenant.Spec.Admin}
			return ctrl.SetControllerReference(&tenant, rb, r.Scheme)
		})
		if err != nil {
			log.Error(err, "unable to create-or-update RoleBinding")
			return err
		}

		if op != controllerutil.OperationResultNone {
			log.Info("reconcile RoleBinding successfully", "op", op)
		}
	}
	return nil
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(&corev1.Namespace{}, namespaceOwnerKey, predicate)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&multitenancyv1.Tenant{}).
		Owns(&corev1.Namespace{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}
