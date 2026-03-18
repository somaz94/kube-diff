package cli

import (
	"github.com/somaz94/kube-diff/internal/source"
	"github.com/spf13/cobra"
)

var kustomizeCmd = &cobra.Command{
	Use:   "kustomize [path]",
	Short: "Compare Kustomize build output against cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := source.NewKustomizeSource(args[0])
		return runDiff(cmd, src)
	},
}

func init() {
	rootCmd.AddCommand(kustomizeCmd)
}
