package main

import (
	"context"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
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

	err = patch(cli)
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

//! [patch]
func patch(cli client.Client) error {
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
	}

	err := cli.Patch(context.Background(), patch, client.Apply, &client.PatchOptions{
		FieldManager: "misc",
	})

	return err
}

//! [patch]
