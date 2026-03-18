package diff

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const lastAppliedAnnotation = "kubectl.kubernetes.io/last-applied-configuration"

// ExtractLastApplied extracts the last-applied-configuration annotation
// from a cluster resource and returns it as an Unstructured object.
// Returns nil if the annotation is not present.
func ExtractLastApplied(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	annotations := obj.GetAnnotations()
	if annotations == nil {
		return nil, nil
	}

	raw, ok := annotations[lastAppliedAnnotation]
	if !ok || raw == "" {
		return nil, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("failed to parse last-applied-configuration: %w", err)
	}

	return &unstructured.Unstructured{Object: data}, nil
}
