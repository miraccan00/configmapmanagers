package configmap

import (
	"context"
	"fmt"

	"github.com/miraccan00/configmapmanagers/models"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// UpdateConfigMap updates a ConfigMap with given updates
func UpdateConfigMap(clientset *kubernetes.Clientset, namespace string, configMapSpec models.ConfigMapSpec) error {
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapSpec.Name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get ConfigMap %s: %v", configMapSpec.Name, err)
	}

	for _, update := range configMapSpec.Updates {
		configMap.Data[update.Key] = update.NewValue
	}

	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to update ConfigMap %s: %v", configMapSpec.Name, err)
	}

	fmt.Printf("ConfigMap %s updated in namespace %s with updates: %v\n", configMapSpec.Name, namespace, configMapSpec.Updates)
	return nil
}
