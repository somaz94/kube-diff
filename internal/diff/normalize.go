package diff

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// fieldsToRemove are cluster-managed fields that should be ignored in diffs.
var fieldsToRemove = []string{
	"metadata.managedFields",
	"metadata.resourceVersion",
	"metadata.uid",
	"metadata.creationTimestamp",
	"metadata.generation",
	"metadata.selfLink",
	"metadata.annotations.kubectl.kubernetes.io/last-applied-configuration",
	"status",
}

// Normalize removes cluster-managed fields from a resource for clean comparison.
func Normalize(obj *unstructured.Unstructured) *unstructured.Unstructured {
	if obj == nil {
		return nil
	}

	normalized := obj.DeepCopy()

	// Remove top-level status
	delete(normalized.Object, "status")

	// Remove metadata fields
	if metadata, ok := normalized.Object["metadata"].(map[string]interface{}); ok {
		delete(metadata, "managedFields")
		delete(metadata, "resourceVersion")
		delete(metadata, "uid")
		delete(metadata, "creationTimestamp")
		delete(metadata, "generation")
		delete(metadata, "selfLink")

		// Remove specific annotations
		if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
			delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
			delete(annotations, "deployment.kubernetes.io/revision")
			if len(annotations) == 0 {
				delete(metadata, "annotations")
			}
		}
	}

	return normalized
}
