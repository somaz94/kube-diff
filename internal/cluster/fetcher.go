package cluster

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// ResourceFetcher is the interface for retrieving resources from a Kubernetes cluster.
type ResourceFetcher interface {
	Get(ctx context.Context, apiVersion, kind, namespace, name string) (*unstructured.Unstructured, error)
}

// Fetcher retrieves resources from a Kubernetes cluster.
type Fetcher struct {
	client dynamic.Interface
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

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Fetcher{
		client: client,
	}, nil
}

// Get retrieves a single resource from the cluster.
func (f *Fetcher) Get(ctx context.Context, apiVersion, kind, namespace, name string) (*unstructured.Unstructured, error) {
	gvr, err := f.resolveGVR(apiVersion, kind)
	if err != nil {
		return nil, err
	}

	var resource *unstructured.Unstructured
	if namespace != "" {
		resource, err = f.client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		resource, err = f.client.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (f *Fetcher) resolveGVR(apiVersion, kind string) (schema.GroupVersionResource, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("invalid apiVersion %q: %w", apiVersion, err)
	}

	// Simple mapping for common resources
	resourceName := guessResourceName(kind)

	return schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: resourceName,
	}, nil
}

// guessResourceName converts Kind to plural resource name.
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
