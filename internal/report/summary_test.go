package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/somaz94/kube-diff/internal/diff"
)

func makeResults() []*diff.DiffResult {
	return []*diff.DiffResult{
		{Kind: "Deployment", Name: "app-1", Namespace: "default", Status: diff.StatusChanged, Diff: "--- cluster\n+++ local\n@@ -1 +1 @@\n-replicas: 2\n+replicas: 3"},
		{Kind: "Service", Name: "svc-1", Namespace: "default", Status: diff.StatusUnchanged},
		{Kind: "ConfigMap", Name: "cm-new", Namespace: "default", Status: diff.StatusNew},
		{Kind: "Secret", Name: "old-secret", Namespace: "default", Status: diff.StatusDeleted},
	}
}

func TestNewSummaryCounts(t *testing.T) {
	s := NewSummary(makeResults())

	if s.Total != 4 {
		t.Errorf("expected Total=4, got %d", s.Total)
	}
	if s.Changed != 1 {
		t.Errorf("expected Changed=1, got %d", s.Changed)
	}
	if s.Unchanged != 1 {
		t.Errorf("expected Unchanged=1, got %d", s.Unchanged)
	}
	if s.New != 1 {
		t.Errorf("expected New=1, got %d", s.New)
	}
	if s.Deleted != 1 {
		t.Errorf("expected Deleted=1, got %d", s.Deleted)
	}
}

func TestHasChanges(t *testing.T) {
	tests := []struct {
		name     string
		results  []*diff.DiffResult
		expected bool
	}{
		{
			name:     "has changes",
			results:  makeResults(),
			expected: true,
		},
		{
			name: "no changes",
			results: []*diff.DiffResult{
				{Kind: "Service", Name: "svc", Status: diff.StatusUnchanged},
			},
			expected: false,
		},
		{
			name: "new resource counts as change",
			results: []*diff.DiffResult{
				{Kind: "ConfigMap", Name: "cm", Status: diff.StatusNew},
			},
			expected: true,
		},
		{
			name: "deleted resource counts as change",
			results: []*diff.DiffResult{
				{Kind: "Secret", Name: "s", Status: diff.StatusDeleted},
			},
			expected: true,
		},
		{
			name:     "empty results",
			results:  []*diff.DiffResult{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSummary(tt.results)
			if s.HasChanges() != tt.expected {
				t.Errorf("expected HasChanges()=%v, got %v", tt.expected, s.HasChanges())
			}
		})
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		results  []*diff.DiffResult
		expected int
	}{
		{
			name:     "changes exist",
			results:  makeResults(),
			expected: 1,
		},
		{
			name: "no changes",
			results: []*diff.DiffResult{
				{Kind: "Service", Name: "svc", Status: diff.StatusUnchanged},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSummary(tt.results)
			if s.ExitCode() != tt.expected {
				t.Errorf("expected ExitCode()=%d, got %d", tt.expected, s.ExitCode())
			}
		})
	}
}

func TestPrintColorOutput(t *testing.T) {
	s := NewSummary(makeResults())
	var buf bytes.Buffer
	s.PrintColor(&buf)
	output := buf.String()

	// Check status markers
	if !strings.Contains(output, "NEW") {
		t.Error("expected NEW marker in output")
	}
	if !strings.Contains(output, "CHANGED") {
		t.Error("expected CHANGED marker in output")
	}
	if !strings.Contains(output, "OK") {
		t.Error("expected OK marker in output")
	}
	if !strings.Contains(output, "DELETED") {
		t.Error("expected DELETED marker in output")
	}

	// Check resource names
	if !strings.Contains(output, "app-1") {
		t.Error("expected resource name app-1 in output")
	}

	// Check summary line
	if !strings.Contains(output, "4 resources") {
		t.Error("expected summary with total resources")
	}
}

