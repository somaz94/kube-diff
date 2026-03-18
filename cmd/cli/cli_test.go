package cli

import (
	"testing"
)

func TestExecute(t *testing.T) {
	// Execute with no args should print help and succeed
	err := Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileCommandRequiresArg(t *testing.T) {
	rootCmd.SetArgs([]string{"file"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when file command called without args")
	}
}

func TestHelmCommandRequiresArg(t *testing.T) {
	rootCmd.SetArgs([]string{"helm"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when helm command called without args")
	}
}

func TestKustomizeCommandRequiresArg(t *testing.T) {
	rootCmd.SetArgs([]string{"kustomize"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when kustomize command called without args")
	}
}

func TestFileCommandWithInvalidPath(t *testing.T) {
	rootCmd.SetArgs([]string{"file", "/tmp/nonexistent-kube-diff-test"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestHelmCommandWithInvalidChart(t *testing.T) {
	rootCmd.SetArgs([]string{"helm", "/tmp/nonexistent-kube-diff-chart"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent chart")
	}
}

func TestKustomizeCommandWithInvalidPath(t *testing.T) {
	rootCmd.SetArgs([]string{"kustomize", "/tmp/nonexistent-kube-diff-overlay"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent overlay")
	}
}

func TestRootCommandFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	tests := []struct {
		name     string
		flag     string
		defValue string
	}{
		{"kubeconfig", "kubeconfig", ""},
		{"context", "context", ""},
		{"namespace", "namespace", ""},
		{"summary-only", "summary-only", "false"},
		{"output", "output", "color"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Fatalf("flag %q not found", tt.flag)
			}
			if f.DefValue != tt.defValue {
				t.Errorf("flag %q default: got %q, want %q", tt.flag, f.DefValue, tt.defValue)
			}
		})
	}
}

func TestHelmCommandFlags(t *testing.T) {
	f := helmCmd.Flags().Lookup("values")
	if f == nil {
		t.Fatal("values flag not found on helm command")
	}

	r := helmCmd.Flags().Lookup("release")
	if r == nil {
		t.Fatal("release flag not found on helm command")
	}
	if r.DefValue != "release" {
		t.Errorf("expected release default=release, got %s", r.DefValue)
	}
}

func TestUnknownCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"unknown-cmd"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}
