package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/somaz94/kube-diff/internal/cluster"
	"github.com/somaz94/kube-diff/internal/diff"
	"github.com/somaz94/kube-diff/internal/report"
	"github.com/somaz94/kube-diff/internal/source"
	"github.com/spf13/cobra"
)

// diffFlags holds all extracted CLI flags for a diff run.
type diffFlags struct {
	kubeconfig   string
	kubeContext  string
	namespace    string
	kinds        []string
	names        []string
	selector     string
	summaryOnly  bool
	output       string
	ignoreFields []string
	contextLines int
	noExitCode   bool
	diffStrategy string
}

// extractFlags reads all relevant flags from the cobra command.
func extractFlags(cmd *cobra.Command) diffFlags {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	kubeContext, _ := cmd.Flags().GetString("context")
	namespace, _ := cmd.Flags().GetString("namespace")
	kinds, _ := cmd.Flags().GetStringSlice("kind")
	names, _ := cmd.Flags().GetStringSlice("name")
	selector, _ := cmd.Flags().GetString("selector")
	summaryOnly, _ := cmd.Flags().GetBool("summary-only")
	output, _ := cmd.Flags().GetString("output")
	ignoreFields, _ := cmd.Flags().GetStringSlice("ignore-field")
	contextLines, _ := cmd.Flags().GetInt("context-lines")
	noExitCode, _ := cmd.Flags().GetBool("exit-code")
	diffStrategy, _ := cmd.Flags().GetString("diff-strategy")

	return diffFlags{
		kubeconfig:   kubeconfig,
		kubeContext:  kubeContext,
		namespace:    namespace,
		kinds:        kinds,
		names:        names,
		selector:     selector,
		summaryOnly:  summaryOnly,
		output:       output,
		ignoreFields: ignoreFields,
		contextLines: contextLines,
		noExitCode:   noExitCode,
		diffStrategy: diffStrategy,
	}
}

// filterResources applies a predicate function to filter resources.
func filterResources(resources []source.Resource, predicate func(source.Resource) bool) []source.Resource {
	var filtered []source.Resource
	for _, r := range resources {
		if predicate(r) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// toStringSet converts a string slice to a set for O(1) lookups.
func toStringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

// applyFilters applies namespace, kind, name, and selector filters to resources.
func applyFilters(resources []source.Resource, f diffFlags) ([]source.Resource, error) {
	if f.namespace != "" {
		resources = filterResources(resources, func(r source.Resource) bool {
			return r.Namespace == f.namespace || r.Namespace == ""
		})
	}

	if len(f.kinds) > 0 {
		kindSet := toStringSet(f.kinds)
		resources = filterResources(resources, func(r source.Resource) bool {
			return kindSet[r.Kind]
		})
	}

	if len(f.names) > 0 {
		nameSet := toStringSet(f.names)
		resources = filterResources(resources, func(r source.Resource) bool {
			return nameSet[r.Name]
		})
	}

	if f.selector != "" {
		selectorLabels, err := parseSelector(f.selector)
		if err != nil {
			return nil, fmt.Errorf("invalid selector: %w", err)
		}
		resources = filterResources(resources, func(r source.Resource) bool {
			return matchesLabels(r, selectorLabels)
		})
	}

	return resources, nil
}

// buildCompareOptions constructs CompareOptions from flags.
func buildCompareOptions(f diffFlags) diff.CompareOptions {
	strategy := diff.StrategyLive
	if f.diffStrategy == "last-applied" {
		strategy = diff.StrategyLastApplied
	}
	return diff.CompareOptions{
		ContextLines: f.contextLines,
		IgnoreFields: f.ignoreFields,
		Strategy:     strategy,
	}
}

// printReport outputs the summary in the requested format.
func printReport(w io.Writer, summary *report.Summary, f diffFlags) error {
	if f.summaryOnly {
		summary.PrintSummaryOnly(w)
		return nil
	}
	switch f.output {
	case "json":
		return summary.PrintJSON(w)
	case "plain":
		summary.PrintPlain(w)
	case "markdown":
		summary.PrintMarkdown(w)
	case "table":
		summary.PrintTable(w)
	default:
		summary.PrintColor(w)
	}
	return nil
}

// runDiff is the shared logic for file, helm, and kustomize commands.
func runDiff(cmd *cobra.Command, src source.Source) error {
	f := extractFlags(cmd)

	resources, err := src.Load()
	if err != nil {
		return fmt.Errorf("failed to load resources: %w", err)
	}

	resources, err = applyFilters(resources, f)
	if err != nil {
		return err
	}

	if len(resources) == 0 {
		fmt.Println("No resources found matching filters.")
		return nil
	}

	fetcher, err := cluster.NewFetcher(f.kubeconfig, f.kubeContext)
	if err != nil {
		return fmt.Errorf("failed to create cluster client: %w", err)
	}

	err = executeDiff(cmd.Context(), fetcher, resources, f, os.Stdout)
	if errors.Is(err, ErrChangesDetected) {
		os.Exit(1)
	}
	return err
}

// executeDiff runs comparison, prints report, and returns ErrChangesDetected
// when differences are found and noExitCode is not set.
func executeDiff(ctx context.Context, fetcher cluster.ResourceFetcher, resources []source.Resource, f diffFlags, w io.Writer) error {
	opts := buildCompareOptions(f)
	results, err := compareResources(ctx, fetcher, resources, opts)
	if err != nil {
		return err
	}

	summary := report.NewSummary(results)
	if err := printReport(w, summary, f); err != nil {
		return err
	}

	if summary.HasChanges() && !f.noExitCode {
		return ErrChangesDetected
	}
	return nil
}

// compareResources compares local resources against the cluster using the given fetcher.
func compareResources(ctx context.Context, fetcher cluster.ResourceFetcher, resources []source.Resource, opts diff.CompareOptions) ([]*diff.DiffResult, error) {
	var results []*diff.DiffResult
	for _, r := range resources {
		clusterObj, err := fetcher.Get(ctx, r.APIVersion, r.Kind, r.Namespace, r.Name)
		if err != nil {
			// Resource not found in cluster → new
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
