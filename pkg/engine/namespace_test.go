package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/somaz94/kube-diff/internal/testutil"
	"github.com/somaz94/kube-diff/pkg/cluster"
	"github.com/somaz94/kube-diff/pkg/source"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// fakeResolver answers Namespaced from scopes, keyed by kind, and fails with
// errs[kind] when set.
type fakeResolver struct {
	scopes map[string]bool
	errs   map[string]error
}

func (r *fakeResolver) Namespaced(_ context.Context, _, kind string) (bool, error) {
	if err, ok := r.errs[kind]; ok {
		return false, err
	}
	return r.scopes[kind], nil
}

func TestDefaultNamespaceFillsOnlyNamespacedKinds(t *testing.T) {
	deploy := testutil.NewTestObj("apps/v1", "Deployment", "web", "", nil)
	role := testutil.NewTestObj("rbac.authorization.k8s.io/v1", "ClusterRole", "reader", "", nil)
	pinned := testutil.NewTestObj("v1", "ConfigMap", "cfg", "other", nil)
	widget := testutil.NewTestObj("example.com/v1", "Widget", "w", "", nil)
	resources := []source.Resource{
		resource("web", deploy), resource("reader", role), resource("cfg", pinned), resource("w", widget),
	}
	resolver := &fakeResolver{
		scopes: map[string]bool{"Deployment": true, "ClusterRole": false, "ConfigMap": true},
		errs: map[string]error{"Widget": &meta.NoKindMatchError{
			GroupKind: schema.GroupKind{Group: "example.com", Kind: "Widget"}, SearchedVersions: []string{"v1"},
		}},
	}

	got, err := DefaultNamespace(context.Background(), resolver, resources, "team-a")
	if err != nil {
		t.Fatalf("DefaultNamespace() error = %v", err)
	}

	want := []string{"team-a", "", "other", ""}
	for i, ns := range want {
		if got[i].Namespace != ns {
			t.Errorf("%s: Namespace = %q, want %q", got[i].Kind, got[i].Namespace, ns)
		}
		if got[i].Object.GetNamespace() != ns {
			t.Errorf("%s: object namespace = %q, want %q", got[i].Kind, got[i].Object.GetNamespace(), ns)
		}
	}
	if deploy.GetNamespace() != "" || resources[0].Namespace != "" {
		t.Error("DefaultNamespace modified its input")
	}
}

func TestDefaultNamespaceSkipsUnknownScope(t *testing.T) {
	cm := testutil.NewTestObj("v1", "ConfigMap", "cfg", "", nil)
	resolver := &fakeResolver{errs: map[string]error{"ConfigMap": cluster.ErrScopeUnknown}}

	got, err := DefaultNamespace(context.Background(), resolver, []source.Resource{resource("cfg", cm)}, "team-a")
	if err != nil {
		t.Fatalf("DefaultNamespace() error = %v", err)
	}
	if got[0].Namespace != "" {
		t.Errorf("Namespace = %q, want it left empty", got[0].Namespace)
	}
}

func TestDefaultNamespaceReturnsOtherErrors(t *testing.T) {
	cm := testutil.NewTestObj("v1", "ConfigMap", "cfg", "", nil)
	boom := errors.New("discovery unavailable")
	resolver := &fakeResolver{errs: map[string]error{"ConfigMap": boom}}

	_, err := DefaultNamespace(context.Background(), resolver, []source.Resource{resource("cfg", cm)}, "team-a")
	if !errors.Is(err, boom) {
		t.Fatalf("DefaultNamespace() error = %v, want it to wrap %v", err, boom)
	}
}

func TestDefaultNamespaceEmptyNamespaceIsNoop(t *testing.T) {
	cm := testutil.NewTestObj("v1", "ConfigMap", "cfg", "", nil)
	resolver := &fakeResolver{errs: map[string]error{"ConfigMap": errors.New("must not be called")}}

	got, err := DefaultNamespace(context.Background(), resolver, []source.Resource{resource("cfg", cm)}, "")
	if err != nil {
		t.Fatalf("DefaultNamespace() error = %v", err)
	}
	if len(got) != 1 || got[0].Namespace != "" {
		t.Errorf("DefaultNamespace() = %+v, want the input unchanged", got)
	}
}
