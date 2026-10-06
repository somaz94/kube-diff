package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/somaz94/kube-diff/internal/testutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/rest"
	clienttesting "k8s.io/client-go/testing"
)

// The Discoverer doc promises these client-go types fit without an adapter.
var (
	_ Discoverer = (*discovery.DiscoveryClient)(nil)
	_ Discoverer = discovery.DiscoveryInterfaceWithContext(nil)
	_ Discoverer = (*fakediscovery.FakeDiscovery)(nil)
)

var (
	endpointsGVR     = schema.GroupVersionResource{Version: "v1", Resource: "endpoints"}
	namespacesGVR    = schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}
	prometheusGVR    = schema.GroupVersionResource{Group: "monitoring.coreos.com", Version: "v1", Resource: "prometheuses"}
	clusterPolicyGVR = schema.GroupVersionResource{Group: "kyverno.io", Version: "v1", Resource: "clusterpolicies"}
)

// Every kind here pluralizes irregularly or is cluster-scoped, so the old
// lowercase-plus-"s" guess gets each one wrong.
func coreResources() *metav1.APIResourceList {
	return &metav1.APIResourceList{GroupVersion: "v1", APIResources: []metav1.APIResource{
		{Name: "endpoints", Kind: "Endpoints", Namespaced: true},
		{Name: "namespaces", Kind: "Namespace", Namespaced: false},
	}}
}

// The status subresource is listed first on purpose: it shares the kind, and
// resolving to it would request the wrong path.
func prometheusResources() *metav1.APIResourceList {
	return &metav1.APIResourceList{GroupVersion: "monitoring.coreos.com/v1", APIResources: []metav1.APIResource{
		{Name: "prometheuses/status", Kind: "Prometheus", Namespaced: true},
		{Name: "prometheuses", Kind: "Prometheus", Namespaced: true},
	}}
}

func clusterPolicyResources() *metav1.APIResourceList {
	return &metav1.APIResourceList{GroupVersion: "kyverno.io/v1", APIResources: []metav1.APIResource{
		{Name: "clusterpolicies", Kind: "ClusterPolicy", Namespaced: false},
	}}
}

// newDiscoveryFetcher returns a Fetcher over a fake discovery serving lists and
// an empty fake dynamic client, plus both fakes for the test to drive.
func newDiscoveryFetcher(t *testing.T, lists ...*metav1.APIResourceList) (*Fetcher, *fakediscovery.FakeDiscovery, *dynamicfake.FakeDynamicClient) {
	t.Helper()

	disc := &fakediscovery.FakeDiscovery{Fake: &clienttesting.Fake{Resources: lists}}
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	return NewFetcherWithDiscovery(client, disc), disc, client
}

// mustCreate stores obj under gvr. The fake's constructor would file it under a
// guessed resource name instead, which is the very bug under test.
func mustCreate(t *testing.T, client *dynamicfake.FakeDynamicClient, gvr schema.GroupVersionResource, obj *unstructured.Unstructured) {
	t.Helper()

	var err error
	if ns := obj.GetNamespace(); ns != "" {
		_, err = client.Resource(gvr).Namespace(ns).Create(context.Background(), obj, metav1.CreateOptions{})
	} else {
		_, err = client.Resource(gvr).Create(context.Background(), obj, metav1.CreateOptions{})
	}
	if err != nil {
		t.Fatalf("create %s %s: %v", gvr.Resource, obj.GetName(), err)
	}
}

// discoveryCalls counts the discovery documents the fake has served.
func discoveryCalls(disc *fakediscovery.FakeDiscovery) int {
	n := 0
	for _, a := range disc.Actions() {
		if a.GetResource().Resource == "resource" {
			n++
		}
	}
	return n
}

