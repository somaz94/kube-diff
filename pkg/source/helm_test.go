package source

import (
	"testing"
)

func TestNewHelmSource(t *testing.T) {
	h := NewHelmSource("./chart", "my-release", []string{"values.yaml", "values-prod.yaml"})

	if h.ChartPath != "./chart" {
		t.Errorf("expected ChartPath=./chart, got %s", h.ChartPath)
	}
	if h.ReleaseName != "my-release" {
		t.Errorf("expected ReleaseName=my-release, got %s", h.ReleaseName)
	}
	if len(h.ValuesFiles) != 2 {
		t.Fatalf("expected 2 values files, got %d", len(h.ValuesFiles))
	}
	if h.ValuesFiles[0] != "values.yaml" {
		t.Errorf("expected values.yaml, got %s", h.ValuesFiles[0])
	}
}

func TestNewHelmSourceNoValues(t *testing.T) {
	h := NewHelmSource("./chart", "release", nil)

	if h.ChartPath != "./chart" {
		t.Errorf("expected ChartPath=./chart, got %s", h.ChartPath)
	}
	if h.ValuesFiles != nil {
		t.Errorf("expected nil values files, got %v", h.ValuesFiles)
	}
}

func TestHelmSourceLoadFailsWithMissingChart(t *testing.T) {
	h := NewHelmSource("/nonexistent/chart/path", "test", nil)
	_, err := h.Load()
	if err == nil {
		t.Fatal("expected error for missing chart path")
	}
}

func TestHelmSourceLoadFailsWithInvalidChart(t *testing.T) {
	dir := t.TempDir()
	h := NewHelmSource(dir, "test", nil)
	_, err := h.Load()
	if err == nil {
		t.Fatal("expected error for invalid chart directory")
	}
}
