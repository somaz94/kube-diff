package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/somaz94/kube-diff/internal/cluster"
	"github.com/somaz94/kube-diff/internal/diff"
	"github.com/somaz94/kube-diff/internal/report"
	"github.com/somaz94/kube-diff/internal/source"
	"github.com/spf13/cobra"
)

// runDiff is the shared logic for file, helm, and kustomize commands.
func runDiff(cmd *cobra.Command, src source.Source) error {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	kubeContext, _ := cmd.Flags().GetString("context")
	namespace, _ := cmd.Flags().GetString("namespace")
	kinds, _ := cmd.Flags().GetStringSlice("kind")
	selector, _ := cmd.Flags().GetString("selector")
	summaryOnly, _ := cmd.Flags().GetBool("summary-only")
	output, _ := cmd.Flags().GetString("output")

	// Load local resources
	resources, err := src.Load()
	if err != nil {
		return fmt.Errorf("failed to load resources: %w", err)
	}

	// Filter by namespace
	if namespace != "" {
		var filtered []source.Resource
		for _, r := range resources {
			if r.Namespace == namespace || r.Namespace == "" {
				filtered = append(filtered, r)
			}
		}
		resources = filtered
	}

	// Filter by kind
	if len(kinds) > 0 {
		kindSet := make(map[string]bool)
		for _, k := range kinds {
			kindSet[k] = true
		}
		var filtered []source.Resource
		for _, r := range resources {
			if kindSet[r.Kind] {
				filtered = append(filtered, r)
			}
		}
		resources = filtered
	}

	// Filter by label selector
	if selector != "" {
		selectorLabels, err := parseSelector(selector)
		if err != nil {
			return fmt.Errorf("invalid selector: %w", err)
		}
		var filtered []source.Resource
		for _, r := range resources {
			if matchesLabels(r, selectorLabels) {
				filtered = append(filtered, r)
			}
		}
		resources = filtered
	}

	if len(resources) == 0 {
		fmt.Println("No resources found matching filters.")
		return nil
	}

	// Create cluster fetcher
	fetcher, err := cluster.NewFetcher(kubeconfig, kubeContext)
	if err != nil {
		return fmt.Errorf("failed to create cluster client: %w", err)
	}

	// Compare each resource
	var results []*diff.DiffResult
	for _, r := range resources {
		ctx := context.Background()
		clusterObj, err := fetcher.Get(ctx, r.APIVersion, r.Kind, r.Namespace, r.Name)
		if err != nil {
			// Resource not found in cluster → new
			result, compareErr := diff.Compare(r.Object, nil)
			if compareErr != nil {
				return compareErr
			}
			results = append(results, result)
			continue
		}

		result, compareErr := diff.Compare(r.Object, clusterObj)
		if compareErr != nil {
			return compareErr
		}
		results = append(results, result)
	}

	// Generate report
	summary := report.NewSummary(results)

	switch output {
	case "json":
		return summary.PrintJSON(os.Stdout)
	case "plain":
		summary.PrintPlain(os.Stdout)
	case "markdown":
		summary.PrintMarkdown(os.Stdout)
	default:
		if summaryOnly {
			summary.PrintSummaryOnly(os.Stdout)
		} else {
			summary.PrintColor(os.Stdout)
		}
	}

	// Exit with appropriate code
	if summary.HasChanges() {
		os.Exit(1)
	}
	return nil
}

// parseSelector parses a label selector string like "app=nginx,env=prod"
// into a map of key-value pairs.
func parseSelector(selector string) (map[string]string, error) {
	labels := make(map[string]string)
	parts := strings.Split(selector, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid label selector %q: must be key=value", part)
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if key == "" {
			return nil, fmt.Errorf("invalid label selector %q: empty key", part)
		}
		labels[key] = val
	}
	if len(labels) == 0 {
		return nil, fmt.Errorf("empty label selector")
	}
	return labels, nil
}

// matchesLabels checks if a resource's metadata.labels contain all selector labels.
func matchesLabels(r source.Resource, selectorLabels map[string]string) bool {
	if r.Object == nil {
		return false
	}
	labels := r.Object.GetLabels()
	if labels == nil {
		return false
	}
	for k, v := range selectorLabels {
		if labels[k] != v {
			return false
		}
	}
	return true
}