func TestGetResolvesKindsThroughDiscovery(t *testing.T) {
	f, disc, client := newDiscoveryFetcher(t, coreResources(), prometheusResources(), clusterPolicyResources())
	mustCreate(t, client, endpointsGVR, testutil.NewTestObj("v1", "Endpoints", "kubernetes", "default", nil))
	mustCreate(t, client, namespacesGVR, testutil.NewTestObj("v1", "Namespace", "team-a", "", nil))
	mustCreate(t, client, prometheusGVR, testutil.NewTestObj("monitoring.coreos.com/v1", "Prometheus", "k8s", "monitoring", nil))
	mustCreate(t, client, clusterPolicyGVR, testutil.NewTestObj("kyverno.io/v1", "ClusterPolicy", "require-labels", "", nil))

	tests := []struct {
		name       string
		apiVersion string
		kind       string
		namespace  string
		objName    string
	}{
		{"irregular core plural", "v1", "Endpoints", "default", "kubernetes"},
		{"CRD with a custom plural", "monitoring.coreos.com/v1", "Prometheus", "monitoring", "k8s"},
		{"cluster-scoped core kind ignores namespace", "v1", "Namespace", "default", "team-a"},
		{"cluster-scoped CRD ignores namespace", "kyverno.io/v1", "ClusterPolicy", "default", "require-labels"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.Get(context.Background(), tt.apiVersion, tt.kind, tt.namespace, tt.objName)
			if err != nil {
				t.Fatalf("Get() error = %v, want nil", err)
			}
			if got.GetName() != tt.objName {
				t.Errorf("Get() name = %q, want %q", got.GetName(), tt.objName)
			}
		})
	}

	if got, want := discoveryCalls(disc), 3; got != want {
		t.Errorf("discovery documents fetched = %d, want %d (one per group version)", got, want)
	}
}

func TestGetUnservedKindIsNoMatch(t *testing.T) {
	f, _, _ := newDiscoveryFetcher(t, coreResources(), prometheusResources())

	tests := []struct {
		name       string
		apiVersion string
		kind       string
	}{
		{"group not served", "example.com/v1", "Widget"},
		{"kind not in a served group", "monitoring.coreos.com/v1", "Alertmanager"},
		{"version not served", "monitoring.coreos.com/v2", "Prometheus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.Get(context.Background(), tt.apiVersion, tt.kind, "default", "x")
			if !meta.IsNoMatchError(err) {
				t.Fatalf("Get() error = %v, want a no-match error", err)
			}
		})
	}
}

// Requesting a namespaced kind without a namespace is sent cluster-wide, as it
// was before discovery; the API server has no such route and answers NotFound.
func TestGetNamespacedKindWithoutNamespace(t *testing.T) {
	f, _, client := newDiscoveryFetcher(t, coreResources())
	mustCreate(t, client, endpointsGVR, testutil.NewTestObj("v1", "Endpoints", "kubernetes", "default", nil))

	_, err := f.Get(context.Background(), "v1", "Endpoints", "", "kubernetes")
	if !apierrors.IsNotFound(err) {
		t.Fatalf("Get() error = %v, want NotFound", err)
	}
}

func TestGetFindsKindAddedAfterDiscovery(t *testing.T) {
	f, disc, client := newDiscoveryFetcher(t, coreResources())
	clock := time.Unix(1_700_000_000, 0)
	f.now = func() time.Time { return clock }
	mustCreate(t, client, prometheusGVR, testutil.NewTestObj("monitoring.coreos.com/v1", "Prometheus", "k8s", "monitoring", nil))

	get := func() error {
		_, err := f.Get(context.Background(), "monitoring.coreos.com/v1", "Prometheus", "monitoring", "k8s")
		return err
	}

	if err := get(); !meta.IsNoMatchError(err) {
		t.Fatalf("Get() before the CRD is installed: error = %v, want a no-match error", err)
	}

	disc.Resources = append(disc.Resources, prometheusResources())

	clock = clock.Add(rediscoveryInterval - time.Second)
	if err := get(); !meta.IsNoMatchError(err) {
		t.Fatalf("Get() inside the rediscovery interval: error = %v, want the cached no-match", err)
	}
	if got := discoveryCalls(disc); got != 1 {
		t.Fatalf("discovery documents fetched inside the interval = %d, want 1", got)
	}

	clock = clock.Add(time.Second)
	if err := get(); err != nil {
		t.Fatalf("Get() after the interval: error = %v, want the newly installed kind", err)
	}
	if got := discoveryCalls(disc); got != 2 {
		t.Errorf("discovery documents fetched after the interval = %d, want 2", got)
	}
}

