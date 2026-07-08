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
)

// Compare compares each already-loaded local resource against the live cluster
// state using the given fetcher, returning one Result per resource. A
// resource absent from the cluster is reported with StatusNew. It performs no
// filtering and no output rendering — callers decide what to do with the
// structured results.
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
			// Resource not found in cluster → treat as new.
			result, compareErr := diff.Compare(r.Object, nil, opts)
			if compareErr != nil {
				return nil, compareErr
			}
			results = append(results, result)
			continue
		}

		result, compareErr := diff.Compare(r.Object, clusterObj, opts)
		if compareErr != nil {
			return nil, compareErr
		}
		results = append(results, result)
	}
	return results, nil
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
