package cluster

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestNamespacedReportsScope(t *testing.T) {
	f, _, _ := newDiscoveryFetcher(t, coreResources(), clusterPolicyResources())

	cases := []struct {
		apiVersion, kind string
		want             bool
	}{
		{"v1", "Endpoints", true},
		{"v1", "Namespace", false},
		{"kyverno.io/v1", "ClusterPolicy", false},
	}
	for _, c := range cases {
		got, err := f.Namespaced(context.Background(), c.apiVersion, c.kind)
		if err != nil {
			t.Fatalf("Namespaced(%s %s) error = %v", c.apiVersion, c.kind, err)
		}
		if got != c.want {
			t.Errorf("Namespaced(%s %s) = %v, want %v", c.apiVersion, c.kind, got, c.want)
		}
	}
}

func TestNamespacedUnservedKindIsNoMatch(t *testing.T) {
	f, _, _ := newDiscoveryFetcher(t, coreResources())

	_, err := f.Namespaced(context.Background(), "example.com/v1", "Widget")
	if !meta.IsNoMatchError(err) {
		t.Fatalf("Namespaced() error = %v, want a no-match error", err)
	}
}

func TestNamespacedWithoutDiscovery(t *testing.T) {
	f := NewFetcherFromClient(dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()))

	_, err := f.Namespaced(context.Background(), "v1", "ConfigMap")
	if !errors.Is(err, ErrScopeUnknown) {
		t.Fatalf("Namespaced() error = %v, want ErrScopeUnknown", err)
	}
}

const testKubeconfig = `apiVersion: v1
kind: Config
clusters:
- name: c
  cluster:
    server: https://127.0.0.1:6443
users:
- name: u
  user:
    token: t
contexts:
- name: with-ns
  context:
    cluster: c
    user: u
    namespace: team-a
- name: without-ns
  context:
    cluster: c
    user: u
current-context: with-ns
`

func writeKubeconfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(path, []byte(testKubeconfig), 0o600); err != nil {
		t.Fatalf("write kubeconfig: %v", err)
	}
	return path
}

func TestContextNamespace(t *testing.T) {
	path := writeKubeconfig(t)

	cases := []struct {
		context, want string
	}{
		{"", "team-a"},
		{"with-ns", "team-a"},
		{"without-ns", "default"},
	}
	for _, c := range cases {
		got, err := ContextNamespace(path, c.context)
		if err != nil {
			t.Fatalf("ContextNamespace(%q) error = %v", c.context, err)
		}
		if got != c.want {
			t.Errorf("ContextNamespace(%q) = %q, want %q", c.context, got, c.want)
		}
	}
}

func TestContextNamespaceMissingKubeconfig(t *testing.T) {
	_, err := ContextNamespace(filepath.Join(t.TempDir(), "missing"), "")
	if err == nil {
		t.Fatal("ContextNamespace() with a missing kubeconfig returned no error")
	}
}
