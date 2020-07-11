package controllers

import (
	"context"
	"time"

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
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

func (r *TenantReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tenant", req.NamespacedName)

	// your logic here

	var tenant multitenancyv1.Tenant
	err := r.Get(ctx, req.NamespacedName, &tenant)
	if err != nil {
		log.Error(err, "unable to get tenant", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}

	updated, err := r.reconcile(ctx, log, tenant)
	if err != nil {
		log.Error(err, "unable to reconcile", "name", tenant.Name)
		setCondition(&tenant.Status.Conditions, multitenancyv1.TenantCondition{
			Type:    multitenancyv1.ConditionReady,
			Status:  corev1.ConditionFalse,
			Reason:  "Error",
			Message: err.Error(),
		})
		stErr := r.Status().Update(ctx, &tenant)
		if stErr != nil {
			log.Error(stErr, "failed to update status", "name", tenant.Name)
		}
		return ctrl.Result{}, err
	}

	if updated {
		setCondition(&tenant.Status.Conditions, multitenancyv1.TenantCondition{
			Type:   multitenancyv1.ConditionReady,
			Status: corev1.ConditionTrue,
		})
		err = r.Status().Update(ctx, &tenant)
		if err != nil {
			log.Error(err, "failed to update status", "name", tenant.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *TenantReconciler) reconcile(ctx context.Context, log logr.Logger, tenant multitenancyv1.Tenant) (bool, error) {
	nsUpdated, err := r.reconcileNamespaces(ctx, log, tenant)
	if err != nil {
		return nsUpdated, err
	}
	rbUpdated, err := r.reconcileRBAC(ctx, log, tenant)
	if err != nil {
		return rbUpdated, err
	}
	return nsUpdated || rbUpdated, nil
}

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

func (r *TenantReconciler) reconcileNamespaces(ctx context.Context, log logr.Logger, tenant multitenancyv1.Tenant) (bool, error) {
	var namespaces corev1.NamespaceList
	err := r.List(ctx, &namespaces, client.MatchingFields(map[string]string{namespaceOwnerKey: tenant.Name}))
	if err != nil {
		log.Error(err, "unable to fetch namespaces")
		return false, err
	}
	namespaceNames := make(map[string]corev1.Namespace)
	for _, ns := range namespaces.Items {
		namespaceNames[ns.Name] = ns
	}

	updated := false
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
			return updated, err
		}
		log.Info("creating the new namespace", "name", name)
		err = r.Create(ctx, &target, &client.CreateOptions{})
		if err != nil {
			log.Error(err, "unable to create the namespace", "name", name)
			return updated, err
		}
		updated = true
		delete(namespaceNames, name)
	}

	for _, ns := range namespaceNames {
		log.Info("deleting the new namespace", "name", ns.Name)
		err = r.Delete(ctx, &ns, &client.DeleteOptions{})
		if err != nil {
			log.Error(err, "unable to delete the namespace", "name", ns.Name)
			return updated, err
		}
		updated = true
	}

	return updated, nil
}

func (r *TenantReconciler) reconcileRBAC(ctx context.Context, log logr.Logger, tenant multitenancyv1.Tenant) (bool, error) {
	updated := false
	for _, ns := range tenant.Spec.Namespaces {
		name := tenant.Spec.NamespacePrefix + ns

		role := &rbacv1.ClusterRole{}
		role.SetName(name + "-admin-role")
		op, err := ctrl.CreateOrUpdate(ctx, r.Client, role, func() error {
			role.Rules = []rbacv1.PolicyRule{
				{
					Verbs:         []string{"get", "list", "watch", "update", "patch", "delete"},
					APIGroups:     []string{multitenancyv1.GroupVersion.Group},
					Resources:     []string{"tenants"},
					ResourceNames: []string{tenant.Name},
				},
				{
					Verbs:         []string{"get", "list", "watch"},
					APIGroups:     []string{""},
					Resources:     []string{"namespaces"},
					ResourceNames: []string{name},
				},
			}
			return ctrl.SetControllerReference(&tenant, role, r.Scheme)
		})
		if err != nil {
			log.Error(err, "unable to create-or-update RoleBinding")
			return updated, err
		}

		if op != controllerutil.OperationResultNone {
			updated = true
			log.Info("reconcile RoleBinding successfully", "op", op)
		}

		rb := &rbacv1.RoleBinding{}
		rb.SetNamespace(name)
		rb.SetName(name + "-admin-rolebinding")

		op, err = ctrl.CreateOrUpdate(ctx, r.Client, rb, func() error {
			rb.RoleRef = rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     name + "-admin-role",
			}
			rb.Subjects = []rbacv1.Subject{tenant.Spec.Admin}
			return ctrl.SetControllerReference(&tenant, rb, r.Scheme)
		})
		if err != nil {
			log.Error(err, "unable to create-or-update RoleBinding")
			return updated, err
		}

		if op != controllerutil.OperationResultNone {
			updated = true
			log.Info("reconcile RoleBinding successfully", "op", op)
		}
	}
	return updated, nil
}

func setCondition(conditions *[]multitenancyv1.TenantCondition, newCondition multitenancyv1.TenantCondition) {
	if conditions == nil {
		conditions = &[]multitenancyv1.TenantCondition{}
	}
	current := findCondition(*conditions, newCondition.Type)
	if current == nil {
		newCondition.LastTransitionTime = metav1.NewTime(time.Now())
		*conditions = append(*conditions, newCondition)
		return
	}
	if current.Status != newCondition.Status {
		current.Status = newCondition.Status
		current.LastTransitionTime = metav1.NewTime(time.Now())
	}
	current.Reason = newCondition.Reason
	current.Message = newCondition.Message
}

func findCondition(conditions []multitenancyv1.TenantCondition, conditionType multitenancyv1.TenantConditionType) *multitenancyv1.TenantCondition {
	for _, c := range conditions {
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
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
