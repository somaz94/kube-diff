package diff

import (
	"fmt"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DiffStatus represents the status of a resource comparison.
type DiffStatus string

const (
	StatusUnchanged DiffStatus = "unchanged"
	StatusChanged   DiffStatus = "changed"
	StatusNew       DiffStatus = "new"
	StatusDeleted   DiffStatus = "deleted"
)

// CompareOptions holds options for the comparison.
type CompareOptions struct {
	ContextLines int      // number of context lines in diff output (default: 3)
	IgnoreFields []string // field paths to ignore (e.g., "metadata.annotations.some-key")
}

// DefaultCompareOptions returns the default comparison options.
func DefaultCompareOptions() CompareOptions {
	return CompareOptions{
		ContextLines: 3,
	}
}

// DiffResult holds the comparison result for a single resource.
type DiffResult struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
	Status     DiffStatus
	Diff       string // unified diff text
}

// ResourceKey returns a human-readable identifier for the resource.
func (d *DiffResult) ResourceKey() string {
	if d.Namespace != "" {
		return fmt.Sprintf("%s/%s (namespace: %s)", d.Kind, d.Name, d.Namespace)
	}
	return fmt.Sprintf("%s/%s", d.Kind, d.Name)
}

// Compare compares a local resource against a cluster resource.
func Compare(local, cluster *unstructured.Unstructured, opts ...CompareOptions) (*DiffResult, error) {
	opt := DefaultCompareOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}

	result := &DiffResult{
		APIVersion: local.GetAPIVersion(),
		Kind:       local.GetKind(),
		Name:       local.GetName(),
		Namespace:  local.GetNamespace(),
	}

	if cluster == nil {
		result.Status = StatusNew
		return result, nil
	}

	normalizedLocal := Normalize(local)
	normalizedCluster := Normalize(cluster)

	// Remove user-specified ignore fields
	if len(opt.IgnoreFields) > 0 {
		RemoveFields(normalizedLocal, opt.IgnoreFields)
		RemoveFields(normalizedCluster, opt.IgnoreFields)
	}

	localYAML, err := toYAML(normalizedLocal)
	if err != nil {
		return nil, fmt.Errorf("error marshaling local resource: %w", err)
	}

	clusterYAML, err := toYAML(normalizedCluster)
	if err != nil {
		return nil, fmt.Errorf("error marshaling cluster resource: %w", err)
	}

	if localYAML == clusterYAML {
		result.Status = StatusUnchanged
		return result, nil
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(clusterYAML),
		B:        difflib.SplitLines(localYAML),
		FromFile: "cluster",
		ToFile:   "local",
		Context:  opt.ContextLines,
	}

	diffText, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return nil, fmt.Errorf("error generating diff: %w", err)
	}

	result.Status = StatusChanged
	result.Diff = diffText

	return result, nil
}

func toYAML(obj *unstructured.Unstructured) (string, error) {
	data, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
