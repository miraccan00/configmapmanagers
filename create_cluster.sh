#!/bin/bash

# Define the cluster name
CLUSTER_NAME="my-cluster"

# Create the kind cluster
kind create cluster --name ${CLUSTER_NAME}

# Check if the cluster creation was successful
if [ $? -ne 0 ]; then
  echo "Failed to create kind cluster"
  exit 1
fi

# Get the kubeconfig path for the kind cluster
KUBECONFIG_PATH=$(kind get kubeconfig --name ${CLUSTER_NAME})

# Export the kubeconfig to .kubeconf file
echo "${KUBECONFIG_PATH}" > .kubeconf

# Check if the kubeconfig file was created successfully
if [ -f .kubeconf ]; then
  echo "Kubeconfig file has been exported to .kubeconf"
else
  echo "Failed to export kubeconfig file"
  exit 1
fi

# Export the KUBECONFIG environment variable
export KUBECONFIG=$(pwd)/.kubeconfig

echo "KUBECONFIG environment variable has been set to $(pwd)/.kubeconf"
