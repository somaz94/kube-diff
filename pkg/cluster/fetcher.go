package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// rediscoveryInterval bounds how often a group version that lacked the
// requested kind is fetched again. Without it, a manifest set full of objects
// whose CRD is not installed would refetch once per object.
const rediscoveryInterval = 30 * time.Second

// maxDiscoveryAge bounds how long a cached discovery document is trusted, so a
// CRD recreated with a different plural or scope is picked up without a restart.
const maxDiscoveryAge = 10 * time.Minute

// ResourceFetcher is the interface for retrieving resources from a Kubernetes cluster.
type ResourceFetcher interface {
	Get(ctx context.Context, apiVersion, kind, namespace, name string) (*unstructured.Unstructured, error)
}

// Discoverer returns the resources the API server serves for one group version.
// For a group version the server does not serve it must return an error
// satisfying apierrors.IsNotFound, and every call must reach the server so a
// newly installed CRD is seen; the Fetcher does its own caching. client-go's
// *discovery.DiscoveryClient and its fake qualify. memory.NewMemCacheClient does
// not: it answers ErrCacheNotFound and never refetches until invalidated.
type Discoverer interface {
	ServerResourcesForGroupVersionWithContext(ctx context.Context, groupVersion string) (*metav1.APIResourceList, error)
}

// Fetcher retrieves resources from a Kubernetes cluster. It is safe for
// concurrent use and is meant to be long-lived: discovery results are cached.
type Fetcher struct {
	client dynamic.Interface
	// discovery is nil for a Fetcher built by NewFetcherFromClient, which
	// guesses resource names instead of asking the API server.
	discovery Discoverer
	now       func() time.Time

	mu         sync.Mutex
	discovered map[schema.GroupVersion]discoveredGV
	inflight   map[schema.GroupVersion]chan struct{}
}

// discoveredGV is one cached discovery document; resources is nil when the
// API server does not serve the group version.
type discoveredGV struct {
	resources []metav1.APIResource
	fetchedAt time.Time
}

// NewFetcher creates a new cluster Fetcher.
func NewFetcher(kubeconfig, kubecontext string) (*Fetcher, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		rules.ExplicitPath = kubeconfig
	}

	overrides := &clientcmd.ConfigOverrides{}
	if kubecontext != "" {
		overrides.CurrentContext = kubecontext
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	return NewFetcherFromConfig(config)
}

// NewFetcherFromConfig creates a new cluster Fetcher from an existing
// *rest.Config. This is the entry point for in-cluster consumers such as a
// controller-runtime manager, which already hold a REST config and should not
// go through kubeconfig file loading. Kinds are resolved through API
// discovery, as with NewFetcherWithDiscovery.
func NewFetcherFromConfig(config *rest.Config) (*Fetcher, error) {
	if config == nil {
		return nil, errors.New("rest config must not be nil")
	}

	cfg := dynamic.ConfigFor(config)
	httpClient, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	client, err := dynamic.NewForConfigAndClient(config, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	restClient, err := rest.UnversionedRESTClientForConfigAndClient(cfg, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	return NewFetcherWithDiscovery(client, restDiscovery{client: restClient}), nil
}

// NewFetcherWithDiscovery creates a Fetcher that resolves each kind to its
// resource name and scope by reading the discovery document of the kind's
// group version. A cached document is trusted for maxDiscoveryAge (10 minutes),
// and a kind it lacks triggers a refetch at most once per rediscoveryInterval
// (30 seconds), so a CRD installed, or recreated with a new plural or scope,
// after the Fetcher was built is found. If refreshing an aged document fails, a
// kind it still lists is served from it. A nil disc gives the same Fetcher as
// NewFetcherFromClient.
func NewFetcherWithDiscovery(client dynamic.Interface, disc Discoverer) *Fetcher {
	return &Fetcher{
		client:     client,
		discovery:  disc,
		now:        time.Now,
		discovered: make(map[schema.GroupVersion]discoveredGV),
		inflight:   make(map[schema.GroupVersion]chan struct{}),
	}
}

// NewFetcherFromClient creates a Fetcher from an existing dynamic client.
// Useful for tests and for consumers that build a dynamic client themselves.
// The client must be non-nil; a nil client defers failure to the first Get call.
//
// Without discovery the Fetcher guesses the resource name from the kind
// (lowercase plus "s", with a fixed list of exceptions) and sends the
// namespace exactly as given. A kind the guess gets wrong, such as Endpoints or
// a CRD with an irregular plural, is reported by the API server as NotFound.
// Use NewFetcherWithDiscovery to resolve kinds the way the API server serves them.
func NewFetcherFromClient(client dynamic.Interface) *Fetcher {
	return &Fetcher{client: client}
}

// Get retrieves a single resource from the cluster.
//
// With discovery, a kind the API server does not serve returns an error
// satisfying meta.IsNoMatchError, and a cluster-scoped kind is fetched without
// the namespace. A namespaced kind with an empty namespace is requested without
// one, which the API server answers with NotFound; callers that want a default
// namespace must fill it in before calling Get.
func (f *Fetcher) Get(ctx context.Context, apiVersion, kind, namespace, name string) (*unstructured.Unstructured, error) {
	gvr, namespaced, err := f.resolve(ctx, apiVersion, kind)
	if err != nil {
		return nil, err
	}

	var resource *unstructured.Unstructured
	if namespaced && namespace != "" {
		resource, err = f.client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		resource, err = f.client.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return nil, err
	}

	return resource, nil
}

// resolve returns the resource to request for apiVersion and kind, and whether
// the namespace belongs in the request.
func (f *Fetcher) resolve(ctx context.Context, apiVersion, kind string) (schema.GroupVersionResource, bool, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return schema.GroupVersionResource{}, false, fmt.Errorf("invalid apiVersion %q: %w", apiVersion, err)
	}

	if f.discovery == nil {
		return gv.WithResource(guessResourceName(kind)), true, nil
	}
	// An empty version would read /api/ or /apis/<group>/, which are not
	// resource lists and would decode as "nothing served".
	if gv.Version == "" {
		return schema.GroupVersionResource{}, false, fmt.Errorf("invalid apiVersion %q: missing version", apiVersion)
	}

	resource, err := f.lookup(ctx, gv, kind)
	if err != nil {
		return schema.GroupVersionResource{}, false, err
	}
	if resource == nil {
		noMatch := &meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: gv.Group, Kind: kind}, SearchedVersions: []string{gv.Version}}
		return schema.GroupVersionResource{}, false, fmt.Errorf("resolve resource for kind %s: %w", kind, noMatch)
	}
	return gv.WithResource(resource.Name), resource.Namespaced, nil
}

