package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/somaz94/kube-diff/internal/testutil"
	"github.com/somaz94/kube-diff/pkg/diff"
	"github.com/somaz94/kube-diff/pkg/source"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// fakeFetcher implements cluster.ResourceFetcher. A name in errs fails with
// that error, a name in objs is returned as-is, and anything else gets the
// NotFound a real API server returns for a missing object.
type fakeFetcher struct {
	objs map[string]*unstructured.Unstructured
	errs map[string]error
}

func (f *fakeFetcher) Get(_ context.Context, _, _, _, name string) (*unstructured.Unstructured, error) {
	if err, ok := f.errs[name]; ok {
		return nil, err
	}
	if obj, ok := f.objs[name]; ok {
		return obj, nil
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, name)
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

func TestCompare_AbsentErrorsReportNew(t *testing.T) {
	configMaps := schema.GroupResource{Resource: "configmaps"}
	tests := []struct {
		name string
		err  error
	}{
		{"not found", apierrors.NewNotFound(configMaps, "cm")},
		{"wrapped not found", fmt.Errorf("fetch: %w", apierrors.NewNotFound(configMaps, "cm"))},
		{"kind not served", &meta.NoKindMatchError{
			GroupKind:        schema.GroupKind{Group: "example.com", Kind: "Widget"},
			SearchedVersions: []string{"v1"},
		}},
		{"wrapped kind not served", fmt.Errorf("map: %w", &meta.NoKindMatchError{
			GroupKind:        schema.GroupKind{Group: "example.com", Kind: "Widget"},
			SearchedVersions: []string{"v1"},
		})},
		{"resource not served", &meta.NoResourceMatchError{
			PartialResource: schema.GroupVersionResource{Group: "example.com", Version: "v1", Resource: "widgets"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", nil)
			fetcher := &fakeFetcher{errs: map[string]error{"cm": tt.err}}

			results, err := Compare(context.Background(), fetcher, []source.Resource{resource("cm", obj)}, diff.DefaultCompareOptions())
			if err != nil {
				t.Fatalf("Compare returned error: %v", err)
			}
			if len(results) != 1 || results[0].Status != diff.StatusNew {
				t.Fatalf("want 1 new result, got %+v", results)
			}
		})
	}
}

func TestCompare_OtherFetchErrorsAreReturned(t *testing.T) {
	errRefused := errors.New("connection refused")
	tests := []struct {
		name      string
		obj       *unstructured.Unstructured
		err       error
		wantID    string
		wantMatch func(error) bool
	}{
		{
			name:      "forbidden",
			obj:       testutil.NewTestObj("apps/v1", "Deployment", "app", "default", nil),
			err:       apierrors.NewForbidden(schema.GroupResource{Group: "apps", Resource: "deployments"}, "app", errors.New("RBAC denied")),
			wantID:    "get apps/v1 Deployment default/app: ",
			wantMatch: apierrors.IsForbidden,
		},
		{
			name:      "server timeout",
			obj:       testutil.NewTestObj("v1", "ConfigMap", "app", "default", nil),
			err:       apierrors.NewTimeoutError("request timed out", 1),
			wantID:    "get v1 ConfigMap default/app: ",
			wantMatch: apierrors.IsTimeout,
		},
		{
			name:      "generic error on a cluster-scoped kind",
			obj:       testutil.NewTestObj("v1", "Namespace", "app", "", nil),
			err:       errRefused,
			wantID:    "get v1 Namespace app: ",
			wantMatch: func(err error) bool { return errors.Is(err, errRefused) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// "first" resolves fine, proving an error mid-list still aborts the run.
			first := testutil.NewTestObj("v1", "ConfigMap", "first", "default", nil)
			fetcher := &fakeFetcher{
				objs: map[string]*unstructured.Unstructured{"first": first},
				errs: map[string]error{"app": tt.err},
			}
			resources := []source.Resource{resource("first", first), resource("app", tt.obj)}

			results, err := Compare(context.Background(), fetcher, resources, diff.DefaultCompareOptions())
			if err == nil {
				t.Fatalf("want error, got results %+v", results)
			}
			if results != nil {
				t.Errorf("want nil results on error, got %+v", results)
			}
			if !tt.wantMatch(err) {
				t.Errorf("error %v does not unwrap to the fetch error", err)
			}
			if !strings.HasPrefix(err.Error(), tt.wantID) {
				t.Errorf("error %q, want prefix %q", err, tt.wantID)
			}
		})
	}
}

func TestRun_FetchError(t *testing.T) {
	obj := testutil.NewTestObj("v1", "ConfigMap", "cm", "default", nil)
	src := &fakeSource{resources: []source.Resource{resource("cm", obj)}}
	fetcher := &fakeFetcher{errs: map[string]error{
		"cm": apierrors.NewForbidden(schema.GroupResource{Resource: "configmaps"}, "cm", errors.New("RBAC denied")),
	}}

	_, err := Run(context.Background(), src, fetcher, diff.DefaultCompareOptions())
	if !apierrors.IsForbidden(err) {
		t.Fatalf("want Forbidden to propagate from Run, got %v", err)
	}
}
