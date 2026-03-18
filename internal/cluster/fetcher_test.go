package cluster

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestGuessResourceName(t *testing.T) {
	tests := []struct {
		kind     string
		expected string
	}{
		{"Deployment", "deployments"},
		{"Service", "services"},
		{"ConfigMap", "configmaps"},
		{"Secret", "secrets"},
		{"Pod", "pods"},
		{"Namespace", "namespaces"},
		{"Node", "nodes"},
		{"Ingress", "ingresses"},
		{"NetworkPolicy", "networkpolicies"},
		{"EndpointSlice", "endpointslices"},
		{"StorageClass", "storageclasses"},
		{"IngressClass", "ingressclasses"},
		{"ResourceQuota", "resourcequotas"},
		{"PriorityClass", "priorityclasses"},
		{"RuntimeClass", "runtimeclasses"},
		{"ServiceAccount", "serviceaccounts"},
		{"DaemonSet", "daemonsets"},
		{"StatefulSet", "statefulsets"},
		{"ReplicaSet", "replicasets"},
		{"Job", "jobs"},
		{"CronJob", "cronjobs"},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			result := guessResourceName(tt.kind)
			if result != tt.expected {
				t.Errorf("guessResourceName(%q) = %q, want %q", tt.kind, result, tt.expected)
			}
		})
	}
}

func TestResolveGVR(t *testing.T) {
	f := &Fetcher{mapper: &restMapper{}}

	tests := []struct {
		name       string
		apiVersion string
		kind       string
		wantGroup  string
		wantVer    string
		wantRes    string
	}{
		{
			name:       "core v1 ConfigMap",
			apiVersion: "v1",
			kind:       "ConfigMap",
			wantGroup:  "",
			wantVer:    "v1",
			wantRes:    "configmaps",
		},
		{
			name:       "apps/v1 Deployment",
			apiVersion: "apps/v1",
			kind:       "Deployment",
			wantGroup:  "apps",
			wantVer:    "v1",
			wantRes:    "deployments",
		},
		{
			name:       "networking.k8s.io/v1 NetworkPolicy",
			apiVersion: "networking.k8s.io/v1",
			kind:       "NetworkPolicy",
			wantGroup:  "networking.k8s.io",
			wantVer:    "v1",
			wantRes:    "networkpolicies",
		},
		{
			name:       "batch/v1 CronJob",
			apiVersion: "batch/v1",
			kind:       "CronJob",
			wantGroup:  "batch",
			wantVer:    "v1",
			wantRes:    "cronjobs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gvr, err := f.resolveGVR(tt.apiVersion, tt.kind)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gvr.Group != tt.wantGroup {
				t.Errorf("Group: got %q, want %q", gvr.Group, tt.wantGroup)
			}
			if gvr.Version != tt.wantVer {
				t.Errorf("Version: got %q, want %q", gvr.Version, tt.wantVer)
			}
			if gvr.Resource != tt.wantRes {
				t.Errorf("Resource: got %q, want %q", gvr.Resource, tt.wantRes)
			}
		})
	}
}

func TestResolveGVRInvalidApiVersion(t *testing.T) {
	f := &Fetcher{mapper: &restMapper{}}
	_, err := f.resolveGVR("invalid/version/extra", "Pod")
	if err == nil {
		t.Fatal("expected error for invalid apiVersion")
	}
}

func TestNewFetcherWithFakeKubeconfig(t *testing.T) {
	dir := t.TempDir()
	kubeconfig := filepath.Join(dir, "config")
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: fake-token
`
	if err := os.WriteFile(kubeconfig, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	f, err := NewFetcher(kubeconfig, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewFetcherWithContext(t *testing.T) {
	dir := t.TempDir()
	kubeconfig := filepath.Join(dir, "config")
	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: cluster-a
- cluster:
    server: https://127.0.0.1:6444
  name: cluster-b
contexts:
- context:
    cluster: cluster-a
    user: user-a
  name: context-a
- context:
    cluster: cluster-b
    user: user-b
  name: context-b
current-context: context-a
users:
- name: user-a
  user:
    token: token-a
- name: user-b
  user:
    token: token-b
`
	if err := os.WriteFile(kubeconfig, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	f, err := NewFetcher(kubeconfig, "context-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewFetcherInvalidKubeconfig(t *testing.T) {
	dir := t.TempDir()
	kubeconfig := filepath.Join(dir, "bad-config")
	os.WriteFile(kubeconfig, []byte(`not valid yaml{{{`), 0600)

	_, err := NewFetcher(kubeconfig, "")
	if err == nil {
		t.Fatal("expected error for invalid kubeconfig")
	}
}

func TestNewFetcherNonExistentKubeconfig(t *testing.T) {
	_, err := NewFetcher("/nonexistent/kubeconfig", "")
	if err == nil {
		t.Fatal("expected error for nonexistent kubeconfig")
	}
}

func TestGetNamespacedResource(t *testing.T) {
	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-cm",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key": "value",
			},
		},
	}

	fakeClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			gvr: "ConfigMapList",
		}, obj)

	f := &Fetcher{
		client: fakeClient,
		mapper: &restMapper{},
	}

	result, err := f.Get(context.Background(), "v1", "ConfigMap", "default", "test-cm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GetName() != "test-cm" {
		t.Errorf("expected name test-cm, got %s", result.GetName())
	}
	if result.GetNamespace() != "default" {
		t.Errorf("expected namespace default, got %s", result.GetNamespace())
	}
}

func TestGetClusterScopedResource(t *testing.T) {
	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": "test-ns",
			},
		},
	}

	fakeClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			gvr: "NamespaceList",
		}, obj)

	f := &Fetcher{
		client: fakeClient,
		mapper: &restMapper{},
	}

	result, err := f.Get(context.Background(), "v1", "Namespace", "", "test-ns")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GetName() != "test-ns" {
		t.Errorf("expected name test-ns, got %s", result.GetName())
	}
}

func TestGetResourceNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}

	fakeClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			gvr: "ConfigMapList",
		})

	f := &Fetcher{
		client: fakeClient,
		mapper: &restMapper{},
	}

	_, err := f.Get(context.Background(), "v1", "ConfigMap", "default", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent resource")
	}
}

func TestGetWithInvalidApiVersion(t *testing.T) {
	scheme := runtime.NewScheme()
	fakeClient := dynamicfake.NewSimpleDynamicClient(scheme)

	f := &Fetcher{
		client: fakeClient,
		mapper: &restMapper{},
	}

	_, err := f.Get(context.Background(), "invalid/v/extra", "Pod", "default", "test")
	if err == nil {
		t.Fatal("expected error for invalid apiVersion")
	}
}

func TestGetMultipleResources(t *testing.T) {
	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}

	obj1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "cm-1",
				"namespace": "default",
			},
		},
	}
	obj2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "cm-2",
				"namespace": "default",
			},
		},
	}

	fakeClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme,
		map[schema.GroupVersionResource]string{
			gvr: "ConfigMapList",
		}, obj1, obj2)

	f := &Fetcher{
		client: fakeClient,
		mapper: &restMapper{},
	}

	r1, err := f.Get(context.Background(), "v1", "ConfigMap", "default", "cm-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r1.GetName() != "cm-1" {
		t.Errorf("expected cm-1, got %s", r1.GetName())
	}

	r2, err := f.Get(context.Background(), "v1", "ConfigMap", "default", "cm-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r2.GetName() != "cm-2" {
		t.Errorf("expected cm-2, got %s", r2.GetName())
	}
}
