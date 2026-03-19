package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/somaz94/kube-diff/internal/diff"
	"github.com/somaz94/kube-diff/internal/report"
	"github.com/somaz94/kube-diff/internal/source"
	"github.com/somaz94/kube-diff/internal/testutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// mockFetcher implements cluster.ResourceFetcher for testing.
type mockFetcher struct {
	resources map[string]*unstructured.Unstructured
}

func (m *mockFetcher) Get(_ context.Context, apiVersion, kind, namespace, name string) (*unstructured.Unstructured, error) {
	key := fmt.Sprintf("%s/%s/%s/%s", apiVersion, kind, namespace, name)
	if obj, ok := m.resources[key]; ok {
		return obj, nil
	}
	return nil, fmt.Errorf("not found: %s", key)
}

func TestExecute(t *testing.T) {
	// Execute with no args should print help and succeed
	err := Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileCommandRequiresArg(t *testing.T) {
	rootCmd.SetArgs([]string{"file"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when file command called without args")
	}
}

func TestHelmCommandRequiresArg(t *testing.T) {
	rootCmd.SetArgs([]string{"helm"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when helm command called without args")
	}
}

func TestKustomizeCommandRequiresArg(t *testing.T) {
	rootCmd.SetArgs([]string{"kustomize"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when kustomize command called without args")
	}
}

func TestFileCommandWithInvalidPath(t *testing.T) {
	rootCmd.SetArgs([]string{"file", "/tmp/nonexistent-kube-diff-test"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestHelmCommandWithInvalidChart(t *testing.T) {
	rootCmd.SetArgs([]string{"helm", "/tmp/nonexistent-kube-diff-chart"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent chart")
	}
}

func TestKustomizeCommandWithInvalidPath(t *testing.T) {
	rootCmd.SetArgs([]string{"kustomize", "/tmp/nonexistent-kube-diff-overlay"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent overlay")
	}
}

func TestParseSelector(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  map[string]string
		wantError bool
	}{
		{"single label", "app=nginx", map[string]string{"app": "nginx"}, false},
		{"multiple labels", "app=nginx,env=prod", map[string]string{"app": "nginx", "env": "prod"}, false},
		{"with spaces", " app = nginx , env = prod ", map[string]string{"app": "nginx", "env": "prod"}, false},
		{"invalid no equals", "app", nil, true},
		{"empty key", "=value", nil, true},
		{"empty selector", "", nil, true},
		{"empty after trim", " , ", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSelector(tt.input)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected %s=%s, got %s=%s", k, v, k, result[k])
				}
			}
		})
	}
}

func TestMatchesLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		selector map[string]string
		nilObj   bool
		expected bool
	}{
		{
			name:     "matching single label",
			labels:   map[string]string{"app": "nginx"},
			selector: map[string]string{"app": "nginx"},
			expected: true,
		},
		{
			name:     "matching multiple labels",
			labels:   map[string]string{"app": "nginx", "env": "prod", "tier": "frontend"},
			selector: map[string]string{"app": "nginx", "env": "prod"},
			expected: true,
		},
		{
			name:     "non-matching label value",
			labels:   map[string]string{"app": "nginx"},
			selector: map[string]string{"app": "apache"},
			expected: false,
		},
		{
			name:     "missing label key",
			labels:   map[string]string{"app": "nginx"},
			selector: map[string]string{"env": "prod"},
			expected: false,
		},
		{
			name:     "nil object",
			nilObj:   true,
			selector: map[string]string{"app": "nginx"},
			expected: false,
		},
		{
			name:     "no labels on resource",
			labels:   nil,
			selector: map[string]string{"app": "nginx"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := source.Resource{}
			if !tt.nilObj {
				obj := &unstructured.Unstructured{}
				obj.SetLabels(tt.labels)
				r.Object = obj
			}
			if matchesLabels(r, tt.selector) != tt.expected {
				t.Errorf("expected matchesLabels()=%v", tt.expected)
			}
		})
	}
}

func TestMatchesLabelsSelectorFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("selector")
	if f == nil {
		t.Fatal("selector flag not found")
	}
	if f.DefValue != "" {
		t.Errorf("expected selector default='', got %s", f.DefValue)
	}
}

func TestRootCommandNewFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	tests := []struct {
		name     string
		flag     string
		defValue string
	}{
		{"name", "name", "[]"},
		{"ignore-field", "ignore-field", "[]"},
		{"context-lines", "context-lines", "3"},
		{"exit-code", "exit-code", "false"},
		{"diff-strategy", "diff-strategy", "live"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Fatalf("flag %q not found", tt.flag)
			}
			if f.DefValue != tt.defValue {
				t.Errorf("flag %q default: got %q, want %q", tt.flag, f.DefValue, tt.defValue)
			}
		})
	}
}

