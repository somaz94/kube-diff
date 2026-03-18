package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var kustomizeCmd = &cobra.Command{
	Use:   "kustomize [path]",
	Short: "Compare Kustomize build output against cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("kustomize command not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(kustomizeCmd)
}
