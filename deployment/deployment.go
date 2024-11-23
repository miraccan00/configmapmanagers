package deployment

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RestartDeploymentsUsingConfigMap restarts deployments using a ConfigMap
func RestartDeploymentsUsingConfigMap(clientset *kubernetes.Clientset, namespace, configMapName string) error {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to list deployments in namespace %s: %v", namespace, err)
	}

	for _, deployment := range deployments.Items {
		restart := false
		for _, container := range deployment.Spec.Template.Spec.Containers {
			for _, envFrom := range container.EnvFrom {
				if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == configMapName {
					restart = true
					break
				}
			}
		}

		if restart {
			if deployment.Spec.Template.Annotations == nil {
				deployment.Spec.Template.Annotations = make(map[string]string)
			}
			deployment.Spec.Template.Annotations["configmap-update-timestamp"] = time.Now().Format(time.RFC3339)

			_, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), &deployment, v1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("Failed to restart deployment %s: %v", deployment.Name, err)
			}

			fmt.Printf("Restarted deployment: %s in namespace %s using ConfigMap %s\n", deployment.Name, namespace, configMapName)
		}
	}

	return nil
}
