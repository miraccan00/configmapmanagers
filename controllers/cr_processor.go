package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	configmap "github.com/miraccan00/configmapmanagers/configmap"
	deployment "github.com/miraccan00/configmapmanagers/deployment"
	"github.com/miraccan00/configmapmanagers/models"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// ProcessCustomResources processes ConfigMapManager Custom Resources
func ProcessCustomResources(clientset *kubernetes.Clientset, dynamicClient dynamic.Interface) error {
	resource := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "blacksyrius.ci.com",
		Version:  "v1",
		Resource: "configmapmanagers",
	})

	crList, err := resource.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list custom resources: %v", err)
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list namespaces: %v", err)
	}

	for _, ns := range namespaces.Items {
		if owner, ok := ns.Annotations["owner"]; !ok || !strings.EqualFold(owner, "blacksyrius") {
			continue
		}

		namespaceName := ns.Name
		fmt.Printf("Processing namespace: %s\n", namespaceName)

		for _, item := range crList.Items {
			specData, ok := item.Object["spec"].(map[string]interface{})
			if !ok {
				fmt.Printf("Invalid spec in custom resource\n")
				continue
			}

			specJSON, err := json.Marshal(specData)
			if err != nil {
				fmt.Printf("Failed to marshal spec: %v\n", err)
				continue
			}

			var spec models.ConfigMapManagerSpec
			if err := json.Unmarshal(specJSON, &spec); err != nil {
				fmt.Printf("Failed to unmarshal spec: %v\n", err)
				continue
			}

			for _, configMapSpec := range spec.ConfigMaps {
				if err := configmap.UpdateConfigMap(clientset, namespaceName, configMapSpec); err != nil {
					fmt.Printf("Failed to update ConfigMap %s in namespace %s: %v\n", configMapSpec.Name, namespaceName, err)
					continue
				}

				if err := deployment.RestartDeploymentsUsingConfigMap(clientset, namespaceName, configMapSpec.Name); err != nil {
					fmt.Printf("Failed to restart deployments for %s in namespace %s: %v\n", configMapSpec.Name, namespaceName, err)
					continue
				}
			}
		}
	}

	return nil
}
