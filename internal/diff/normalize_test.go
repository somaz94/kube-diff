package diff

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNormalizeRemovesStatus(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "test",
			},
			"status": map[string]interface{}{
				"loadBalancer": map[string]interface{}{},
			},
		},
	}

	result := Normalize(obj)

	if _, ok := result.Object["status"]; ok {
		t.Error("expected status to be removed")
	}
}

func TestNormalizeRemovesMetadataFields(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":              "test",
				"uid":               "abc-123",
				"resourceVersion":   "999",
				"creationTimestamp":  "2024-01-01T00:00:00Z",
				"generation":        int64(5),
				"selfLink":          "/api/v1/configmaps/test",
				"managedFields":     []interface{}{},
			},
		},
	}

	result := Normalize(obj)
	metadata := result.Object["metadata"].(map[string]interface{})

	removedFields := []string{"uid", "resourceVersion", "creationTimestamp", "generation", "selfLink", "managedFields"}
	for _, field := range removedFields {
		if _, ok := metadata[field]; ok {
			t.Errorf("expected %s to be removed from metadata", field)
		}
	}

	if metadata["name"] != "test" {
		t.Error("expected name to be preserved")
	}
}

func TestNormalizeRemovesKubectlAnnotation(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "test",
				"annotations": map[string]interface{}{
					"kubectl.kubernetes.io/last-applied-configuration": "{}",
					"my-annotation": "keep-me",
				},
			},
		},
	}

	result := Normalize(obj)
	metadata := result.Object["metadata"].(map[string]interface{})
	annotations := metadata["annotations"].(map[string]interface{})

	if _, ok := annotations["kubectl.kubernetes.io/last-applied-configuration"]; ok {
		t.Error("expected kubectl annotation to be removed")
	}
	if annotations["my-annotation"] != "keep-me" {
		t.Error("expected custom annotation to be preserved")
	}
}

func TestNormalizeRemovesDeploymentRevisionAnnotation(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "test",
				"annotations": map[string]interface{}{
					"deployment.kubernetes.io/revision": "3",
				},
			},
		},
	}

	result := Normalize(obj)
	metadata := result.Object["metadata"].(map[string]interface{})

	// annotations should be removed entirely since it's empty
	if _, ok := metadata["annotations"]; ok {
		t.Error("expected empty annotations map to be removed")
	}
}

func TestNormalizeNilInput(t *testing.T) {
	result := Normalize(nil)
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestNormalizeDoesNotModifyOriginal(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":            "test",
				"uid":             "abc-123",
				"resourceVersion": "999",
			},
			"status": map[string]interface{}{
				"phase": "Active",
			},
		},
	}

	_ = Normalize(obj)

	// Original should be untouched
	if _, ok := obj.Object["status"]; !ok {
		t.Error("original object status should not be modified")
	}
	metadata := obj.Object["metadata"].(map[string]interface{})
	if _, ok := metadata["uid"]; !ok {
		t.Error("original object uid should not be modified")
	}
}
