package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kube-diff",
	Short: "Compare local Kubernetes manifests against live cluster state",
	Long: `kube-diff compares your local Kubernetes manifests (plain YAML, Helm charts,
or Kustomize overlays) against the actual state in your cluster, providing
a clear, colorized diff with a summary report.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("kubeconfig", "", "", "path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	rootCmd.PersistentFlags().StringP("context", "", "", "kubernetes context to use")
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "filter by namespace")
	rootCmd.PersistentFlags().StringSliceP("kind", "k", nil, "filter by resource kind (e.g., Deployment,Service)")
	rootCmd.PersistentFlags().StringP("selector", "l", "", "filter by label selector (e.g., app=nginx,env=prod)")
	rootCmd.PersistentFlags().BoolP("summary-only", "s", false, "show summary only, no diff details")
	rootCmd.PersistentFlags().StringP("output", "o", "color", "output format: color, plain, json, markdown, table")
	rootCmd.PersistentFlags().StringSlice("ignore-field", nil, "field paths to ignore in diff (e.g., metadata.annotations.some-key)")
	rootCmd.PersistentFlags().IntP("context-lines", "C", 3, "number of context lines in diff output")
	rootCmd.PersistentFlags().Bool("exit-code", false, "exit 0 even when changes are detected (disable exit 1)")
}
