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
	// main prints the error once; cobra would print it a second time.
	SilenceErrors: true,
	// Runs after flag and argument validation, so usage still shows for those
	// errors but not for runtime failures such as a missing path.
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		cmd.SilenceUsage = true
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("kubeconfig", "", "", "path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	rootCmd.PersistentFlags().StringP("context", "", "", "kubernetes context to use")
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "namespace for manifests that set none (default: kubeconfig context, then \"default\"), and filter by namespace")
	rootCmd.PersistentFlags().StringSliceP("kind", "k", nil, "filter by resource kind (e.g., Deployment,Service)")
	rootCmd.PersistentFlags().StringSliceP("name", "N", nil, "filter by resource name (e.g., my-app,my-config)")
	rootCmd.PersistentFlags().StringP("selector", "l", "", "filter by label selector (e.g., app=nginx,env=prod)")
	rootCmd.PersistentFlags().BoolP("summary-only", "s", false, "show summary only, no diff details")
	rootCmd.PersistentFlags().StringP("output", "o", "color", "output format: color, plain, json, markdown, table")
	rootCmd.PersistentFlags().StringSlice("ignore-field", nil, "field paths to ignore in diff (e.g., metadata.annotations.some-key)")
	rootCmd.PersistentFlags().IntP("context-lines", "C", 3, "number of context lines in diff output")
	rootCmd.PersistentFlags().Bool("exit-code", false, "exit 0 even when changes are detected (disable exit 1)")
	rootCmd.PersistentFlags().Bool("server-side", false, "diff a server-side apply dry run of each manifest, so defaults and mutating webhooks are included")
	rootCmd.PersistentFlags().String("field-manager", "kube-diff", "field manager for --server-side; use the one you apply with, so fields removed from a manifest show as removed")
	rootCmd.PersistentFlags().Bool("deleted", false, "also report cluster objects matching --selector that the manifests no longer contain")
	rootCmd.PersistentFlags().Bool("show-secrets", false, "print Secret values in the diff (redacted by default)")
	rootCmd.PersistentFlags().String("diff-strategy", "live", "comparison strategy: live (cluster state) or last-applied (last-applied-configuration annotation)")
}
