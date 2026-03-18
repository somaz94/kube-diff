package source

import (
	"testing"
)

func TestNewKustomizeSource(t *testing.T) {
	k := NewKustomizeSource("./overlays/prod")

	if k.Path != "./overlays/prod" {
		t.Errorf("expected Path=./overlays/prod, got %s", k.Path)
	}
}

func TestKustomizeSourceLoadFailsWithMissingPath(t *testing.T) {
	k := NewKustomizeSource("/nonexistent/path")
	_, err := k.Load()
	if err == nil {
		t.Fatal("expected error for missing kustomize path")
	}
}

func TestKustomizeSourceLoadFailsWithInvalidPath(t *testing.T) {
	dir := t.TempDir()
	k := NewKustomizeSource(dir)
	_, err := k.Load()
	if err == nil {
		t.Fatal("expected error for directory without kustomization.yaml")
	}
}
