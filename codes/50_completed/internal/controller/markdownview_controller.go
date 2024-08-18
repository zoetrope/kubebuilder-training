/*
Copyright 2024.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

//! [import]

//! [reconciler]

// MarkdownViewReconciler reconciles a MarkdownView object
type MarkdownViewReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//! [reconciler]

//! [rbac]
// +kubebuilder:rbac:groups=view.zoetrope.github.io,resources=markdownviews,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=view.zoetrope.github.io,resources=markdownviews/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=view.zoetrope.github.io,resources=markdownviews/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *MarkdownViewReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("start reconcile MarkdownView")
	//! [call-remove-metrics]
	var mdView viewv1.MarkdownView
	err := r.Get(ctx, req.NamespacedName, &mdView)
	if errors.IsNotFound(err) {
		r.removeMetrics(mdView)
		return ctrl.Result{}, nil
	}
	//! [call-remove-metrics]
	if err != nil {
		logger.Error(err, "unable to get MarkdownView", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if !mdView.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	err = r.reconcileConfigMap(ctx, mdView)
	if err != nil {
		result, err2 := r.updateStatus(ctx, mdView)
		logger.Error(err2, "unable to update status")
		return result, err
	}
	err = r.reconcileDeployment(ctx, mdView)
	if err != nil {
		result, err2 := r.updateStatus(ctx, mdView)
		logger.Error(err2, "unable to update status")
		return result, err
	}
	err = r.reconcileService(ctx, mdView)
	if err != nil {
		result, err2 := r.updateStatus(ctx, mdView)
		logger.Error(err2, "unable to update status")
		return result, err
	}

	return r.updateStatus(ctx, mdView)
}

//! [reconcile]

//! [reconcile-configmap]

func (r *MarkdownViewReconciler) reconcileConfigMap(ctx context.Context, mdView viewv1.MarkdownView) error {
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
		return err
	}
	if op != controllerutil.OperationResultNone {
		logger.Info("reconcile ConfigMap successfully", "op", op)
	}
	return nil
}

//! [reconcile-configmap]

//! [reconcile-deployment]

func (r *MarkdownViewReconciler) reconcileDeployment(ctx context.Context, mdView viewv1.MarkdownView) error {
	logger := log.FromContext(ctx)

	depName := "viewer-" + mdView.Name
	viewerImage := "peaceiris/mdbook:latest"
	if len(mdView.Spec.ViewerImage) != 0 {
		viewerImage = mdView.Spec.ViewerImage
	}

	owner, err := controllerReference(mdView, r.Scheme)
	if err != nil {
		return err
	}

	dep := appsv1apply.Deployment(depName, mdView.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "mdbook",
			"app.kubernetes.io/instance":   mdView.Name,
			"app.kubernetes.io/created-by": "markdown-view-controller",
		}).
		WithOwnerReferences(owner).
		WithSpec(appsv1apply.DeploymentSpec().
			WithReplicas(mdView.Spec.Replicas).
			WithSelector(metav1apply.LabelSelector().WithMatchLabels(map[string]string{
				"app.kubernetes.io/name":       "mdbook",
				"app.kubernetes.io/instance":   mdView.Name,
				"app.kubernetes.io/created-by": "markdown-view-controller",
			})).
			WithTemplate(corev1apply.PodTemplateSpec().
				WithLabels(map[string]string{
					"app.kubernetes.io/name":       "mdbook",
					"app.kubernetes.io/instance":   mdView.Name,
					"app.kubernetes.io/created-by": "markdown-view-controller",
				}).
				WithSpec(corev1apply.PodSpec().
					WithContainers(corev1apply.Container().
						WithName("mdbook").
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
						).
						WithLivenessProbe(corev1apply.Probe().
							WithHTTPGet(corev1apply.HTTPGetAction().
								WithPort(intstr.FromString("http")).
								WithPath("/").
								WithScheme(corev1.URISchemeHTTP),
							),
						).
						WithReadinessProbe(corev1apply.Probe().
							WithHTTPGet(corev1apply.HTTPGetAction().
								WithPort(intstr.FromString("http")).
								WithPath("/").
								WithScheme(corev1.URISchemeHTTP),
							),
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
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current appsv1.Deployment
	err = r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: depName}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	currApplyConfig, err := appsv1apply.ExtractDeployment(&current, "markdown-view-controller")
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(dep, currApplyConfig) {
		return nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "markdown-view-controller",
		Force:        pointer.Bool(true),
	})

	if err != nil {
		logger.Error(err, "unable to create or update Deployment")
		return err
	}
	logger.Info("reconcile Deployment successfully", "name", mdView.Name)
	return nil
}

//! [reconcile-deployment]

//! [reconcile-service]

func (r *MarkdownViewReconciler) reconcileService(ctx context.Context, mdView viewv1.MarkdownView) error {
	logger := log.FromContext(ctx)
	//! [service-apply-configuration]
	svcName := "viewer-" + mdView.Name

	owner, err := controllerReference(mdView, r.Scheme)
	if err != nil {
		return err
	}

	svc := corev1apply.Service(svcName, mdView.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       "mdbook",
			"app.kubernetes.io/instance":   mdView.Name,
			"app.kubernetes.io/created-by": "markdown-view-controller",
		}).
		WithOwnerReferences(owner).
		WithSpec(corev1apply.ServiceSpec().
			WithSelector(map[string]string{
				"app.kubernetes.io/name":       "mdbook",
				"app.kubernetes.io/instance":   mdView.Name,
				"app.kubernetes.io/created-by": "markdown-view-controller",
			}).
			WithType(corev1.ServiceTypeClusterIP).
			WithPorts(corev1apply.ServicePort().
				WithProtocol(corev1.ProtocolTCP).
				WithPort(80).
				WithTargetPort(intstr.FromInt(3000)),
			),
		)
	//! [service-apply-configuration]

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(svc)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	var current corev1.Service
	err = r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: svcName}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	currApplyConfig, err := corev1apply.ExtractService(&current, "markdown-view-controller")
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(svc, currApplyConfig) {
		return nil
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "markdown-view-controller",
		Force:        pointer.Bool(true),
	})
	if err != nil {
		logger.Error(err, "unable to create or update Service")
		return err
	}

	logger.Info("reconcile Service successfully", "name", mdView.Name)
	return nil
}

//! [reconcile-service]

//! [update-status]

func (r *MarkdownViewReconciler) updateStatus(ctx context.Context, mdView viewv1.MarkdownView) (ctrl.Result, error) {
	prevStatus := mdView.Status.DeepCopy()

	meta.SetStatusCondition(&mdView.Status.Conditions, metav1.Condition{
		Type:   viewv1.TypeMarkdownViewAvailable,
		Status: metav1.ConditionTrue,
		Reason: "OK",
	})
	meta.SetStatusCondition(&mdView.Status.Conditions, metav1.Condition{
		Type:   viewv1.TypeMarkdownViewDegraded,
		Status: metav1.ConditionFalse,
		Reason: "OK",
	})

	var cm corev1.ConfigMap
	err := r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: "markdowns-" + mdView.Name}, &cm)
	if errors.IsNotFound(err) {
		meta.SetStatusCondition(&mdView.Status.Conditions, metav1.Condition{
			Type:    viewv1.TypeMarkdownViewDegraded,
			Status:  metav1.ConditionTrue,
			Reason:  "Reconciling",
			Message: "ConfigMap not found",
		})
		meta.SetStatusCondition(&mdView.Status.Conditions, metav1.Condition{
			Type:   viewv1.TypeMarkdownViewAvailable,
			Status: metav1.ConditionFalse,
			Reason: "Reconciling",
		})
	} else if err != nil {
		return ctrl.Result{}, err
	}

	var svc corev1.Service
	err = r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: "viewer-" + mdView.Name}, &svc)
	if errors.IsNotFound(err) {
		meta.SetStatusCondition(&mdView.Status.Conditions, metav1.Condition{
			Type:    viewv1.TypeMarkdownViewDegraded,
			Status:  metav1.ConditionTrue,
			Reason:  "Reconciling",
			Message: "Service not found",
		})
		meta.SetStatusCondition(&mdView.Status.Conditions, metav1.Condition{
			Type:   viewv1.TypeMarkdownViewAvailable,
			Status: metav1.ConditionFalse,
			Reason: "Reconciling",
		})
	} else if err != nil {
		return ctrl.Result{}, err
	}

	var dep appsv1.Deployment
	err = r.Get(ctx, client.ObjectKey{Namespace: mdView.Namespace, Name: "viewer-" + mdView.Name}, &dep)
	if errors.IsNotFound(err) {
		meta.SetStatusCondition(&mdView.Status.Conditions, metav1.Condition{
			Type:    viewv1.TypeMarkdownViewDegraded,
			Status:  metav1.ConditionTrue,
			Reason:  "Reconciling",
			Message: "Deployment not found",
		})
		meta.SetStatusCondition(&mdView.Status.Conditions, metav1.Condition{
			Type:   viewv1.TypeMarkdownViewAvailable,
			Status: metav1.ConditionFalse,
			Reason: "Reconciling",
		})
	} else if err != nil {
		return ctrl.Result{}, err
	}

	result := ctrl.Result{}
	if dep.Status.AvailableReplicas == 0 {
		meta.SetStatusCondition(&mdView.Status.Conditions, metav1.Condition{
			Type:    viewv1.TypeMarkdownViewAvailable,
			Status:  metav1.ConditionFalse,
			Reason:  "Unavailable",
			Message: "AvailableReplicas is 0",
		})
		result = ctrl.Result{Requeue: true}
	}

	//! [call-set-metrics]
	r.setMetrics(mdView)
	//! [call-set-metrics]

	//! [call-recorder-event]
	if meta.IsStatusConditionFalse(prevStatus.Conditions, viewv1.TypeMarkdownViewDegraded) &&
		meta.IsStatusConditionTrue(mdView.Status.Conditions, viewv1.TypeMarkdownViewDegraded) {
		r.Recorder.Event(&mdView, corev1.EventTypeWarning, "Degraded", fmt.Sprintf("MarkdownView(%s:%s) degraded", mdView.Namespace, mdView.Name))
	}
	if meta.IsStatusConditionFalse(prevStatus.Conditions, viewv1.TypeMarkdownViewAvailable) &&
		meta.IsStatusConditionTrue(mdView.Status.Conditions, viewv1.TypeMarkdownViewAvailable) {
		r.Recorder.Event(&mdView, corev1.EventTypeNormal, "Available", fmt.Sprintf("MarkdownView(%s:%s) available", mdView.Namespace, mdView.Name))
	}
	//! [call-recorder-event]

	err = r.Status().Update(ctx, &mdView)
	return result, err
}

//! [update-status]

//! [set-metrics]

func (r *MarkdownViewReconciler) setMetrics(mdView viewv1.MarkdownView) {
	if meta.IsStatusConditionTrue(mdView.Status.Conditions, viewv1.TypeMarkdownViewAvailable) {
		AvailableVec.WithLabelValues(mdView.Name, mdView.Namespace).Set(1)
	} else {
		AvailableVec.WithLabelValues(mdView.Name, mdView.Namespace).Set(0)
	}
}

//! [set-metrics]

//! [remove-metrics]

func (r *MarkdownViewReconciler) removeMetrics(mdView viewv1.MarkdownView) {
	AvailableVec.DeleteLabelValues(mdView.Name, mdView.Namespace)
}

//! [remove-metrics]

//! [controller-reference]

func controllerReference(mdView viewv1.MarkdownView, scheme *runtime.Scheme) (*metav1apply.OwnerReferenceApplyConfiguration, error) {
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

//! [controller-reference]

//! [index]

const ownerControllerField = ".metadata.ownerReference.controller"

func indexByOwnerMarkdownView(obj client.Object) []string {
	cm := obj.(*corev1.ConfigMap)
	owner := metav1.GetControllerOf(cm)
	if owner == nil {
		return nil
	}
	if owner.APIVersion != viewv1.GroupVersion.String() || owner.Kind != "MarkdownView" {
		return nil
	}
	return []string{owner.Name}
}

//! [index]

// SetupWithManager sets up the controller with the Manager.
func (r *MarkdownViewReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, ch chan event.TypedGenericEvent[*viewv1.MarkdownView]) error {
	//! [index-field]
	err := mgr.GetFieldIndexer().IndexField(ctx, &corev1.ConfigMap{}, ownerControllerField, indexByOwnerMarkdownView)
	if err != nil {
		return err
	}
	//! [index-field]

	//! [managedby]
	return ctrl.NewControllerManagedBy(mgr).
		For(&viewv1.MarkdownView{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		WatchesRawSource(source.Channel(ch, &handler.TypedEnqueueRequestForObject[*viewv1.MarkdownView]{})).
		Complete(r)
	//! [managedby]
}
