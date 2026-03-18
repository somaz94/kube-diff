package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helmCmd = &cobra.Command{
	Use:   "helm [chart-path]",
	Short: "Compare Helm chart template output against cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("helm command not yet implemented")
		return nil
	},
}

func init() {
	helmCmd.Flags().StringSliceP("values", "f", nil, "values files")
	helmCmd.Flags().StringP("release", "r", "release", "release name for helm template")
	rootCmd.AddCommand(helmCmd)
}
