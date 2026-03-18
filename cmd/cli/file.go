package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file [path]",
	Short: "Compare plain YAML manifests against cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("file command not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)
}
