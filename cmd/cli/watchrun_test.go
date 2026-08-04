package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// captureStderr runs fn with os.Stderr redirected and returns what it wrote.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	fn()

	_ = w.Close()
	os.Stderr = orig
	return <-done
}

// newWatchCmd builds a command carrying the flags the watch path reads, as
// local flags so they resolve without going through cobra execution. Using a
// dedicated command also keeps the shared rootCmd free of test mutations.
func newWatchCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("exit-code", false, "")
	cmd.Flags().String("kubeconfig", "", "")
	cmd.Flags().String("context", "", "")
	cmd.Flags().StringSlice("values", nil, "")
	cmd.Flags().String("release", "", "")
	return cmd
}

// manifestDir writes one valid manifest so the diff gets past resource loading.
func manifestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	body := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: probe\n  namespace: default\n"
	if err := os.WriteFile(filepath.Join(dir, "cm.yaml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

// An unknown source type yields no source; the run must report it and return
// rather than proceeding with a nil source.
func TestRunDiffForWatch_ReportsMissingSource(t *testing.T) {
	out := captureStderr(t, func() {
		runDiffForWatch(newWatchCmd(t), "invalid", "/tmp")
	})

	if !strings.Contains(out, "failed to create source") {
		t.Errorf("stderr = %q, want it to report the missing source", out)
	}
}

// A diff failure inside watch mode is reported and swallowed: the watcher has
// to survive it, so runDiffForWatch must neither propagate nor exit.
func TestRunDiffForWatch_ReportsDiffFailure(t *testing.T) {
	cmd := newWatchCmd(t)
	if err := cmd.Flags().Set("kubeconfig", "/nonexistent-kubeconfig"); err != nil {
		t.Fatal(err)
	}

	out := captureStderr(t, func() {
		runDiffForWatch(cmd, "file", manifestDir(t))
	})

	if !strings.Contains(out, "Error:") {
		t.Errorf("stderr = %q, want the diff failure reported", out)
	}
}

// Watch mode forces --exit-code so a detected change cannot terminate the
// watcher through os.Exit(1).
func TestRunDiffForWatch_ForcesExitCodeFlag(t *testing.T) {
	cmd := newWatchCmd(t)
	if err := cmd.Flags().Set("kubeconfig", "/nonexistent-kubeconfig"); err != nil {
		t.Fatal(err)
	}
	if got := cmd.Flags().Lookup("exit-code").Value.String(); got != "false" {
		t.Fatalf("exit-code = %q before the run, want %q", got, "false")
	}

	captureStderr(t, func() {
		runDiffForWatch(cmd, "file", manifestDir(t))
	})

	if got := cmd.Flags().Lookup("exit-code").Value.String(); got != "true" {
		t.Errorf("exit-code = %q after a watch run, want %q", got, "true")
	}
}