func TestCompareResourcesWithOptions(t *testing.T) {
	localObj := testutil.NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "new-value"},
	})
	clusterObj := testutil.NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "old-value"},
	})

	fetcher := &mockFetcher{
		resources: map[string]*unstructured.Unstructured{
			"v1/ConfigMap/default/my-cm": clusterObj,
		},
	}

	resources := []source.Resource{
		{APIVersion: "v1", Kind: "ConfigMap", Name: "my-cm", Namespace: "default", Object: localObj},
	}

	// With ignore-field, data should be ignored → unchanged
	opts := diff.CompareOptions{ContextLines: 3, IgnoreFields: []string{"data"}}
	results, err := compareResources(fetcher, resources, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != diff.StatusUnchanged {
		t.Errorf("expected unchanged with ignored field, got %s", results[0].Status)
	}
}

func TestRootCommandFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	tests := []struct {
		name     string
		flag     string
		defValue string
	}{
		{"kubeconfig", "kubeconfig", ""},
		{"context", "context", ""},
		{"namespace", "namespace", ""},
		{"selector", "selector", ""},
		{"summary-only", "summary-only", "false"},
		{"output", "output", "color"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Fatalf("flag %q not found", tt.flag)
			}
			if f.DefValue != tt.defValue {
				t.Errorf("flag %q default: got %q, want %q", tt.flag, f.DefValue, tt.defValue)
			}
		})
	}
}

func TestHelmCommandFlags(t *testing.T) {
	f := helmCmd.Flags().Lookup("values")
	if f == nil {
		t.Fatal("values flag not found on helm command")
	}

	r := helmCmd.Flags().Lookup("release")
	if r == nil {
		t.Fatal("release flag not found on helm command")
	}
	if r.DefValue != "release" {
		t.Errorf("expected release default=release, got %s", r.DefValue)
	}
}

func TestUnknownCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"unknown-cmd"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func writeTestYAML(t *testing.T, dir, filename, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestFileCommandNamespaceFilter(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "deploy.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  namespace: production
spec:
  replicas: 1
`)
	// Filter by non-matching namespace → "No resources found"
	rootCmd.SetArgs([]string{"file", dir, "-n", "staging"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileCommandKindFilter(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "deploy.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  namespace: default
spec:
  replicas: 1
`)
	// Filter by non-matching kind → "No resources found"
	rootCmd.SetArgs([]string{"file", dir, "-k", "Service"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileCommandNameFilter(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "resources.yaml", `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: default
data:
  key: value
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: other-config
  namespace: default
data:
  key: value
`)
	// Filter by non-matching name → "No resources found"
	rootCmd.SetArgs([]string{"file", dir, "-N", "nonexistent"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileCommandSelectorFilter(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "deploy.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  namespace: default
  labels:
    app: nginx
spec:
  replicas: 1
`)
	// Filter by non-matching selector → "No resources found"
	rootCmd.SetArgs([]string{"file", dir, "-l", "app=apache"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileCommandInvalidSelector(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "deploy.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 1
`)
	rootCmd.SetArgs([]string{"file", dir, "-l", "invalid-selector"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid selector")
	}
}

func TestWatchCommandExists(t *testing.T) {
	f := rootCmd.Commands()
	found := false
	for _, c := range f {
		if c.Name() == "watch" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("watch command not found")
	}
}

func TestWatchCommandRequiresArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"watch"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when watch command called without args")
	}
}

func TestWatchCommandIntervalFlag(t *testing.T) {
	f := watchCmd.Flags().Lookup("interval")
	if f == nil {
		t.Fatal("interval flag not found on watch command")
	}
	if f.DefValue != "0s" {
		t.Errorf("expected interval default=0s, got %s", f.DefValue)
	}
}

func TestCreateSourceFile(t *testing.T) {
	src := createSource(rootCmd, "file", "/tmp")
	if src == nil {
		t.Fatal("expected non-nil source for file type")
	}
}

func TestCreateSourceInvalid(t *testing.T) {
	src := createSource(rootCmd, "invalid", "/tmp")
	if src != nil {
		t.Fatal("expected nil source for invalid type")
	}
}

func TestIsRelevantChange(t *testing.T) {
	tests := []struct {
		name     string
		event    fsnotify.Event
		expected bool
	}{
		{"yaml write", fsnotify.Event{Name: "test.yaml", Op: fsnotify.Write}, true},
		{"yml create", fsnotify.Event{Name: "test.yml", Op: fsnotify.Create}, true},
		{"json write", fsnotify.Event{Name: "values.json", Op: fsnotify.Write}, true},
		{"txt write", fsnotify.Event{Name: "readme.txt", Op: fsnotify.Write}, false},
		{"yaml remove", fsnotify.Event{Name: "test.yaml", Op: fsnotify.Remove}, false},
		{"go write", fsnotify.Event{Name: "main.go", Op: fsnotify.Write}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isRelevantChange(tt.event) != tt.expected {
				t.Errorf("expected isRelevantChange=%v for %s", tt.expected, tt.name)
			}
		})
	}
}

func TestCompareResourcesWithStrategy(t *testing.T) {
	lastAppliedJSON := `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"my-cm","namespace":"default"},"data":{"key":"value"}}`

	localObj := testutil.NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})
	clusterObj := testutil.NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "cluster-modified"},
	})
	clusterObj.Object["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
		"kubectl.kubernetes.io/last-applied-configuration": lastAppliedJSON,
	}

	fetcher := &mockFetcher{
		resources: map[string]*unstructured.Unstructured{
			"v1/ConfigMap/default/my-cm": clusterObj,
		},
	}

	resources := []source.Resource{
		{APIVersion: "v1", Kind: "ConfigMap", Name: "my-cm", Namespace: "default", Object: localObj},
	}

	// live strategy → changed
	liveOpts := diff.CompareOptions{ContextLines: 3, Strategy: diff.StrategyLive}
	results, err := compareResources(fetcher, resources, liveOpts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != diff.StatusChanged {
		t.Errorf("live: expected changed, got %s", results[0].Status)
	}

	// last-applied strategy → unchanged
	lastAppliedOpts := diff.CompareOptions{ContextLines: 3, Strategy: diff.StrategyLastApplied}
	results, err = compareResources(fetcher, resources, lastAppliedOpts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != diff.StatusUnchanged {
		t.Errorf("last-applied: expected unchanged, got %s", results[0].Status)
	}
}

func TestCompareResourcesUnchanged(t *testing.T) {
	localObj := testutil.NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})
	clusterObj := testutil.NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})

	fetcher := &mockFetcher{
		resources: map[string]*unstructured.Unstructured{
			"v1/ConfigMap/default/my-cm": clusterObj,
		},
	}

	resources := []source.Resource{
		{APIVersion: "v1", Kind: "ConfigMap", Name: "my-cm", Namespace: "default", Object: localObj},
	}

	results, err := compareResources(fetcher, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != diff.StatusUnchanged {
		t.Errorf("expected unchanged, got %s", results[0].Status)
	}
}

func TestCompareResourcesChanged(t *testing.T) {
	localObj := testutil.NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "new-value"},
	})
	clusterObj := testutil.NewTestObj("v1", "ConfigMap", "my-cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "old-value"},
	})

	fetcher := &mockFetcher{
		resources: map[string]*unstructured.Unstructured{
			"v1/ConfigMap/default/my-cm": clusterObj,
		},
	}

	resources := []source.Resource{
		{APIVersion: "v1", Kind: "ConfigMap", Name: "my-cm", Namespace: "default", Object: localObj},
	}

	results, err := compareResources(fetcher, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != diff.StatusChanged {
		t.Errorf("expected changed, got %s", results[0].Status)
	}
	if results[0].Diff == "" {
		t.Error("expected non-empty diff")
	}
}

