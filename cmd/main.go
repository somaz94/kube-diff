package main

import (
	"fmt"
	"os"

	"github.com/somaz94/kube-diff/cmd/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		// 1 is reserved for "changes detected" (cli.ErrChangesDetected).
		os.Exit(2)
	}
}
