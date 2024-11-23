package models

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
