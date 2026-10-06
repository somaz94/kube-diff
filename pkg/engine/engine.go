// Package engine orchestrates the load → fetch → compare pipeline that powers
// kube-diff. It exposes the comparison core as importable functions so that
// out-of-tree consumers (for example an in-cluster drift-detection controller)
// can reuse the exact same diff logic without depending on the CLI layer,
// stdout rendering, or os.Exit behavior.
package engine

import (
	"context"
	"fmt"

	"github.com/somaz94/kube-diff/pkg/cluster"
	"github.com/somaz94/kube-diff/pkg/diff"
	"github.com/somaz94/kube-diff/pkg/source"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
)

// Compare compares each already-loaded local resource against the live cluster
// state using the given fetcher, returning one Result per resource. A resource
// the cluster reports as absent (a NotFound error, or a no-match error for a
// kind the cluster does not serve) is reported with StatusNew. Any other fetch
// error, such as Forbidden or a timeout, aborts the comparison and is returned
// wrapped with the resource identity, so a missing read permission or an
// unreachable API server is never mistaken for drift. It performs no filtering
// and no output rendering — callers decide what to do with the structured
// results.
func Compare(
	ctx context.Context,
	fetcher cluster.ResourceFetcher,
	resources []source.Resource,
	opts diff.CompareOptions,
) ([]*diff.Result, error) {
	results := make([]*diff.Result, 0, len(resources))
	for _, r := range resources {
		clusterObj, err := fetcher.Get(ctx, r.APIVersion, r.Kind, r.Namespace, r.Name)
		if err != nil {
			if !isAbsent(err) {
				id := r.Name
				if r.Namespace != "" {
					id = r.Namespace + "/" + r.Name
				}
				return nil, fmt.Errorf("get %s %s %s: %w", r.APIVersion, r.Kind, id, err)
			}
			// Drop anything returned alongside the error; nil makes diff.Compare report StatusNew.
			clusterObj = nil
		}

		result, compareErr := diff.Compare(r.Object, clusterObj, opts)
		if compareErr != nil {
			return nil, compareErr
		}
		results = append(results, result)
	}
	return results, nil
}

// isAbsent reports whether a fetch error means the object does not exist. An
// unserved kind (CRD not installed) counts: discovery reports it as a no-match
// error, and a request that still reaches the server for it gets a 404.
func isAbsent(err error) bool {
	return apierrors.IsNotFound(err) || meta.IsNoMatchError(err)
}

// Run is a convenience wrapper that loads resources from src and then compares
// them against the cluster via Compare. It is the simplest entry point for
// consumers that do not need the CLI's intermediate filtering step.
func Run(
	ctx context.Context,
	src source.Source,
	fetcher cluster.ResourceFetcher,
	opts diff.CompareOptions,
) ([]*diff.Result, error) {
	resources, err := src.Load()
	if err != nil {
		return nil, fmt.Errorf("load resources: %w", err)
	}
	return Compare(ctx, fetcher, resources, opts)
}
