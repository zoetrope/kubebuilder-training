/*
Copyright 2021.

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

package controllers

import (
	"context"

	viewerv1 "github.com/zoetrope/markdown-viewer/api/v1"
	"github.com/zoetrope/markdown-viewer/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MarkdownViewReconciler reconciles a MarkdownView object
type MarkdownViewReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//! [rbac]
//+kubebuilder:rbac:groups=viewer.zoetrope.github.io,resources=markdownviews,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=viewer.zoetrope.github.io,resources=markdownviews/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=viewer.zoetrope.github.io,resources=markdownviews/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//! [rbac]

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MarkdownView object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *MarkdownViewReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	//! [get]
	var mdView viewerv1.MarkdownView
	err := r.Get(ctx, req.NamespacedName, &mdView)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "unable to get MarkdownView", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}
	//! [get]

	if !mdView.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	_, err = r.reconcileConfigMap(ctx, mdView)
	if err != nil {
		return ctrl.Result{}, err
	}
	_, err = r.reconcileDeployment(ctx, mdView)
	if err != nil {
		return ctrl.Result{}, err
	}
	_, err = r.reconcileService(ctx, mdView)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MarkdownViewReconciler) reconcileConfigMap(ctx context.Context, mdView viewerv1.MarkdownView) (bool, error) {
	logger := log.FromContext(ctx)

	cm := &corev1.ConfigMap{}
	cm.SetNamespace(mdView.Namespace)
	cm.SetName("markdowns-" + mdView.Name)

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		for name, content := range mdView.Spec.Markdowns {
			cm.Data[name] = content
		}
		return ctrl.SetControllerReference(&mdView, cm, r.Scheme)
	})

	if err != nil {
		logger.Error(err, "unable to create or update ConfigMap")
		return false, err
	}
	if op != controllerutil.OperationResultNone {
		logger.Info("reconcile ConfigMap successfully", "op", op)
		return true, nil
	}
	return false, nil
}

func (r *MarkdownViewReconciler) reconcileDeployment(ctx context.Context, mdView viewerv1.MarkdownView) (bool, error) {
	depName := "viewer-" + mdView.Name
	viewerImage := constants.DefaultViewerImage
	if len(mdView.Spec.ViewerImage) != 0 {
		viewerImage = mdView.Spec.ViewerImage
	}

	owner, err := ownerRef(mdView, r.Scheme)
	if err != nil {
		return false, err
	}

	dep := appsv1apply.Deployment(depName, mdView.Namespace).
		WithLabels(labelSet(mdView)).
		WithOwnerReferences(owner).
		WithSpec(appsv1apply.DeploymentSpec().
			WithReplicas(mdView.Spec.Replicas).
			WithSelector(metav1apply.LabelSelector().WithMatchLabels(labelSet(mdView))).
			WithTemplate(corev1apply.PodTemplateSpec().
				WithLabels(labelSet(mdView)).
				WithSpec(corev1apply.PodSpec().
					WithContainers(corev1apply.Container().
						WithName(constants.ViewerName).
						WithImage(viewerImage).
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommand("mdbook").
						WithArgs("serve", "--hostname", "0.0.0.0").
						WithVolumeMounts(corev1apply.VolumeMount().
							WithName("markdowns").
							WithMountPath("/book/src"),
						).
						WithPorts(corev1apply.ContainerPort().
							WithName("http").
							WithProtocol(corev1.ProtocolTCP).
							WithContainerPort(3000),
						),
					).
					WithVolumes(corev1apply.Volume().
						WithName("markdowns").
						WithConfigMap(corev1apply.ConfigMapVolumeSource().
							WithName("markdowns-" + mdView.Name),
						),
					),
				),
			),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(dep)
	if err != nil {
		return false, err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current appsv1.Deployment
	err = r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: depName}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}

	currApplyConfig, err := appsv1apply.ExtractDeployment(&current, constants.ControllerName)
	if err != nil {
		return false, err
	}

	if equality.Semantic.DeepEqual(dep, currApplyConfig) {
		return false, nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: constants.ControllerName,
	})
	return true, err
}

func (r *MarkdownViewReconciler) reconcileService(ctx context.Context, mdView viewerv1.MarkdownView) (bool, error) {
	svcName := "viewer-" + mdView.Name

	owner, err := ownerRef(mdView, r.Scheme)
	if err != nil {
		return false, err
	}

	svc := corev1apply.Service(svcName, mdView.Namespace).
		WithLabels(labelSet(mdView)).
		WithOwnerReferences(owner).
		WithSpec(corev1apply.ServiceSpec().
			WithSelector(labelSet(mdView)).
			WithPorts(corev1apply.ServicePort().
				WithProtocol(corev1.ProtocolTCP).
				WithPort(80).
				WithTargetPort(intstr.FromInt(3000)),
			),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(svc)
	if err != nil {
		return false, err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current corev1.Service
	err = r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: svcName}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}

	currApplyConfig, err := corev1apply.ExtractService(&current, constants.ControllerName)
	if err != nil {
		return false, err
	}

	if equality.Semantic.DeepEqual(svc, currApplyConfig) {
		return false, nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: constants.ControllerName,
	})
	return true, err
}

func ownerRef(mdView viewerv1.MarkdownView, scheme *runtime.Scheme) (*metav1apply.OwnerReferenceApplyConfiguration, error) {
	gvk, err := apiutil.GVKForObject(&mdView, scheme)
	if err != nil {
		return nil, err
	}
	ref := metav1apply.OwnerReference().
		WithAPIVersion(gvk.GroupVersion().String()).
		WithKind(gvk.Kind).
		WithName(mdView.Name).
		WithUID(mdView.GetUID()).
		WithBlockOwnerDeletion(true).
		WithController(true)
	return ref, nil
}

func labelSet(mdView viewerv1.MarkdownView) map[string]string {
	labels := map[string]string{
		constants.LabelAppName:      constants.ViewerName,
		constants.LabelAppInstance:  mdView.Name,
		constants.LabelAppCreatedBy: constants.ControllerName,
	}
	return labels
}

// SetupWithManager sets up the controller with the Manager.
func (r *MarkdownViewReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//! [managedby]
	return ctrl.NewControllerManagedBy(mgr).
		For(&viewerv1.MarkdownView{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
	//! [managedby]
}