func TestGetDiscoveryErrorIsReturned(t *testing.T) {
	f, disc, _ := newDiscoveryFetcher(t, coreResources())
	errBoom := errors.New("discovery unavailable")
	disc.PrependReactor("get", "resource", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, errBoom
	})

	_, err := f.Get(context.Background(), "v1", "Endpoints", "default", "kubernetes")
	if !errors.Is(err, errBoom) {
		t.Fatalf("Get() error = %v, want it to wrap %v", err, errBoom)
	}
	if meta.IsNoMatchError(err) {
		t.Errorf("Get() error = %v, must not read as a no-match", err)
	}
}

// A failed rediscovery must surface too, not leave the stale no-match standing.
func TestGetRediscoveryErrorIsReturned(t *testing.T) {
	f, disc, _ := newDiscoveryFetcher(t, coreResources())
	clock := time.Unix(1_700_000_000, 0)
	f.now = func() time.Time { return clock }

	if _, err := f.Get(context.Background(), "example.com/v1", "Widget", "default", "w"); !meta.IsNoMatchError(err) {
		t.Fatalf("first Get() error = %v, want a no-match error", err)
	}

	errBoom := errors.New("discovery unavailable")
	disc.PrependReactor("get", "resource", func(clienttesting.Action) (bool, runtime.Object, error) {
		return true, nil, errBoom
	})
	clock = clock.Add(rediscoveryInterval)

	if _, err := f.Get(context.Background(), "example.com/v1", "Widget", "default", "w"); !errors.Is(err, errBoom) {
		t.Fatalf("Get() error = %v, want it to wrap %v", err, errBoom)
	}
}

func TestGetConcurrentUse(t *testing.T) {
	f, _, client := newDiscoveryFetcher(t, coreResources())
	mustCreate(t, client, endpointsGVR, testutil.NewTestObj("v1", "Endpoints", "kubernetes", "default", nil))

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Go(func() {
			_, err := f.Get(context.Background(), "v1", "Endpoints", "default", "kubernetes")
			errs <- err
		})
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("Get() error = %v, want nil", err)
		}
	}
}

func TestNewFetcherWithDiscoveryNilFallsBackToGuessing(t *testing.T) {
	f := NewFetcherWithDiscovery(fakeConfigMapClient(t), nil)

	got, err := f.Get(context.Background(), "v1", "ConfigMap", "default", "from-constructor")
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if got.GetName() != "from-constructor" {
		t.Errorf("Get() name = %q, want %q", got.GetName(), "from-constructor")
	}
}

