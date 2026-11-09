package cluster

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
)

func labeled(kind, namespace, name string, labels map[string]interface{}) *unstructured.Unstructured {
	meta := map[string]interface{}{"name": name, "labels": labels}
	if namespace != "" {
		meta["namespace"] = namespace
	}
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1", "kind": kind, "metadata": meta,
	}}
}

func listFetcher(objs ...runtime.Object) *Fetcher {
	client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Version: "v1", Resource: "configmaps"}: "ConfigMapList",
			{Version: "v1", Resource: "namespaces"}: "NamespaceList",
		}, objs...)
	return &Fetcher{client: client}
}

func names(items []unstructured.Unstructured) string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.GetNamespace()+"/"+it.GetName())
	}
	return strings.Join(out, ",")
}

func TestListNamespacedFiltersBySelectorAndNamespace(t *testing.T) {
	app := map[string]interface{}{"app": "web"}
	f := listFetcher(
		labeled("ConfigMap", "default", "a", app),
		labeled("ConfigMap", "default", "other", map[string]interface{}{"app": "db"}),
		labeled("ConfigMap", "prod", "b", app),
	)

	items, err := f.List(context.Background(), "v1", "ConfigMap", "default", "app=web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := names(items); got != "default/a" {
		t.Errorf("expected only default/a, got %q", got)
	}
}

func TestListClusterScoped(t *testing.T) {
	f := listFetcher(
		labeled("Namespace", "", "team-a", map[string]interface{}{"team": "a"}),
		labeled("Namespace", "", "team-b", map[string]interface{}{"team": "b"}),
	)

	items, err := f.List(context.Background(), "v1", "Namespace", "", "team=a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := names(items); got != "/team-a" {
		t.Errorf("expected only team-a, got %q", got)
	}
}

func TestListNamespacedKindNeedsNamespace(t *testing.T) {
	f, _, _ := newDiscoveryFetcher(t, coreResources())
	if _, err := f.List(context.Background(), "v1", "Endpoints", "", "app=web"); err == nil {
		t.Fatal("expected an error for a namespaced kind without a namespace")
	}
}

func TestListClusterScopedWithDiscovery(t *testing.T) {
	client := listFetcher(labeled("Namespace", "", "team-a", map[string]interface{}{"team": "a"})).client
	f := NewFetcherWithDiscovery(client, &fakediscovery.FakeDiscovery{
		Fake: &clienttesting.Fake{Resources: []*metav1.APIResourceList{coreResources()}},
	})

	items, err := f.List(context.Background(), "v1", "Namespace", "", "team=a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := names(items); got != "/team-a" {
		t.Errorf("expected team-a, got %q", got)
	}
}

func TestListUnresolvableKind(t *testing.T) {
	f := listFetcher()
	if _, err := f.List(context.Background(), "not a version/x/y", "Thing", "default", "a=b"); err == nil {
		t.Fatal("expected an error for an unparsable apiVersion")
	}
}
