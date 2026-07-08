package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/somaz94/kube-diff/internal/testutil"
	"github.com/somaz94/kube-diff/pkg/diff"
	"github.com/somaz94/kube-diff/pkg/source"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// fakeFetcher implements cluster.ResourceFetcher. A resource keyed by name is
// returned as-is; anything else returns notFound to exercise the "new" path.
type fakeFetcher struct {
	objs map[string]*unstructured.Unstructured
}

func (f *fakeFetcher) Get(_ context.Context, _, _, _, name string) (*unstructured.Unstructured, error) {
	if obj, ok := f.objs[name]; ok {
		return obj, nil
	}
	return nil, errors.New("not found")
}

// fakeSource implements source.Source with a fixed result or error.
type fakeSource struct {
	resources []source.Resource
	err       error
}

func (s *fakeSource) Load() ([]source.Resource, error) {
	return s.resources, s.err
}

func resource(name string, obj *unstructured.Unstructured) source.Resource {
	return source.Resource{
		APIVersion: obj.GetAPIVersion(),
		Kind:       obj.GetKind(),
		Name:       name,
		Namespace:  obj.GetNamespace(),
		Object:     obj,
	}
}

func TestCompare_NewChangedUnchanged(t *testing.T) {
	live := testutil.NewTestObj("v1", "ConfigMap", "existing", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "old"},
	})
	local := testutil.NewTestObj("v1", "ConfigMap", "existing", "default", map[string]interface{}{
		"data": map[string]interface{}{"key": "new"},
	})
	same := testutil.NewTestObj("v1", "ConfigMap", "same", "default", nil)
	missing := testutil.NewTestObj("v1", "ConfigMap", "missing", "default", nil)

	fetcher := &fakeFetcher{objs: map[string]*unstructured.Unstructured{
		"existing": live,
		"same":     same,
	}}
	resources := []source.Resource{
		resource("existing", local),
		resource("same", same),
		resource("missing", missing),
	}

	results, err := Compare(context.Background(), fetcher, resources, diff.DefaultCompareOptions())
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	byName := map[string]*diff.Result{}
	for _, r := range results {
		byName[r.Name] = r
	}
	if got := byName["existing"].Status; got != diff.StatusChanged {
		t.Errorf("existing: want %q, got %q", diff.StatusChanged, got)
	}
	if got := byName["same"].Status; got != diff.StatusUnchanged {
		t.Errorf("same: want %q, got %q", diff.StatusUnchanged, got)
	}
	if got := byName["missing"].Status; got != diff.StatusNew {
		t.Errorf("missing: want %q, got %q", diff.StatusNew, got)
	}
}

func TestCompare_Empty(t *testing.T) {
	results, err := Compare(context.Background(), &fakeFetcher{}, nil, diff.DefaultCompareOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestRun_LoadsThenCompares(t *testing.T) {
	obj := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", nil)
	src := &fakeSource{resources: []source.Resource{resource("cm", obj)}}
	fetcher := &fakeFetcher{} // empty → cm is new

	results, err := Run(context.Background(), src, fetcher, diff.DefaultCompareOptions())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(results) != 1 || results[0].Status != diff.StatusNew {
		t.Fatalf("want 1 new result, got %+v", results)
	}
}

func TestRun_LoadError(t *testing.T) {
	src := &fakeSource{err: errors.New("load boom")}
	_, err := Run(context.Background(), src, &fakeFetcher{}, diff.DefaultCompareOptions())
	if err == nil {
		t.Fatal("expected load error to propagate, got nil")
	}
}
