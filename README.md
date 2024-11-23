# ConfigMapGenerator

ConfigMapGenerator is a Kubernetes operator written in Go that automates the management of ConfigMaps and restarts associated Deployments. It uses a Custom Resource Definition (CRD) called `ConfigMapManager` to define updates for multiple ConfigMaps and automatically applies these changes to Kubernetes resources across namespaces.

---

## **Features**
- Supports bulk updates to multiple ConfigMaps using a single `ConfigMapManager` Custom Resource.
- Automatically restarts Deployments that are linked to updated ConfigMaps using `envFrom`.
- Monitors namespaces with a specific annotation (e.g., `owner: blacksyrius`) for targeted ConfigMap updates.
- Modular design for easy maintenance and extensibility.
- Built with Go and uses the Kubernetes client-go library.

---

## **Folder Structure**

```
project/
├── cmd/
│   └── main.go               # Entry point of the application
├── controllers/
│   ├── cr_processor.go       # Handles Custom Resource processing
│   ├── configmap.go          # Manages ConfigMap updates
│   └── deployment.go         # Restarts Deployments linked to ConfigMaps
├── clients/
│   └── k8s_client.go         # Kubernetes client initialization
├── models/
│   └── configmap_manager.go  # Data models for ConfigMapManager CRD
├── utils/
│   └── logger.go             # Centralized logging utility
├── go.mod                    # Go module dependencies
└── go.sum                    # Dependency checksums
```

## **Custom Resource Definition (CRD)**

The *ConfigMapManager* CRD defines which ConfigMaps to update and their associated key-value updates. Here is an example:

## **CRD Example**

```
apiVersion: blacksyrius.ci.com/v1
kind: ConfigMapManager
metadata:
  name: my-configmap-update
  namespace: default
spec:
  configMaps:
    - name: example-config
      updates:
        - key: LOG_LEVEL
          newValue: "info"
        - key: DB_HOST
          newValue: "mysql.example.com"
    - name: another-config
      updates:
        - key: FEATURE_FLAG
          newValue: "true"
        - key: RETRY_LIMIT
          newValue: "5"

```

## **How It Works**

- Custom Resource Deployment:
    - Users create a ConfigMapManager resource to define which ConfigMaps and keys to update.

- Namespace Filtering:
    - The operator processes only namespaces with the annotation owner: blacksyrius.
        

- ConfigMap Updates:
    - The operator updates the specified ConfigMaps with the provided key-value pairs.

- Deployment Restarts:
    - Deployments that reference the updated ConfigMaps via envFrom are automatically restarted by adding a timestamp annotation to their pod templates.

## **Installation**

### Prerequisites
- Kubernetes cluster (v1.20+ recommended)
- Go (v1.20 or higher)
- kubectl CLI configured to access your cluster


## **Steps**

- Clone the Repository:
    ```
    git clone https://github.com/miraccan00/configmapmanagers.git

    cd configmapmanagers
    ```

- Build the Operator:
```
go build -o configmapgenerator ./cmd/main.go
```
- Deploy the CRD: Apply the ConfigMapManager CRD to your cluster:
    ```
    kubectl apply -f deploy/crd.yaml
    ```
- Run the Operator: Run the operator binary:
    ```
    ./configmapgenerator
    ```
## Usage

Create a Namespace with Annotations: Ensure the namespace you want to target has the annotation owner: blacksyrius. 

Example:

```
apiVersion: v1
kind: Namespace
metadata:
  name: example-namespace
  annotations:
    owner: blacksyrius
```
Apply it with:

```kubectl apply -f namespace.yaml```

Create a ConfigMapManager Resource: Create a ConfigMapManager YAML file defining the ConfigMaps and updates. Example:

```
apiVersion: blacksyrius.ci.com/v1
kind: ConfigMapManager
metadata:
  name: my-configmap-update
  namespace: example-namespace
spec:
  configMaps:
    - name: example-config
      updates:
        - key: LOG_LEVEL
          newValue: "info"
        - key: DB_HOST
          newValue: "mysql.example.com"


```

Apply it to your cluster:

```kubectl apply -f configmapmanager.yaml```

Verify the Updates:
- Check the updated ConfigMap:
    - ```kubectl get configmap example-config -n example-namespace -o yaml```
- Ensure the associated Deployments have restarted:
    - ```kubectl describe deployment <deployment-name> -n example-namespace```

## **Development**

**Adding Features**

To add a new feature, create a new Go module under the appropriate folder (e.g., controllers, utils).


## **Testing**

- Run unit tests:

    ```go test ./...```

- Simulate Custom Resource updates in a local cluster:

    ```kubectl apply -f deploy/test-configmapmanager.yaml```

##  **Debugging**

Use the logs to debug issues:

```tail -f application.log```

Contributing

Contributions are welcome! Please fork the repository, make changes in a separate branch, and submit a pull request.

Licence
This project is licensed under the MIT License. See the LICENSE file for more details.

Contact

For support or inquiries, reach out to miraccanyilmazme@gmail.com.