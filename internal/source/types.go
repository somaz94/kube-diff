package source

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// Resource represents a parsed Kubernetes resource from a local source.
type Resource struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
	Object     *unstructured.Unstructured
}

// Source loads Kubernetes resources from a local source (file, helm, kustomize).
type Source interface {
	Load() ([]Resource, error)
}
