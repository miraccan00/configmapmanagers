package main

import (
	"fmt"
	"time"

	"github.com/miraccan00/configmapmanagers/clients"
	"github.com/miraccan00/configmapmanagers/controllers"
	"github.com/miraccan00/configmapmanagers/utils"
)

func main() {
	// Initialize logger
	logger, err := utils.NewLogger("application.log")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	logger.Info("Starting Kubernetes Custom Resource monitoring...")

	// Kubernetes clients
	clientset, dynamicClient, err := clients.GetKubernetesClients()
	if err != nil {
		logger.Error(fmt.Errorf("Failed to create Kubernetes clients: %v", err))
		return
	}

	// Custom Resource monitoring loop
	ticker := time.NewTicker(10 * time.Second) // 10 seconds interval
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("Checking Custom Resources...")
			if err := controllers.ProcessCustomResources(clientset, dynamicClient); err != nil {
				logger.Error(fmt.Errorf("Failed to process Custom Resources: %v", err))
			}
		}
	}
}
