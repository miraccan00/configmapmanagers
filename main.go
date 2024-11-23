package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ConfigMapUpdate represents an individual update for a key-value pair in a ConfigMap.
type ConfigMapUpdate struct {
	Key      string `json:"key"`
	NewValue string `json:"newValue"`
}

// ConfigMapSpec represents a ConfigMap and its updates.
type ConfigMapSpec struct {
	Name    string            `json:"name"`
	Updates []ConfigMapUpdate `json:"updates"`
}

// ConfigMapManagerSpec represents the spec of the Custom Resource.
type ConfigMapManagerSpec struct {
	ConfigMaps []ConfigMapSpec `json:"configMaps"`
}

func main() {
	// Kubernetes client oluştur
	clientset, dynamicClient, err := getKubernetesClients()
	if err != nil {
		fmt.Printf("Failed to create Kubernetes client: %v\n", err)
		return
	}

	// Belirli bir interval ile CR'leri kontrol eden sonsuz döngü
	ticker := time.NewTicker(10 * time.Second) // 10 saniye aralıkla çalışacak
	defer ticker.Stop()

	fmt.Println("Starting Custom Resource monitoring...")

	for {
		select {
		case <-ticker.C:
			fmt.Println("Checking Custom Resources...")
			if err := processCustomResources(clientset, dynamicClient); err != nil {
				fmt.Printf("Failed to process Custom Resources: %v\n", err)
			}
		}
	}
}

func getKubernetesClients() (*kubernetes.Clientset, dynamic.Interface, error) {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		// Try in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to build kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create Kubernetes clientset: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create dynamic client: %v", err)
	}

	return clientset, dynamicClient, nil
}

func processCustomResources(clientset *kubernetes.Clientset, dynamicClient dynamic.Interface) error {
	resource := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    "blacksyrius.ci.com",
		Version:  "v1",
		Resource: "configmapmanagers",
	})

	// Custom Resource'ları listele
	crList, err := resource.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list custom resources: %v", err)
	}

	// Tüm namespace'leri listele
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list namespaces: %v", err)
	}

	// Sadece `owner: blacksyrius` annotation'ı içeren namespace'leri işleme al
	for _, ns := range namespaces.Items {
		if owner, ok := ns.Annotations["owner"]; !ok || !strings.EqualFold(owner, "blacksyrius") {
			continue
		}

		namespaceName := ns.Name
		fmt.Printf("Processing namespace: %s\n", namespaceName)

		// Her Custom Resource için ConfigMap ve Deployment'ları işleme al
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

			var spec ConfigMapManagerSpec
			if err := json.Unmarshal(specJSON, &spec); err != nil {
				fmt.Printf("Failed to unmarshal spec: %v\n", err)
				continue
			}

			// ConfigMap'leri işle
			for _, configMapSpec := range spec.ConfigMaps {
				// ConfigMap'i güncelle
				if err := updateConfigMap(clientset, namespaceName, configMapSpec.Name, configMapSpec.Updates); err != nil {
					fmt.Printf("Failed to update ConfigMap %s in namespace %s: %v\n", configMapSpec.Name, namespaceName, err)
					continue
				}

				// Deployment'ları yeniden başlat
				if err := restartDeploymentsUsingConfigMap(clientset, namespaceName, configMapSpec.Name); err != nil {
					fmt.Printf("Failed to restart deployments for %s in namespace %s: %v\n", configMapSpec.Name, namespaceName, err)
					continue
				}
			}
		}
	}

	return nil
}

func updateConfigMap(clientset *kubernetes.Clientset, namespace, configMapName string, updates []ConfigMapUpdate) error {
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get ConfigMap %s: %v", configMapName, err)
	}

	// ConfigMap'i güncelle
	for _, update := range updates {
		configMap.Data[update.Key] = update.NewValue
	}

	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to update ConfigMap %s: %v", configMapName, err)
	}

	fmt.Printf("ConfigMap %s updated in namespace %s with updates: %v\n", configMapName, namespace, updates)
	return nil
}

func restartDeploymentsUsingConfigMap(clientset *kubernetes.Clientset, namespace, configMapName string) error {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list deployments in namespace %s: %v", namespace, err)
	}

	for _, deployment := range deployments.Items {
		restart := false
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.EnvFrom != nil {
				for _, envFrom := range container.EnvFrom {
					if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == configMapName {
						restart = true
						break
					}
				}
			}
			if restart {
				break
			}
		}

		if restart {
			// Annotation ekleyerek pod'u yeniden başlat
			if deployment.Spec.Template.Annotations == nil {
				deployment.Spec.Template.Annotations = make(map[string]string)
			}
			deployment.Spec.Template.Annotations["configmap-update-timestamp"] = time.Now().Format(time.RFC3339)

			_, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), &deployment, v1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to restart deployment %s: %v", deployment.Name, err)
			}

			fmt.Printf("Restarted deployment: %s in namespace %s\n", deployment.Name, namespace)
		}
	}

	return nil
}
