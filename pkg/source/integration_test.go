//go:build integration

package source

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

func helmAvailable() bool {
	_, err := exec.LookPath("helm")
	return err == nil
}

func kustomizeAvailable() bool {
	_, err := exec.LookPath("kustomize")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("kubectl")
	return err == nil
}

func TestHelmSourceLoadSuccess(t *testing.T) {
	if !helmAvailable() {
		t.Skip("helm not installed")
	}

	chartPath := filepath.Join(testdataDir(t), "helm", "test-chart")
	src := NewHelmSource(chartPath, "test-release", nil)
	resources, err := src.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) < 2 {
		t.Fatalf("expected at least 2 resources, got %d", len(resources))
	}

	kinds := map[string]bool{}
	for _, r := range resources {
		kinds[r.Kind] = true
	}
	if !kinds["Deployment"] {
		t.Error("expected Deployment resource from helm template")
	}
	if !kinds["Service"] {
		t.Error("expected Service resource from helm template")
	}
}

func TestHelmSourceLoadWithValues(t *testing.T) {
	if !helmAvailable() {
		t.Skip("helm not installed")
	}

	chartPath := filepath.Join(testdataDir(t), "helm", "test-chart")
	valuesFile := filepath.Join(testdataDir(t), "helm", "custom-values.yaml")
	src := NewHelmSource(chartPath, "custom", []string{valuesFile})
	resources, err := src.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) < 2 {
		t.Fatalf("expected at least 2 resources, got %d", len(resources))
	}

	// Check release name is applied
	for _, r := range resources {
		if r.Kind == "Deployment" && r.Name != "custom-app" {
			t.Errorf("expected deployment name custom-app, got %s", r.Name)
		}
	}
}

func TestHelmSourceLoadReleaseName(t *testing.T) {
	if !helmAvailable() {
		t.Skip("helm not installed")
	}

	chartPath := filepath.Join(testdataDir(t), "helm", "test-chart")
	src := NewHelmSource(chartPath, "my-release", nil)
	resources, err := src.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, r := range resources {
		if r.Kind == "Service" && r.Name == "my-release-svc" {
			found = true
		}
	}
	if !found {
		t.Error("expected service named my-release-svc")
	}
}

func TestKustomizeSourceLoadSuccess(t *testing.T) {
	if !kustomizeAvailable() {
		t.Skip("kustomize/kubectl not installed")
	}

	kustomizePath := filepath.Join(testdataDir(t), "kustomize", "base")
	src := NewKustomizeSource(kustomizePath)
	resources, err := src.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	kinds := map[string]bool{}
	for _, r := range resources {
		kinds[r.Kind] = true
	}
	if !kinds["Deployment"] {
		t.Error("expected Deployment resource from kustomize build")
	}
	if !kinds["Service"] {
		t.Error("expected Service resource from kustomize build")
	}
}

func TestKustomizeSourceLoadResourceNames(t *testing.T) {
	if !kustomizeAvailable() {
		t.Skip("kustomize/kubectl not installed")
	}

	kustomizePath := filepath.Join(testdataDir(t), "kustomize", "base")
	src := NewKustomizeSource(kustomizePath)
	resources, err := src.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	names := map[string]bool{}
	for _, r := range resources {
		names[r.Name] = true
	}
	if !names["kustomize-app"] {
		t.Error("expected resource named kustomize-app")
	}
	if !names["kustomize-svc"] {
		t.Error("expected resource named kustomize-svc")
	}
}

func TestHelmSourceLoadNonExistentValues(t *testing.T) {
	if !helmAvailable() {
		t.Skip("helm not installed")
	}

	chartPath := filepath.Join(testdataDir(t), "helm", "test-chart")
	src := NewHelmSource(chartPath, "test", []string{"/nonexistent/values.yaml"})
	_, err := src.Load()
	if err == nil {
		t.Fatal("expected error for nonexistent values file")
	}
}

func TestFileSourceLoadTestdata(t *testing.T) {
	// Test loading actual testdata YAML files
	kustomizeBase := filepath.Join(testdataDir(t), "kustomize", "base")

	// Only load deployment.yaml
	deployPath := filepath.Join(kustomizeBase, "deployment.yaml")
	if _, err := os.Stat(deployPath); err != nil {
		t.Skipf("testdata not found: %v", err)
	}

	src := NewFileSource(deployPath)
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
	if resources[0].Name != "kustomize-app" {
		t.Errorf("expected kustomize-app, got %s", resources[0].Name)
	}
}
