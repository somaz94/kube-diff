package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/somaz94/kube-diff/internal/source"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [source] [path]",
	Short: "Watch for file changes and re-run diff automatically",
	Long: `Watch monitors the specified path for file changes and automatically
re-runs kube-diff when changes are detected.

Examples:
  kube-diff watch file ./manifests/
  kube-diff watch helm ./my-chart/ -f values.yaml
  kube-diff watch kustomize ./overlays/production/`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceType := args[0]
		path := args[1]

		interval, _ := cmd.Flags().GetDuration("interval")

		return runWatch(cmd, sourceType, path, interval)
	},
}

func init() {
	watchCmd.Flags().Duration("interval", 0, "minimum interval between re-runs (e.g., 5s, 1m). 0 means run on every change")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, sourceType, path string, interval time.Duration) error {
	// Validate source type
	switch sourceType {
	case "file", "helm", "kustomize":
	default:
		return fmt.Errorf("invalid source type %q: must be file, helm, or kustomize", sourceType)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	// Add path and subdirectories
	if err := addWatchPaths(watcher, absPath); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Watching %s for changes (press Ctrl+C to stop)...\n\n", absPath)

	// Initial run
	runDiffForWatch(cmd, sourceType, path)

	var lastRun time.Time
	debounceTimer := time.NewTimer(0)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if !isRelevantChange(event) {
				continue
			}
			// Debounce: wait 500ms after last change before re-running
			debounceTimer.Reset(500 * time.Millisecond)

		case <-debounceTimer.C:
			if interval > 0 && time.Since(lastRun) < interval {
				continue
			}
			fmt.Fprintf(os.Stderr, "\n--- Change detected at %s ---\n\n", time.Now().Format("15:04:05"))
			runDiffForWatch(cmd, sourceType, path)
			lastRun = time.Now()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "Watch error: %v\n", err)
		}
	}
}

func runDiffForWatch(cmd *cobra.Command, sourceType, path string) {
	src := createSource(cmd, sourceType, path)
	if src == nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create source\n")
		return
	}

	// Force --exit-code in watch mode to prevent os.Exit(1) on changes
	_ = cmd.Flags().Set("exit-code", "true")

	if err := runDiff(cmd, src); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func createSource(cmd *cobra.Command, sourceType, path string) source.Source {
	switch sourceType {
	case "file":
		return source.NewFileSource(path)
	case "helm":
		values, _ := cmd.Flags().GetStringSlice("values")
		release, _ := cmd.Flags().GetString("release")
		return source.NewHelmSource(path, release, values)
	case "kustomize":
		return source.NewKustomizeSource(path)
	default:
		return nil
	}
}

func addWatchPaths(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
}

func isRelevantChange(event fsnotify.Event) bool {
	// Only react to write and create events
	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
		return false
	}
	// Only watch yaml/yml files and Chart.yaml, kustomization.yaml
	name := filepath.Base(event.Name)
	ext := filepath.Ext(name)
	return ext == ".yaml" || ext == ".yml" || ext == ".json"
}
