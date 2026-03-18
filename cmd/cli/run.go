package cli

import (
	"context"
	"fmt"
	"os"

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
		clusterObj, err := fetcher.Get(ctx, r.APIVersion, r.Kind, r.Name, r.Namespace)
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