// apiServer serves the discovery document of each list, answers each path in
// failPaths with that status code instead, returns NotFound for any other
// discovery document, and hands every other request to h.
func apiServer(t *testing.T, lists []*metav1.APIResourceList, failPaths map[string]int, h http.HandlerFunc) *httptest.Server {
	t.Helper()

	docs := map[string]*metav1.APIResourceList{}
	for _, list := range lists {
		if strings.Contains(list.GroupVersion, "/") {
			docs["/apis/"+list.GroupVersion] = list
		} else {
			docs["/api/"+list.GroupVersion] = list
		}
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code, failed := failPaths[r.URL.Path]
		doc, served := docs[r.URL.Path]
		switch {
		case failed:
		case served:
			writeJSON(w, http.StatusOK, doc)
			return
		case isDiscoveryPath(r.URL.Path):
			code = http.StatusNotFound
		default:
			h(w, r)
			return
		}
		writeJSON(w, code, metav1.Status{
			TypeMeta: metav1.TypeMeta{Kind: "Status", APIVersion: "v1"},
			Status:   metav1.StatusFailure,
			Message:  http.StatusText(code),
			Code:     int32(code),
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// isDiscoveryPath matches /api/<version> and /apis/<group>/<version>.
func isDiscoveryPath(path string) bool {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	return (len(parts) == 2 && parts[0] == "api") || (len(parts) == 3 && parts[0] == "apis")
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// NewFetcherFromConfig must wire discovery end to end: the request path proves
// the resource name and scope came from the API server, not from a guess.
func TestNewFetcherFromConfigResolvesThroughDiscovery(t *testing.T) {
	tests := []struct {
		name       string
		apiVersion string
		kind       string
		namespace  string
		wantPath   string
	}{
		{"irregular core plural", "v1", "Endpoints", "default", "/api/v1/namespaces/default/endpoints/obj"},
		{"CRD with a custom plural", "monitoring.coreos.com/v1", "Prometheus", "monitoring", "/apis/monitoring.coreos.com/v1/namespaces/monitoring/prometheuses/obj"},
		{"cluster-scoped kind drops the namespace", "kyverno.io/v1", "ClusterPolicy", "default", "/apis/kyverno.io/v1/clusterpolicies/obj"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := apiServer(t, []*metav1.APIResourceList{coreResources(), prometheusResources(), clusterPolicyResources()}, nil,
				func(w http.ResponseWriter, r *http.Request) {
					gotPath = r.URL.Path
					writeJSON(w, http.StatusOK, testutil.NewTestObj(tt.apiVersion, tt.kind, "obj", "", nil).Object)
				})

			f, err := NewFetcherFromConfig(&rest.Config{Host: srv.URL})
			if err != nil {
				t.Fatalf("NewFetcherFromConfig() error = %v", err)
			}
			if _, err := f.Get(context.Background(), tt.apiVersion, tt.kind, tt.namespace, "obj"); err != nil {
				t.Fatalf("Get() error = %v, want nil", err)
			}
			if gotPath != tt.wantPath {
				t.Errorf("request path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

// A discovery document that cannot be read (an unavailable aggregated API) must
// not make its kinds look absent, nor break healthy groups.
func TestNewFetcherFromConfigDiscoveryErrors(t *testing.T) {
	metricsResources := &metav1.APIResourceList{GroupVersion: "metrics.example.com/v1", APIResources: []metav1.APIResource{
		{Name: "nodemetrics", Kind: "NodeMetrics", Namespaced: false},
	}}
	srv := apiServer(t,
		[]*metav1.APIResourceList{coreResources(), metricsResources},
		map[string]int{"/apis/metrics.example.com/v1": http.StatusServiceUnavailable},
		func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, testutil.NewTestObj("v1", "Endpoints", "kubernetes", "default", nil).Object)
		})

	f, err := NewFetcherFromConfig(&rest.Config{Host: srv.URL})
	if err != nil {
		t.Fatalf("NewFetcherFromConfig() error = %v", err)
	}

	_, err = f.Get(context.Background(), "metrics.example.com/v1", "NodeMetrics", "", "node-1")
	if !apierrors.IsServiceUnavailable(err) || meta.IsNoMatchError(err) {
		t.Errorf("Get() on an unavailable group version: error = %v, want ServiceUnavailable, not a no-match", err)
	}

	if _, err := f.Get(context.Background(), "unserved.example.com/v1", "Relic", "default", "r"); !meta.IsNoMatchError(err) {
		t.Errorf("Get() on an unserved group version: error = %v, want a no-match error", err)
	}

	if _, err := f.Get(context.Background(), "v1", "Endpoints", "default", "kubernetes"); err != nil {
		t.Errorf("Get() on a healthy group: error = %v, want nil", err)
	}
}

func TestNewFetcherFromConfigUndecodableDiscovery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	t.Cleanup(srv.Close)

	f, err := NewFetcherFromConfig(&rest.Config{Host: srv.URL})
	if err != nil {
		t.Fatalf("NewFetcherFromConfig() error = %v", err)
	}

	_, err = f.Get(context.Background(), "example.com/v1", "Widget", "default", "w")
	if err == nil || meta.IsNoMatchError(err) || !strings.Contains(err.Error(), "decode discovery document") {
		t.Errorf("Get() error = %v, want a decode error", err)
	}
}
