/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

//! [import]
import (
	"context"
	"fmt"

	viewv1 "github.com/zoetrope/markdown-view/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyappsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applycorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applymetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//! [import]

//! [reconciler]

// MarkdownViewReconciler reconciles a MarkdownView object
type MarkdownViewReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//! [reconciler]

//! [rbac]
// +kubebuilder:rbac:groups=view.zoetrope.github.io,resources=markdownviews,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=view.zoetrope.github.io,resources=markdownviews/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=view.zoetrope.github.io,resources=markdownviews/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch
//! [rbac]

//! [reconcile]

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MarkdownView object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *MarkdownViewReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	res, err := r.Reconcile_create(ctx, req)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return res, err
	}

	res, err = r.Reconcile_createOrUpdate(ctx, req)
	if err != nil {
		return res, err
	}

RETRY:
	res, err = r.Reconcile_get(ctx, req)
	if apierrors.IsNotFound(err) {
		goto RETRY
	}
	if err != nil {
		return res, err
	}

	res, err = r.Reconcile_list(ctx, req)
	if err != nil {
		return res, err
	}

	res, err = r.Reconcile_pagination(ctx, req)
	if err != nil {
		return res, err
	}

	res, err = r.Reconcile_patchMerge(ctx, req)
	if err != nil {
		return res, err
	}

	res, err = r.Reconcile_patchApply(ctx, req)
	if err != nil {
		return res, err
	}

	res, err = r.Reconcile_patchApplyConfig(ctx, req)
	if err != nil {
		return res, err
	}

	res, err = r.Reconcile_deleteWithPreConditions(ctx, req)
	if err != nil {
		return res, err
	}

	res, err = r.Reconcile_deleteAllOfDeployment(ctx, req)
	if err != nil {
		return res, err
	}

	return res, err
}

//! [reconcile]

//! [managedby]

// SetupWithManager sets up the controller with the Manager.
func (r *MarkdownViewReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&viewv1.MarkdownView{}).
		Complete(r)
}

//! [managedby]

//! [create]

func (r *MarkdownViewReconciler) Reconcile_create(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	dep := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "nginx"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "nginx"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
	err := r.Create(ctx, &dep)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

//! [create]

//! [create-or-update]

func (r *MarkdownViewReconciler) Reconcile_createOrUpdate(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	svc := &corev1.Service{}
	svc.SetNamespace("default")
	svc.SetName("sample")

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = map[string]string{"app": "nginx"}
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(80),
			},
		}
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		fmt.Printf("Service %s\n", op)
	}
	return ctrl.Result{}, nil
}

//! [create-or-update]

//! [get]

func (r *MarkdownViewReconciler) Reconcile_get(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var deployment appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &deployment)
	if err != nil {
		return ctrl.Result{}, err
	}
	fmt.Printf("Got Deployment: %#v\n", deployment)
	return ctrl.Result{}, nil
}

//! [get]

//! [list]

func (r *MarkdownViewReconciler) Reconcile_list(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var services corev1.ServiceList
	err := r.List(ctx, &services, &client.ListOptions{
		Namespace:     "default",
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": "sample"}),
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, svc := range services.Items {
		fmt.Println(svc.Name)
	}
	return ctrl.Result{}, nil
}

//! [list]

//! [pagination]

func (r *MarkdownViewReconciler) Reconcile_pagination(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	token := ""
	for i := 0; ; i++ {
		var services corev1.ServiceList
		err := r.List(ctx, &services, &client.ListOptions{
			Limit:    3,
			Continue: token,
		})
		if err != nil {
			return ctrl.Result{}, err
		}

		fmt.Printf("Page %d:\n", i)
		for _, svc := range services.Items {
			fmt.Println(svc.Name)
		}
		fmt.Println()

		token = services.ListMeta.Continue
		if len(token) == 0 {
			return ctrl.Result{}, nil
		}
	}
}

//! [pagination]

//! [cond]

func (r *MarkdownViewReconciler) Reconcile_deleteWithPreConditions(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var deploy appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &deploy)
	if err != nil {
		return ctrl.Result{}, err
	}
	uid := deploy.GetUID()
	resourceVersion := deploy.GetResourceVersion()
	cond := metav1.Preconditions{
		UID:             &uid,
		ResourceVersion: &resourceVersion,
	}
	err = r.Delete(ctx, &deploy, &client.DeleteOptions{
		Preconditions: &cond,
	})
	return ctrl.Result{}, err
}

//! [cond]

//! [delete-all-of]

func (r *MarkdownViewReconciler) Reconcile_deleteAllOfDeployment(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	err := r.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace("default"))
	return ctrl.Result{}, err
}

//! [delete-all-of]

//! [update-status]

func (r *MarkdownViewReconciler) updateStatus(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var dep appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &dep)
	if err != nil {
		return ctrl.Result{}, err
	}

	dep.Status.AvailableReplicas = 3
	err = r.Status().Update(ctx, &dep)
	return ctrl.Result{}, err
}

//! [update-status]

//! [patch-merge]

func (r *MarkdownViewReconciler) Reconcile_patchMerge(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var dep appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &dep)
	if err != nil {
		return ctrl.Result{}, err
	}

	newDep := dep.DeepCopy()
	newDep.Spec.Replicas = pointer.Int32Ptr(3)
	patch := client.MergeFrom(&dep)

	err = r.Patch(ctx, newDep, patch)

	return ctrl.Result{}, err
}

//! [patch-merge]

//! [patch-apply]

func (r *MarkdownViewReconciler) Reconcile_patchApply(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	patch := &unstructured.Unstructured{}
	patch.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	patch.SetNamespace("default")
	patch.SetName("sample2")
	patch.UnstructuredContent()["spec"] = map[string]interface{}{
		"replicas": 2,
		"selector": map[string]interface{}{
			"matchLabels": map[string]string{
				"app": "nginx",
			},
		},
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]string{
					"app": "nginx",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "nginx",
						"image": "nginx:latest",
					},
				},
			},
		},
	}

	err := r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "client-sample",
		Force:        pointer.Bool(true),
	})

	return ctrl.Result{}, err
}

//! [patch-apply]

//! [patch-apply-config]

func (r *MarkdownViewReconciler) Reconcile_patchApplyConfig(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	dep := applyappsv1.Deployment("sample3", "default").
		WithSpec(applyappsv1.DeploymentSpec().
			WithReplicas(3).
			WithSelector(applymetav1.LabelSelector().WithMatchLabels(map[string]string{"app": "nginx"})).
			WithTemplate(applycorev1.PodTemplateSpec().
				WithLabels(map[string]string{"app": "nginx"}).
				WithSpec(applycorev1.PodSpec().
					WithContainers(applycorev1.Container().
						WithName("nginx").
						WithImage("nginx:latest"),
					),
				),
			),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(dep)
	if err != nil {
		return ctrl.Result{}, err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current appsv1.Deployment
	err = r.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample3"}, &current)
	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	currApplyConfig, err := applyappsv1.ExtractDeployment(&current, "client-sample")
	if err != nil {
		return ctrl.Result{}, err
	}

	if equality.Semantic.DeepEqual(dep, currApplyConfig) {
		return ctrl.Result{}, nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "client-sample",
		Force:        pointer.Bool(true),
	})
	return ctrl.Result{}, err
}

//! [patch-apply-config]
