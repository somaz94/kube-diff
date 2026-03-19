package testutil

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// NewTestObj creates an unstructured Kubernetes object for testing.
func NewTestObj(apiVersion, kind, name, namespace string, extra map[string]interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}
	if namespace != "" {
		obj.Object["metadata"].(map[string]interface{})["namespace"] = namespace
	}
	for k, v := range extra {
		obj.Object[k] = v
	}
	return obj
}