func TestPrintColorShowsDiff(t *testing.T) {
	results := []*diff.DiffResult{
		{Kind: "Deployment", Name: "app", Status: diff.StatusChanged, Diff: "--- cluster\n+++ local\n-old\n+new"},
	}
	s := NewSummary(results)
	var buf bytes.Buffer
	s.PrintColor(&buf)
	output := buf.String()

	if !strings.Contains(output, "old") || !strings.Contains(output, "new") {
		t.Error("expected diff content in color output")
	}
}

func TestPrintJSON(t *testing.T) {
	s := NewSummary(makeResults())
	var buf bytes.Buffer
	if err := s.PrintJSON(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if report["total"].(float64) != 4 {
		t.Errorf("expected total=4, got %v", report["total"])
	}
	if report["changed"].(float64) != 1 {
		t.Errorf("expected changed=1, got %v", report["changed"])
	}
	if report["new"].(float64) != 1 {
		t.Errorf("expected new=1, got %v", report["new"])
	}
	if report["deleted"].(float64) != 1 {
		t.Errorf("expected deleted=1, got %v", report["deleted"])
	}

	resources := report["resources"].([]interface{})
	if len(resources) != 4 {
		t.Errorf("expected 4 resources in JSON, got %d", len(resources))
	}

	// Verify first resource structure
	first := resources[0].(map[string]interface{})
	if first["kind"] != "Deployment" {
		t.Errorf("expected kind=Deployment, got %v", first["kind"])
	}
	if first["status"] != "changed" {
		t.Errorf("expected status=changed, got %v", first["status"])
	}
}

func TestPrintJSONEmptyResults(t *testing.T) {
	s := NewSummary([]*diff.DiffResult{})
	var buf bytes.Buffer
	if err := s.PrintJSON(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if report["total"].(float64) != 0 {
		t.Errorf("expected total=0, got %v", report["total"])
	}
}

func TestColorizeDiff(t *testing.T) {
	input := "--- cluster\n+++ local\n@@ -1 +1 @@\n-old line\n+new line\n context"
	result := colorizeDiff(input)

	// Lines starting with - should have red color code
	if !strings.Contains(result, "\033[31m-old line") {
		t.Error("expected red color for removed lines")
	}
	// Lines starting with + should have green color code
	if !strings.Contains(result, "\033[32m+new line") {
		t.Error("expected green color for added lines")
	}
	// Lines starting with @@ should have cyan color code
	if !strings.Contains(result, "\033[36m@@") {
		t.Error("expected cyan color for hunk headers")
	}
}

func TestPrintColorDeletedResource(t *testing.T) {
	results := []*diff.DiffResult{
		{Kind: "Secret", Name: "old-secret", Namespace: "prod", Status: diff.StatusDeleted},
	}
	s := NewSummary(results)
	var buf bytes.Buffer
	s.PrintColor(&buf)
	output := buf.String()

	if !strings.Contains(output, "DELETED") {
		t.Error("expected DELETED marker in output")
	}
	if !strings.Contains(output, "old-secret") {
		t.Error("expected resource name in output")
	}
	if !strings.Contains(output, "1 deleted") {
		t.Error("expected deleted count in summary")
	}
}

func TestPrintColorOnlyUnchanged(t *testing.T) {
	results := []*diff.DiffResult{
		{Kind: "Service", Name: "svc-1", Status: diff.StatusUnchanged},
		{Kind: "Service", Name: "svc-2", Status: diff.StatusUnchanged},
	}
	s := NewSummary(results)
	var buf bytes.Buffer
	s.PrintColor(&buf)
	output := buf.String()

	if !strings.Contains(output, "2 resources") {
		t.Error("expected 2 resources in summary")
	}
	if !strings.Contains(output, "2 unchanged") {
		t.Error("expected 2 unchanged in summary")
	}
	if strings.Contains(output, "changed") && !strings.Contains(output, "unchanged") {
		t.Error("should not show 'changed' when only unchanged exist")
	}
}

func TestPrintPlain(t *testing.T) {
	s := NewSummary(makeResults())
	var buf bytes.Buffer
	s.PrintPlain(&buf)
	output := buf.String()

	if !strings.Contains(output, "* NEW") {
		t.Error("expected * NEW marker in plain output")
	}
	if !strings.Contains(output, "~ CHANGED") {
		t.Error("expected ~ CHANGED marker in plain output")
	}
	if !strings.Contains(output, "OK") {
		t.Error("expected OK marker in plain output")
	}
	if !strings.Contains(output, "x DELETED") {
		t.Error("expected x DELETED marker in plain output")
	}
	// Should not contain ANSI escape codes
	if strings.Contains(output, "\033[") {
		t.Error("plain output should not contain ANSI codes")
	}
}

func TestPrintMarkdown(t *testing.T) {
	s := NewSummary(makeResults())
	var buf bytes.Buffer
	s.PrintMarkdown(&buf)
	output := buf.String()

	if !strings.Contains(output, "## kube-diff Report") {
		t.Error("expected markdown header")
	}
	if !strings.Contains(output, "| Status | Resource | Namespace |") {
		t.Error("expected markdown table header")
	}
	if !strings.Contains(output, "CHANGED") {
		t.Error("expected CHANGED in markdown")
	}
	if !strings.Contains(output, "```diff") {
		t.Error("expected diff code block in markdown")
	}
}

func TestPrintMarkdownNoDiffs(t *testing.T) {
	results := []*diff.DiffResult{
		{Kind: "Service", Name: "svc", Namespace: "default", Status: diff.StatusUnchanged},
	}
	s := NewSummary(results)
	var buf bytes.Buffer
	s.PrintMarkdown(&buf)
	output := buf.String()

	if strings.Contains(output, "```diff") {
		t.Error("should not have diff block when no changes")
	}
}

func TestPrintSummaryOnly(t *testing.T) {
	s := NewSummary(makeResults())
	var buf bytes.Buffer
	s.PrintSummaryOnly(&buf)
	output := buf.String()

	if !strings.Contains(output, "4 resources") {
		t.Error("expected resource count in summary")
	}
	if !strings.Contains(output, "1 changed") {
		t.Error("expected changed count")
	}
	if !strings.Contains(output, "1 new") {
		t.Error("expected new count")
	}
	if !strings.Contains(output, "1 deleted") {
		t.Error("expected deleted count")
	}
}

func TestPrintPlainClusterScoped(t *testing.T) {
	results := []*diff.DiffResult{
		{Kind: "ClusterRole", Name: "admin", Status: diff.StatusNew},
	}
	s := NewSummary(results)
	var buf bytes.Buffer
	s.PrintPlain(&buf)
	output := buf.String()

	if !strings.Contains(output, "ClusterRole/admin") {
		t.Error("expected ClusterRole/admin without namespace")
	}
}

func TestPrintJSONResourceFields(t *testing.T) {
	results := []*diff.DiffResult{
		{Kind: "Deployment", Name: "app", Namespace: "staging", Status: diff.StatusChanged},
		{Kind: "ClusterRole", Name: "admin", Status: diff.StatusNew},
	}
	s := NewSummary(results)
	var buf bytes.Buffer
	if err := s.PrintJSON(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var report map[string]interface{}
	json.Unmarshal(buf.Bytes(), &report)

	resources := report["resources"].([]interface{})
	// Namespaced resource
	first := resources[0].(map[string]interface{})
	if first["namespace"] != "staging" {
		t.Errorf("expected namespace=staging, got %v", first["namespace"])
	}
	// Cluster-scoped resource should omit namespace
	second := resources[1].(map[string]interface{})
	if _, ok := second["namespace"]; ok && second["namespace"] != "" {
		t.Error("expected empty/omitted namespace for cluster-scoped resource")
	}
}