// lookup returns kind's top-level resource in gv, or nil when gv does not serve
// it. The discovery request runs without f.mu held, at most one per group
// version at a time; other callers for that group version wait on it or on
// their own ctx, and callers the cache can answer never wait.
func (f *Fetcher) lookup(ctx context.Context, gv schema.GroupVersion, kind string) (*metav1.APIResource, error) {
	for {
		f.mu.Lock()
		entry, cached := f.discovered[gv]
		res := findKind(entry.resources, kind)
		age := f.now().Sub(entry.fetchedAt)
		if cached && age < maxDiscoveryAge && (res != nil || age < rediscoveryInterval) {
			f.mu.Unlock()
			return res, nil
		}
		if wait, busy := f.inflight[gv]; busy {
			f.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-wait:
				continue
			}
		}
		done := make(chan struct{})
		f.inflight[gv] = done
		f.mu.Unlock()

		fresh, err := f.fetch(ctx, gv, done)
		if err != nil {
			if res != nil {
				// An aged answer beats failing every comparison while discovery is down.
				return res, nil
			}
			return nil, err
		}
		return findKind(fresh.resources, kind), nil
	}
}

// fetch reads gv's discovery document without f.mu held, then caches it and
// wakes waiters, even if the Discoverer panics. NotFound means gv is not served
// and is cached like any answer; any other error is returned uncached, so it is
// never mistaken for absence.
func (f *Fetcher) fetch(ctx context.Context, gv schema.GroupVersion, done chan struct{}) (entry discoveredGV, err error) {
	ok := false
	defer func() {
		f.mu.Lock()
		if ok {
			f.discovered[gv] = entry
		}
		delete(f.inflight, gv)
		f.mu.Unlock()
		close(done)
	}()

	list, err := f.discovery.ServerResourcesForGroupVersionWithContext(ctx, gv.String())
	if err != nil && !apierrors.IsNotFound(err) {
		return discoveredGV{}, fmt.Errorf("discover resources for %s: %w", gv, err)
	}
	entry = discoveredGV{fetchedAt: f.now()}
	if err == nil && list != nil {
		entry.resources = list.APIResources
	}
	ok = true
	return entry, nil
}

// findKind returns the top-level resource serving kind; subresources such as
// deployments/scale share their parent's kind and are skipped.
func findKind(resources []metav1.APIResource, kind string) *metav1.APIResource {
	for i := range resources {
		if resources[i].Kind == kind && !strings.Contains(resources[i].Name, "/") {
			return &resources[i]
		}
	}
	return nil
}

// restDiscovery reads one group version's discovery document over a plain REST
// client. It stands in for client-go's discovery package, which links every
// typed API group into the binary and more than doubles its size.
type restDiscovery struct {
	client rest.Interface
}

func (d restDiscovery) ServerResourcesForGroupVersionWithContext(ctx context.Context, groupVersion string) (*metav1.APIResourceList, error) {
	path := "/apis/" + groupVersion
	if !strings.Contains(groupVersion, "/") {
		path = "/api/" + groupVersion
	}

	// Error() before Raw() keeps the server's Status message (e.g. which APIService is down).
	result := d.client.Get().AbsPath(path).SetHeader("Accept", "application/json").Do(ctx)
	if err := result.Error(); err != nil {
		return nil, err
	}
	body, err := result.Raw()
	if err != nil {
		return nil, err
	}

	list := &metav1.APIResourceList{}
	if err := json.Unmarshal(body, list); err != nil {
		return nil, fmt.Errorf("decode discovery document %s: %w", path, err)
	}
	if list.GroupVersion != groupVersion {
		return nil, fmt.Errorf("discovery document %s is for %q, want %q", path, list.GroupVersion, groupVersion)
	}
	return list, nil
}

// guessResourceName converts Kind to plural resource name. It is the fallback
// for a Fetcher without discovery.
func guessResourceName(kind string) string {
	// Handle special pluralization
	specialPlurals := map[string]string{
		"Ingress":       "ingresses",
		"NetworkPolicy": "networkpolicies",
		"EndpointSlice": "endpointslices",
		"StorageClass":  "storageclasses",
		"IngressClass":  "ingressclasses",
		"ResourceQuota": "resourcequotas",
		"PriorityClass": "priorityclasses",
		"RuntimeClass":  "runtimeclasses",
	}

	if plural, ok := specialPlurals[kind]; ok {
		return plural
	}

	return strings.ToLower(kind) + "s"
}
