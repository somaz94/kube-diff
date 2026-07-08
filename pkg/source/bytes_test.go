package source

import "testing"

func TestBytesSource_Load(t *testing.T) {
	yaml := `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-a
  namespace: default
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dep-b
  namespace: prod
---
# a comment-only / empty document is skipped
---
not-a-kubernetes-doc: true
`

	resources, err := NewBytesSource([]byte(yaml)).Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	// Two valid resources; the empty doc and the kind-less doc are skipped.
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d: %+v", len(resources), resources)
	}
	if resources[0].Kind != "ConfigMap" || resources[0].Name != "cm-a" || resources[0].Namespace != "default" {
		t.Errorf("resource[0] mismatch: %+v", resources[0])
	}
	if resources[1].Kind != "Deployment" || resources[1].Name != "dep-b" || resources[1].Namespace != "prod" {
		t.Errorf("resource[1] mismatch: %+v", resources[1])
	}
}

func TestBytesSource_Empty(t *testing.T) {
	resources, err := NewBytesSource(nil).Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(resources) != 0 {
		t.Fatalf("expected 0 resources, got %d", len(resources))
	}
}
