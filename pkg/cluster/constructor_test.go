package cluster

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/rest"
)

// fakeConfigMapClient returns a dynamic fake preloaded with one ConfigMap.
func fakeConfigMapClient(t *testing.T) *dynamicfake.FakeDynamicClient {
	t.Helper()

	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "from-constructor",
				"namespace": "default",
			},
		},
	}

	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{gvr: "ConfigMapList"}, obj)
}

// The exported constructor is the entry point library consumers use, so it is
// exercised here rather than only the unexported struct literal.
func TestNewFetcherFromClient(t *testing.T) {
	f := NewFetcherFromClient(fakeConfigMapClient(t))
	if f == nil {
		t.Fatal("NewFetcherFromClient() = nil, want a Fetcher")
	}

	got, err := f.Get(context.Background(), "v1", "ConfigMap", "default", "from-constructor")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.GetName() != "from-constructor" {
		t.Errorf("Get() name = %q, want %q", got.GetName(), "from-constructor")
	}
}

func TestNewFetcherFromConfig(t *testing.T) {
	f, err := NewFetcherFromConfig(&rest.Config{Host: "https://127.0.0.1:6443"})
	if err != nil {
		t.Fatalf("NewFetcherFromConfig() error = %v, want nil", err)
	}
	if f == nil {
		t.Fatal("NewFetcherFromConfig() = nil, want a Fetcher")
	}
}

func TestNewFetcherFromConfig_NilConfig(t *testing.T) {
	f, err := NewFetcherFromConfig(nil)
	if err == nil {
		t.Fatal("NewFetcherFromConfig(nil) error = nil, want a rejection")
	}
	if f != nil {
		t.Errorf("NewFetcherFromConfig(nil) = %v, want nil alongside the error", f)
	}
}
