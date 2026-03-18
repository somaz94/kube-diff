package cli

import (
	"github.com/somaz94/kube-diff/internal/source"
	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file [path]",
	Short: "Compare plain YAML manifests against cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := source.NewFileSource(args[0])
		return runDiff(cmd, src)
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)
}
