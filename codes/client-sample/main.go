package main

import (
	"context"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func main() {
	ctx := context.Background()
	cli, err := getClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = create(ctx, cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = createOrUpdate(ctx, cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

RETRY:
	err = get(ctx, cli)
	if errors.IsNotFound(err) {
		goto RETRY
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = list(ctx, cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = pagination(ctx, cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = patchMerge(ctx, cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = patchApply(ctx, cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = patchApplyConfig(ctx, cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = deleteWithPreConditions(ctx, cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = deleteAllOfDeployment(ctx, cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getClient() (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	scm := runtime.NewScheme()
	err = scheme.AddToScheme(scm)
	if err != nil {
		return nil, err
	}
	cli, err := client.New(cfg, client.Options{Scheme: scm})
	if err != nil {
		return nil, err
	}
	return cli, nil
}

//! [create]
func create(ctx context.Context, cli client.Client) error {
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
	err := cli.Create(ctx, &dep)
	if err != nil {
		return err
	}
	return nil
}

//! [create]

//! [create-or-update]
func createOrUpdate(ctx context.Context, cli client.Client) error {
	dep := &appsv1.Deployment{}
	dep.SetNamespace("default")
	dep.SetName("sample")

	op, err := ctrl.CreateOrUpdate(ctx, cli, dep, func() error {
		dep.Spec.Replicas = pointer.Int32Ptr(1)
		dep.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": "nginx"},
		}
		dep.Spec.Template = corev1.PodTemplateSpec{
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
		}
		return nil
	})
	if err != nil {
		return err
	}
	if op != controllerutil.OperationResultNone {
		fmt.Printf("Deployment %s\n", op)
	}
	return nil
}

//! [create-or-update]

//! [get]
func get(ctx context.Context, cli client.Client) error {
	var deployment appsv1.Deployment
	err := cli.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &deployment)
	if err != nil {
		return err
	}
	fmt.Printf("Got Deployment: %#v\n", deployment)
	return nil
}

//! [get]

//! [list]
func list(ctx context.Context, cli client.Client) error {
	var pods corev1.PodList
	err := cli.List(ctx, &pods, &client.ListOptions{
		Namespace:     "default",
		LabelSelector: labels.SelectorFromSet(map[string]string{"app": "sample"}),
	})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
	return nil
}

//! [list]

//! [pagination]
func pagination(ctx context.Context, cli client.Client) error {
	token := ""
	for i := 0; ; i++ {
		var pods corev1.PodList
		err := cli.List(ctx, &pods, &client.ListOptions{
			Limit:    3,
			Continue: token,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Page %d:\n", i)
		for _, pod := range pods.Items {
			fmt.Println(pod.Name)
		}
		fmt.Println()

		token = pods.ListMeta.Continue
		if len(token) == 0 {
			return nil
		}
	}
}

//! [pagination]

//! [cond]
func deleteWithPreConditions(ctx context.Context, cli client.Client) error {
	var deploy appsv1.Deployment
	err := cli.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &deploy)
	if err != nil {
		return err
	}
	uid := deploy.GetUID()
	resourceVersion := deploy.GetResourceVersion()
	cond := metav1.Preconditions{
		UID:             &uid,
		ResourceVersion: &resourceVersion,
	}
	err = cli.Delete(ctx, &deploy, &client.DeleteOptions{
		Preconditions: &cond,
	})
	return err
}

//! [cond]

//! [delete-all-of]
func deleteAllOfDeployment(ctx context.Context, cli client.Client) error {
	err := cli.DeleteAllOf(ctx, &appsv1.Deployment{}, client.InNamespace("default"))
	return err
}

//! [delete-all-of]

//! [update-status]
func updateStatus(ctx context.Context, cli client.Client) error {
	var dep appsv1.Deployment
	err := cli.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &dep)
	if err != nil {
		return err
	}

	dep.Status.AvailableReplicas = 3
	err = cli.Status().Update(ctx, &dep)
	return err
}

//! [update-status]

//! [patch-merge]
func patchMerge(ctx context.Context, cli client.Client) error {
	var dep appsv1.Deployment
	err := cli.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample"}, &dep)
	if err != nil {
		return err
	}

	newDep := dep.DeepCopy()
	newDep.Spec.Replicas = pointer.Int32Ptr(3)
	patch := client.MergeFrom(&dep)

	err = cli.Patch(ctx, newDep, patch)

	return err
}

//! [patch-merge]

//! [patch-apply]
func patchApply(ctx context.Context, cli client.Client) error {
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

	err := cli.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "client-sample",
	})

	return err
}

//! [patch-apply]

//! [patch-apply-config]
func patchApplyConfig(ctx context.Context, cli client.Client) error {
	dep := appsv1apply.Deployment("sample3", "default").
		WithSpec(appsv1apply.DeploymentSpec().
			WithReplicas(3).
			WithSelector(metav1apply.LabelSelector().WithMatchLabels(map[string]string{"app": "nginx"})).
			WithTemplate(corev1apply.PodTemplateSpec().
				WithLabels(map[string]string{"app": "nginx"}).
				WithSpec(corev1apply.PodSpec().
					WithContainers(corev1apply.Container().
						WithName("nginx").
						WithImage("nginx:latest"),
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
	err = cli.Get(ctx, client.ObjectKey{Namespace: "default", Name: "sample3"}, &current)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	currApplyConfig, err := appsv1apply.ExtractDeployment(&current, "client-sample")
	if err != nil {
		return err
	}

	if equality.Semantic.DeepEqual(dep, currApplyConfig) {
		return nil
	}

	err = cli.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "client-sample",
	})
	return err
}

//! [patch-apply-config]
