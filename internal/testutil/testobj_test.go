package testutil

import "testing"

// NewTestObj underpins the fixtures of every other package's tests, so its own
// behavior is pinned here. Without a test in this package its coverage also
// reads as 0%, which makes a heavily used helper look like dead code.
func TestNewTestObj(t *testing.T) {
	obj := NewTestObj("v1", "ConfigMap", "my-cm", "default", nil)

	if got := obj.GetAPIVersion(); got != "v1" {
		t.Errorf("APIVersion = %q, want %q", got, "v1")
	}
	if got := obj.GetKind(); got != "ConfigMap" {
		t.Errorf("Kind = %q, want %q", got, "ConfigMap")
	}
	if got := obj.GetName(); got != "my-cm" {
		t.Errorf("Name = %q, want %q", got, "my-cm")
	}
	if got := obj.GetNamespace(); got != "default" {
		t.Errorf("Namespace = %q, want %q", got, "default")
	}
}

// An empty namespace must be left out entirely rather than written as "", so
// cluster-scoped fixtures do not carry a blank namespace field.
func TestNewTestObj_OmitsEmptyNamespace(t *testing.T) {
	obj := NewTestObj("v1", "Namespace", "kube-system", "", nil)

	metadata, ok := obj.Object["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata = %T, want a map", obj.Object["metadata"])
	}
	if _, present := metadata["namespace"]; present {
		t.Errorf("metadata carries a namespace key, want it omitted for an empty namespace")
	}
}

func TestNewTestObj_MergesExtraFields(t *testing.T) {
	obj := NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})

	data, ok := obj.Object["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data = %T, want a map", obj.Object["data"])
	}
	if data["key"] != "value" {
		t.Errorf("data[key] = %v, want %q", data["key"], "value")
	}
	// The base fields must survive the merge.
	if obj.GetName() != "my-cm" {
		t.Errorf("Name = %q, want the base field preserved", obj.GetName())
	}
}

// An extra field may deliberately replace a base one, which is how tests build
// objects with custom metadata.
func TestNewTestObj_ExtraOverridesBase(t *testing.T) {
	obj := NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"metadata": map[string]interface{}{"name": "overridden"},
	})

	if got := obj.GetName(); got != "overridden" {
		t.Errorf("Name = %q, want the extra field to win", got)
	}
}
