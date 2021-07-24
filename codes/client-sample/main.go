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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	cli, err := getClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = list(cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = pagination(cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = deleteWithPreConditions(cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = deleteWithPropagationPolicy(cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = patchApply(cli)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = patchMerge(cli)
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

//! [list]
func list(cli client.Client) error {
	var pods corev1.PodList
	err := cli.List(context.Background(), &pods, &client.ListOptions{
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
func pagination(cli client.Client) error {
	token := ""
	for i := 0; ; i++ {
		var pods corev1.PodList
		err := cli.List(context.Background(), &pods, &client.ListOptions{
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
func deleteWithPreConditions(cli client.Client) error {
	var deploy appsv1.Deployment
	err := cli.Get(context.Background(), client.ObjectKey{
		Namespace: "default",
		Name:      "test",
	}, &deploy)
	if err != nil {
		return err
	}
	uid := deploy.GetUID()
	resourceVersion := deploy.GetResourceVersion()
	cond := metav1.Preconditions{
		UID:             &uid,
		ResourceVersion: &resourceVersion,
	}
	err = cli.Delete(context.Background(), &deploy, &client.DeleteOptions{
		Preconditions: &cond,
	})
	return err
}

//! [cond]

//! [policy]
func deleteWithPropagationPolicy(cli client.Client) error {
	var deploy appsv1.Deployment
	err := cli.Get(context.Background(), client.ObjectKey{
		Namespace: "default",
		Name:      "test",
	}, &deploy)
	if err != nil {
		return err
	}
	policy := metav1.DeletePropagationOrphan
	err = cli.Delete(context.Background(), &deploy, &client.DeleteOptions{
		PropagationPolicy: &policy,
	})
	return err
}

//! [policy]

//! [patch-merge]
func patchMerge(cli client.Client) error {
	var dep appsv1.Deployment
	err := cli.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "test"}, &dep)
	if err != nil {
		return err
	}

	newDep := dep.DeepCopy()
	newDep.Spec.Replicas = pointer.Int32Ptr(3)
	patch := client.MergeFrom(&dep)

	err = cli.Patch(context.Background(), newDep, patch)

	return err
}

//! [patch-merge]

//! [patch-apply]
func patchApply(cli client.Client) error {
	patch := &unstructured.Unstructured{}
	patch.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})
	patch.SetNamespace("default")
	patch.SetName("test")
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

	err := cli.Patch(context.Background(), patch, client.Apply, &client.PatchOptions{
		FieldManager: "client-sample",
	})

	return err
}

//! [patch-apply]

//! [patch-apply-config]
func patchApplyConfig(cli client.Client) error {
	dep := appsv1apply.Deployment("test", "default").
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
	err = cli.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "test"}, &current)
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

	err = cli.Patch(context.Background(), patch, client.Apply, &client.PatchOptions{
		FieldManager: "client-sample",
	})
	return err
}

//! [patch-apply-config]
