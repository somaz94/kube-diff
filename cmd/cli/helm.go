package cli

import (
	"github.com/somaz94/kube-diff/internal/source"
	"github.com/spf13/cobra"
)

var helmCmd = &cobra.Command{
	Use:   "helm [chart-path]",
	Short: "Compare Helm chart template output against cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		values, _ := cmd.Flags().GetStringSlice("values")
		release, _ := cmd.Flags().GetString("release")
		src := source.NewHelmSource(args[0], release, values)
		return runDiff(cmd, src)
	},
}

func init() {
	helmCmd.Flags().StringSliceP("values", "f", nil, "values files")
	helmCmd.Flags().StringP("release", "r", "release", "release name for helm template")
	rootCmd.AddCommand(helmCmd)
}
