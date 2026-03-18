package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseYAMLSingleDocument(t *testing.T) {
	yamlContent := `
apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: default
spec:
  selector:
    app: my-app
  ports:
    - port: 80
`

	resources, err := parseYAML(strings.NewReader(yamlContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	r := resources[0]
	if r.Kind != "Service" {
		t.Errorf("expected Kind=Service, got %s", r.Kind)
	}
	if r.Name != "my-service" {
		t.Errorf("expected Name=my-service, got %s", r.Name)
	}
	if r.Namespace != "default" {
		t.Errorf("expected Namespace=default, got %s", r.Namespace)
	}
	if r.APIVersion != "v1" {
		t.Errorf("expected APIVersion=v1, got %s", r.APIVersion)
	}
}

func TestParseYAMLMultiDocument(t *testing.T) {
	yamlContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deploy
  namespace: production
`

	resources, err := parseYAML(strings.NewReader(yamlContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resources))
	}

	if resources[0].Name != "config-1" {
		t.Errorf("expected config-1, got %s", resources[0].Name)
	}
	if resources[1].Name != "config-2" {
		t.Errorf("expected config-2, got %s", resources[1].Name)
	}
	if resources[2].Kind != "Deployment" {
		t.Errorf("expected Deployment, got %s", resources[2].Kind)
	}
	if resources[2].Namespace != "production" {
		t.Errorf("expected production, got %s", resources[2].Namespace)
	}
}

func TestParseYAMLSkipsEmptyDocuments(t *testing.T) {
	yamlContent := `
---
apiVersion: v1
kind: Service
metadata:
  name: svc
---
---
`

	resources, err := parseYAML(strings.NewReader(yamlContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
}

func TestParseYAMLSkipsDocumentsWithoutKind(t *testing.T) {
	yamlContent := `
apiVersion: v1
metadata:
  name: no-kind
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: has-kind
`

	resources, err := parseYAML(strings.NewReader(yamlContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].Name != "has-kind" {
		t.Errorf("expected has-kind, got %s", resources[0].Name)
	}
}

func TestParseYAMLInvalidContentSkipped(t *testing.T) {
	// Invalid YAML that can't be decoded as K8s resource is skipped
	resources, err := parseYAML(strings.NewReader(`{invalid: yaml: [}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 0 {
		t.Fatalf("expected 0 resources for invalid content, got %d", len(resources))
	}
}

func TestFileSourceLoadSingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deploy.yaml")
	content := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  namespace: staging
spec:
  replicas: 3
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	src := NewFileSource(path)
	resources, err := src.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	if resources[0].Kind != "Deployment" {
		t.Errorf("expected Deployment, got %s", resources[0].Kind)
	}
}

func TestFileSourceLoadDirectory(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"deploy.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy-1
`,
		"service.yml": `apiVersion: v1
kind: Service
metadata:
  name: svc-1
`,
		"readme.txt": `not a yaml file`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	src := NewFileSource(dir)
	resources, err := src.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources (skipping .txt), got %d", len(resources))
	}
}

func TestFileSourceLoadDirectoryNested(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-root
`), 0644)

	os.WriteFile(filepath.Join(subDir, "b.yaml"), []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-sub
`), 0644)

	src := NewFileSource(dir)
	resources, err := src.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources from nested dirs, got %d", len(resources))
	}
}

func TestFileSourceNonExistentPath(t *testing.T) {
	src := NewFileSource("/nonexistent/path")
	_, err := src.Load()
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}