func TestCompareResourcesNew(t *testing.T) {
	localObj := testutil.NewTestObj("v1", "ConfigMap", "new-cm", "default", nil)

	fetcher := &mockFetcher{
		resources: map[string]*unstructured.Unstructured{},
	}

	resources := []source.Resource{
		{APIVersion: "v1", Kind: "ConfigMap", Name: "new-cm", Namespace: "default", Object: localObj},
	}

	results, err := compareResources(fetcher, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != diff.StatusNew {
		t.Errorf("expected new, got %s", results[0].Status)
	}
}

func TestBuildCompareOptions(t *testing.T) {
	tests := []struct {
		name     string
		flags    diffFlags
		expected diff.CompareOptions
	}{
		{
			name:  "default live strategy",
			flags: diffFlags{contextLines: 3, diffStrategy: "live"},
			expected: diff.CompareOptions{
				ContextLines: 3,
				Strategy:     diff.StrategyLive,
			},
		},
		{
			name:  "last-applied strategy",
			flags: diffFlags{contextLines: 5, diffStrategy: "last-applied", ignoreFields: []string{"status"}},
			expected: diff.CompareOptions{
				ContextLines: 5,
				Strategy:     diff.StrategyLastApplied,
				IgnoreFields: []string{"status"},
			},
		},
		{
			name:  "unknown strategy defaults to live",
			flags: diffFlags{contextLines: 3, diffStrategy: "unknown"},
			expected: diff.CompareOptions{
				ContextLines: 3,
				Strategy:     diff.StrategyLive,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildCompareOptions(tt.flags)
			if opts.ContextLines != tt.expected.ContextLines {
				t.Errorf("ContextLines: got %d, want %d", opts.ContextLines, tt.expected.ContextLines)
			}
			if opts.Strategy != tt.expected.Strategy {
				t.Errorf("Strategy: got %v, want %v", opts.Strategy, tt.expected.Strategy)
			}
			if len(opts.IgnoreFields) != len(tt.expected.IgnoreFields) {
				t.Errorf("IgnoreFields length: got %d, want %d", len(opts.IgnoreFields), len(tt.expected.IgnoreFields))
			}
		})
	}
}

func TestPrintReport(t *testing.T) {
	results := []*diff.DiffResult{
		{Kind: "ConfigMap", Name: "cm", Namespace: "default", Status: diff.StatusNew},
		{Kind: "Deployment", Name: "app", Namespace: "default", Status: diff.StatusChanged, Diff: "--- a\n+++ b\n-old\n+new"},
	}
	summary := report.NewSummary(results)

	tests := []struct {
		name     string
		flags    diffFlags
		contains string
	}{
		{"json output", diffFlags{output: "json"}, `"total":`},
		{"plain output", diffFlags{output: "plain"}, "* NEW"},
		{"markdown output", diffFlags{output: "markdown"}, "## kube-diff Report"},
		{"table output", diffFlags{output: "table"}, "STATUS"},
		{"color output (default)", diffFlags{output: "color"}, "NEW"},
		{"summary-only", diffFlags{summaryOnly: true}, "2 resources"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := printReport(&buf, summary, tt.flags); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output := buf.String(); !containsString(output, tt.contains) {
				t.Errorf("output missing %q:\n%s", tt.contains, output)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && bytes.Contains([]byte(s), []byte(substr))
}

func TestToStringSet(t *testing.T) {
	set := toStringSet([]string{"a", "b", "c"})
	if !set["a"] || !set["b"] || !set["c"] {
		t.Error("expected all items in set")
	}
	if set["d"] {
		t.Error("unexpected item in set")
	}

	empty := toStringSet([]string{})
	if len(empty) != 0 {
		t.Error("expected empty set")
	}
}

func TestCreateSourceAllTypes(t *testing.T) {
	tests := []struct {
		name       string
		sourceType string
		wantNil    bool
	}{
		{"file", "file", false},
		{"helm", "helm", false},
		{"kustomize", "kustomize", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := createSource(rootCmd, tt.sourceType, "/tmp")
			if tt.wantNil && src != nil {
				t.Error("expected nil source")
			}
			if !tt.wantNil && src == nil {
				t.Error("expected non-nil source")
			}
		})
	}
}

func TestAddWatchPaths(t *testing.T) {
	dir := t.TempDir()

	// Create subdirectories
	subDir := filepath.Join(dir, "manifests")
	hiddenDir := filepath.Join(dir, ".hidden")
	os.MkdirAll(subDir, 0755)
	os.MkdirAll(hiddenDir, 0755)

	// Write test files
	os.WriteFile(filepath.Join(dir, "deploy.yaml"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(subDir, "svc.yaml"), []byte("test"), 0644)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	if err := addWatchPaths(watcher, dir); err != nil {
		t.Fatalf("addWatchPaths failed: %v", err)
	}

	// Watcher should have added dir and subDir but not hiddenDir
	watchList := watcher.WatchList()
	found := map[string]bool{}
	for _, p := range watchList {
		found[p] = true
	}

	if !found[dir] {
		t.Error("expected root dir to be watched")
	}
	if !found[subDir] {
		t.Error("expected subDir to be watched")
	}
	if found[hiddenDir] {
		t.Error("hidden dir should not be watched")
	}
}

func TestAddWatchPathsInvalidPath(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	err = addWatchPaths(watcher, "/nonexistent-path-kube-diff-test")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestRunWatchInvalidSourceType(t *testing.T) {
	err := runWatch(rootCmd, "invalid", "/tmp", 0)
	if err == nil {
		t.Fatal("expected error for invalid source type")
	}
	if !containsString(err.Error(), "invalid source type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPrintReportWritesToWriter(t *testing.T) {
	results := []*diff.DiffResult{
		{Kind: "Service", Name: "svc", Namespace: "default", Status: diff.StatusUnchanged},
	}
	summary := report.NewSummary(results)

	var buf bytes.Buffer
	err := printReport(&buf, summary, diffFlags{output: "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestExecuteDiff(t *testing.T) {
	localObj := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})
	clusterObj := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})

	fetcher := &mockFetcher{
		resources: map[string]*unstructured.Unstructured{
			"v1/ConfigMap/default/cm": clusterObj,
		},
	}

	resources := []source.Resource{
		{APIVersion: "v1", Kind: "ConfigMap", Name: "cm", Namespace: "default", Object: localObj},
	}

	tests := []struct {
		name   string
		flags  diffFlags
		errNil bool
	}{
		{"json output", diffFlags{output: "json", contextLines: 3, noExitCode: true}, true},
		{"plain output", diffFlags{output: "plain", contextLines: 3, noExitCode: true}, true},
		{"markdown output", diffFlags{output: "markdown", contextLines: 3, noExitCode: true}, true},
		{"table output", diffFlags{output: "table", contextLines: 3, noExitCode: true}, true},
		{"color output", diffFlags{output: "color", contextLines: 3, noExitCode: true}, true},
		{"summary-only", diffFlags{summaryOnly: true, contextLines: 3, noExitCode: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := executeDiff(fetcher, resources, tt.flags, &buf)
			if tt.errNil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if buf.Len() == 0 {
				t.Error("expected non-empty output")
			}
		})
	}
}

func TestExecuteDiffWithChanges(t *testing.T) {
	localObj := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "new-value"},
	})
	clusterObj := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "old-value"},
	})

	fetcher := &mockFetcher{
		resources: map[string]*unstructured.Unstructured{
			"v1/ConfigMap/default/cm": clusterObj,
		},
	}

	resources := []source.Resource{
		{APIVersion: "v1", Kind: "ConfigMap", Name: "cm", Namespace: "default", Object: localObj},
	}

	// With noExitCode=true, should not call os.Exit
	var buf bytes.Buffer
	err := executeDiff(fetcher, resources, diffFlags{output: "json", contextLines: 3, noExitCode: true}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestExecuteDiffLastApplied(t *testing.T) {
	lastAppliedJSON := `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"cm","namespace":"default"},"data":{"key":"value"}}`

	localObj := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "value"},
	})
	clusterObj := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "modified"},
	})
	clusterObj.Object["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
		"kubectl.kubernetes.io/last-applied-configuration": lastAppliedJSON,
	}

	fetcher := &mockFetcher{
		resources: map[string]*unstructured.Unstructured{
			"v1/ConfigMap/default/cm": clusterObj,
		},
	}

	resources := []source.Resource{
		{APIVersion: "v1", Kind: "ConfigMap", Name: "cm", Namespace: "default", Object: localObj},
	}

	var buf bytes.Buffer
	err := executeDiff(fetcher, resources, diffFlags{
		output:       "plain",
		contextLines: 3,
		diffStrategy: "last-applied",
		noExitCode:   true,
	}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompareResourcesMultiple(t *testing.T) {
	cm := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", map[string]interface{}{
		"data": map[string]interface{}{"k": "v"},
	})
	deploy := testutil.NewTestObj("apps/v1", "Deployment", "app", "default", nil)

	fetcher := &mockFetcher{
		resources: map[string]*unstructured.Unstructured{
			"v1/ConfigMap/default/cm": cm,
		},
	}

	resources := []source.Resource{
		{APIVersion: "v1", Kind: "ConfigMap", Name: "cm", Namespace: "default", Object: cm},
		{APIVersion: "apps/v1", Kind: "Deployment", Name: "app", Namespace: "default", Object: deploy},
	}

	results, err := compareResources(fetcher, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Status != diff.StatusUnchanged {
		t.Errorf("expected cm unchanged, got %s", results[0].Status)
	}
	if results[1].Status != diff.StatusNew {
		t.Errorf("expected deploy new, got %s", results[1].Status)
	}
}
