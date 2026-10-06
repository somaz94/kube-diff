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

// ResourceFetcher is the interface for retrieving resources from a Kubernetes cluster.
type ResourceFetcher interface {
	Get(ctx context.Context, apiVersion, kind, namespace, name string) (*unstructured.Unstructured, error)
}

// Discoverer returns the resources the API server serves for one group version,
// answering NotFound for a group version it does not serve. client-go's
// *discovery.DiscoveryClient and its fake satisfy it, as does any
// discovery.DiscoveryInterface wrapped with discovery.ToDiscoveryInterfaceWithContext.
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

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	restClient, err := rest.UnversionedRESTClientFor(dynamic.ConfigFor(config))
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	return NewFetcherWithDiscovery(client, restDiscovery{client: restClient}), nil
}

// NewFetcherWithDiscovery creates a Fetcher that resolves each kind to its
// resource name and scope by reading the discovery document of the kind's
// group version, once per group version. A kind the cached document lacks
// triggers a refetch on a later Get (at most one per group version per
// rediscoveryInterval, 30 seconds), so a CRD installed after the Fetcher was
// built is found. A nil disc gives the same Fetcher as NewFetcherFromClient.
func NewFetcherWithDiscovery(client dynamic.Interface, disc Discoverer) *Fetcher {
	return &Fetcher{
		client:     client,
		discovery:  disc,
		now:        time.Now,
		discovered: make(map[schema.GroupVersion]discoveredGV),
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
// A kind the API server does not serve returns an error satisfying
// meta.IsNoMatchError. A cluster-scoped kind is fetched without the namespace.
// A namespaced kind with an empty namespace is requested without one, which
// the API server answers with NotFound; callers that want a default namespace
// must fill it in before calling Get.
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

	f.mu.Lock()
	defer f.mu.Unlock()

	entry, cached := f.discovered[gv]
	resource := findKind(entry.resources, kind)
	if !cached || (resource == nil && f.now().Sub(entry.fetchedAt) >= rediscoveryInterval) {
		if entry, err = f.discoverLocked(ctx, gv); err != nil {
			return schema.GroupVersionResource{}, false, err
		}
		resource = findKind(entry.resources, kind)
	}

	if resource == nil {
		noMatch := &meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: gv.Group, Kind: kind}, SearchedVersions: []string{gv.Version}}
		return schema.GroupVersionResource{}, false, fmt.Errorf("resolve resource for kind %s: %w", kind, noMatch)
	}
	return gv.WithResource(resource.Name), resource.Namespaced, nil
}

// discoverLocked fetches and caches gv's discovery document. The caller must
// hold f.mu. NotFound means gv is not served and is cached like any answer;
// any other error is returned uncached, so it is never mistaken for absence.
func (f *Fetcher) discoverLocked(ctx context.Context, gv schema.GroupVersion) (discoveredGV, error) {
	list, err := f.discovery.ServerResourcesForGroupVersionWithContext(ctx, gv.String())
	if err != nil && !apierrors.IsNotFound(err) {
		return discoveredGV{}, fmt.Errorf("discover resources for %s: %w", gv, err)
	}

	entry := discoveredGV{fetchedAt: f.now()}
	if err == nil && list != nil {
		entry.resources = list.APIResources
	}
	f.discovered[gv] = entry
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

	body, err := d.client.Get().AbsPath(path).SetHeader("Accept", "application/json").Do(ctx).Raw()
	if err != nil {
		return nil, err
	}

	list := &metav1.APIResourceList{}
	if err := json.Unmarshal(body, list); err != nil {
		return nil, fmt.Errorf("decode discovery document %s: %w", path, err)
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
