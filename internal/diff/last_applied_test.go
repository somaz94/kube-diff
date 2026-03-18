package diff

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestExtractLastAppliedNil(t *testing.T) {
	result, err := ExtractLastApplied(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for nil input")
	}
}

func TestExtractLastAppliedNoAnnotation(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata":   map[string]interface{}{"name": "test"},
		},
	}

	result, err := ExtractLastApplied(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil when no annotation present")
	}
}

func TestExtractLastAppliedEmptyAnnotation(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "test",
				"annotations": map[string]interface{}{
					"kubectl.kubernetes.io/last-applied-configuration": "",
				},
			},
		},
	}

	result, err := ExtractLastApplied(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for empty annotation")
	}
}

func TestExtractLastAppliedValid(t *testing.T) {
	lastAppliedJSON := `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"test","namespace":"default"},"data":{"key":"value"}}`
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "test",
				"annotations": map[string]interface{}{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedJSON,
				},
			},
		},
	}

	result, err := ExtractLastApplied(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.GetKind() != "ConfigMap" {
		t.Errorf("expected kind ConfigMap, got %s", result.GetKind())
	}
	if result.GetName() != "test" {
		t.Errorf("expected name test, got %s", result.GetName())
	}

	data, ok := result.Object["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data field")
	}
	if data["key"] != "value" {
		t.Errorf("expected key=value, got %v", data["key"])
	}
}

func TestExtractLastAppliedInvalidJSON(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "test",
				"annotations": map[string]interface{}{
					"kubectl.kubernetes.io/last-applied-configuration": "{invalid json",
				},
			},
		},
	}

	_, err := ExtractLastApplied(obj)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCompareWithLastAppliedStrategy(t *testing.T) {
	lastAppliedJSON := `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"test","namespace":"default"},"data":{"key":"value"}}`

	local := newObj("v1", "ConfigMap", "test", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})
	cluster := newObj("v1", "ConfigMap", "test", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "cluster-changed-value"},
	})
	cluster.Object["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
		"kubectl.kubernetes.io/last-applied-configuration": lastAppliedJSON,
	}

	// With live strategy → changed (cluster has different value)
	liveOpts := CompareOptions{ContextLines: 3, Strategy: StrategyLive}
	result, err := Compare(local, cluster, liveOpts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusChanged {
		t.Errorf("live strategy: expected changed, got %s", result.Status)
	}

	// With last-applied strategy → unchanged (last-applied matches local)
	lastAppliedOpts := CompareOptions{ContextLines: 3, Strategy: StrategyLastApplied}
	result, err = Compare(local, cluster, lastAppliedOpts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusUnchanged {
		t.Errorf("last-applied strategy: expected unchanged, got %s", result.Status)
	}
}

func TestCompareWithLastAppliedFallbackToLive(t *testing.T) {
	local := newObj("v1", "ConfigMap", "test", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "new-value"},
	})
	cluster := newObj("v1", "ConfigMap", "test", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "old-value"},
	})
	// No last-applied annotation → should fallback to live

	opts := CompareOptions{ContextLines: 3, Strategy: StrategyLastApplied}
	result, err := Compare(local, cluster, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusChanged {
		t.Errorf("expected changed (fallback to live), got %s", result.Status)
	}
}
