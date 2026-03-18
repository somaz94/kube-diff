package diff

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newObj(apiVersion, kind, name, namespace string, extra map[string]interface{}) *unstructured.Unstructured {
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

func TestCompareNewResource(t *testing.T) {
	local := newObj("v1", "ConfigMap", "my-cm", "default", nil)

	result, err := Compare(local, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusNew {
		t.Errorf("expected StatusNew, got %s", result.Status)
	}
	if result.Kind != "ConfigMap" {
		t.Errorf("expected ConfigMap, got %s", result.Kind)
	}
	if result.Name != "my-cm" {
		t.Errorf("expected my-cm, got %s", result.Name)
	}
}

func TestCompareUnchanged(t *testing.T) {
	local := newObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})
	cluster := newObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})

	result, err := Compare(local, cluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusUnchanged {
		t.Errorf("expected StatusUnchanged, got %s", result.Status)
	}
	if result.Diff != "" {
		t.Errorf("expected empty diff, got %s", result.Diff)
	}
}

func TestCompareChanged(t *testing.T) {
	local := newObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "new-value"},
	})
	cluster := newObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "old-value"},
	})

	result, err := Compare(local, cluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusChanged {
		t.Errorf("expected StatusChanged, got %s", result.Status)
	}
	if result.Diff == "" {
		t.Error("expected non-empty diff")
	}
	if !strings.Contains(result.Diff, "old-value") || !strings.Contains(result.Diff, "new-value") {
		t.Error("diff should contain old and new values")
	}
}

func TestCompareIgnoresClusterFields(t *testing.T) {
	local := newObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})
	cluster := newObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})
	// Add cluster-managed fields
	cluster.Object["metadata"].(map[string]interface{})["uid"] = "abc-123"
	cluster.Object["metadata"].(map[string]interface{})["resourceVersion"] = "999"
	cluster.Object["metadata"].(map[string]interface{})["creationTimestamp"] = "2024-01-01T00:00:00Z"
	cluster.Object["status"] = map[string]interface{}{"phase": "Active"}

	result, err := Compare(local, cluster)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusUnchanged {
		t.Errorf("expected StatusUnchanged after normalization, got %s", result.Status)
	}
}

func TestResourceKeyWithNamespace(t *testing.T) {
	r := &DiffResult{Kind: "Deployment", Name: "app", Namespace: "production"}
	expected := "Deployment/app (namespace: production)"
	if r.ResourceKey() != expected {
		t.Errorf("expected %q, got %q", expected, r.ResourceKey())
	}
}

func TestResourceKeyWithoutNamespace(t *testing.T) {
	r := &DiffResult{Kind: "ClusterRole", Name: "admin"}
	expected := "ClusterRole/admin"
	if r.ResourceKey() != expected {
		t.Errorf("expected %q, got %q", expected, r.ResourceKey())
	}
}

func TestToYAML(t *testing.T) {
	obj := newObj("v1", "ConfigMap", "test", "", nil)
	yamlStr, err := toYAML(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(yamlStr, "kind: ConfigMap") {
		t.Error("expected YAML to contain 'kind: ConfigMap'")
	}
	if !strings.Contains(yamlStr, "name: test") {
		t.Error("expected YAML to contain 'name: test'")
	}
}
