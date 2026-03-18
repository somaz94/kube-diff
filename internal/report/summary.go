package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/somaz94/kube-diff/internal/diff"
)

// Summary holds aggregated diff results.
type Summary struct {
	Results   []*diff.DiffResult
	New       int
	Changed   int
	Unchanged int
	Deleted   int
	Total     int
}

// NewSummary creates a summary from diff results.
func NewSummary(results []*diff.DiffResult) *Summary {
	s := &Summary{
		Results: results,
		Total:   len(results),
	}

	for _, r := range results {
		switch r.Status {
		case diff.StatusNew:
			s.New++
		case diff.StatusChanged:
			s.Changed++
		case diff.StatusUnchanged:
			s.Unchanged++
		case diff.StatusDeleted:
			s.Deleted++
		}
	}

	return s
}

// HasChanges returns true if there are any differences.
func (s *Summary) HasChanges() bool {
	return s.New > 0 || s.Changed > 0 || s.Deleted > 0
}

// ExitCode returns the appropriate exit code.
func (s *Summary) ExitCode() int {
	if s.HasChanges() {
		return 1
	}
	return 0
}

// PrintColor writes a colorized report to the writer.
func (s *Summary) PrintColor(w io.Writer) {
	for _, r := range s.Results {
		switch r.Status {
		case diff.StatusNew:
			fmt.Fprintf(w, "\033[32m★ NEW    %s\033[0m\n", r.ResourceKey())
		case diff.StatusChanged:
			fmt.Fprintf(w, "\033[33m~ CHANGED %s\033[0m\n", r.ResourceKey())
			fmt.Fprintln(w, colorizeDiff(r.Diff))
		case diff.StatusUnchanged:
			fmt.Fprintf(w, "\033[90m✓ OK     %s\033[0m\n", r.ResourceKey())
		case diff.StatusDeleted:
			fmt.Fprintf(w, "\033[31m✗ DELETED %s\033[0m\n", r.ResourceKey())
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "Summary: %d resources — ", s.Total)
	parts := []string{}
	if s.Changed > 0 {
		parts = append(parts, fmt.Sprintf("\033[33m%d changed\033[0m", s.Changed))
	}
	if s.New > 0 {
		parts = append(parts, fmt.Sprintf("\033[32m%d new\033[0m", s.New))
	}
	if s.Deleted > 0 {
		parts = append(parts, fmt.Sprintf("\033[31m%d deleted\033[0m", s.Deleted))
	}
	if s.Unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", s.Unchanged))
	}
	fmt.Fprintln(w, strings.Join(parts, ", "))
}

// PrintJSON writes a JSON report to the writer.
func (s *Summary) PrintJSON(w io.Writer) error {
	type jsonResult struct {
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace,omitempty"`
		Status    string `json:"status"`
	}

	type jsonReport struct {
		Total     int            `json:"total"`
		Changed   int            `json:"changed"`
		New       int            `json:"new"`
		Deleted   int            `json:"deleted"`
		Unchanged int            `json:"unchanged"`
		Resources []jsonResult   `json:"resources"`
	}

	report := jsonReport{
		Total:     s.Total,
		Changed:   s.Changed,
		New:       s.New,
		Deleted:   s.Deleted,
		Unchanged: s.Unchanged,
	}

	for _, r := range s.Results {
		report.Resources = append(report.Resources, jsonResult{
			Kind:      r.Kind,
			Name:      r.Name,
			Namespace: r.Namespace,
			Status:    string(r.Status),
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// PrintPlain writes a plain-text report (no ANSI colors) to the writer.
func (s *Summary) PrintPlain(w io.Writer) {
	for _, r := range s.Results {
		switch r.Status {
		case diff.StatusNew:
			fmt.Fprintf(w, "* NEW    %s\n", r.ResourceKey())
		case diff.StatusChanged:
			fmt.Fprintf(w, "~ CHANGED %s\n", r.ResourceKey())
			fmt.Fprintln(w, r.Diff)
		case diff.StatusUnchanged:
			fmt.Fprintf(w, "  OK     %s\n", r.ResourceKey())
		case diff.StatusDeleted:
			fmt.Fprintf(w, "x DELETED %s\n", r.ResourceKey())
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "Summary: %d resources — ", s.Total)
	parts := []string{}
	if s.Changed > 0 {
		parts = append(parts, fmt.Sprintf("%d changed", s.Changed))
	}
	if s.New > 0 {
		parts = append(parts, fmt.Sprintf("%d new", s.New))
	}
	if s.Deleted > 0 {
		parts = append(parts, fmt.Sprintf("%d deleted", s.Deleted))
	}
	if s.Unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", s.Unchanged))
	}
	fmt.Fprintln(w, strings.Join(parts, ", "))
}

// PrintMarkdown writes a markdown-formatted report to the writer.
func (s *Summary) PrintMarkdown(w io.Writer) {
	fmt.Fprintln(w, "## kube-diff Report")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "**%d** resources — ", s.Total)
	parts := []string{}
	if s.Changed > 0 {
		parts = append(parts, fmt.Sprintf("**%d** changed", s.Changed))
	}
	if s.New > 0 {
		parts = append(parts, fmt.Sprintf("**%d** new", s.New))
	}
	if s.Deleted > 0 {
		parts = append(parts, fmt.Sprintf("**%d** deleted", s.Deleted))
	}
	if s.Unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", s.Unchanged))
	}
	fmt.Fprintln(w, strings.Join(parts, ", "))
	fmt.Fprintln(w)

	fmt.Fprintln(w, "| Status | Resource | Namespace |")
	fmt.Fprintln(w, "|--------|----------|-----------|")
	for _, r := range s.Results {
		status := ""
		switch r.Status {
		case diff.StatusNew:
			status = "🟢 NEW"
		case diff.StatusChanged:
			status = "🟡 CHANGED"
		case diff.StatusUnchanged:
			status = "⚪ OK"
		case diff.StatusDeleted:
			status = "🔴 DELETED"
		}
		ns := r.Namespace
		if ns == "" {
			ns = "-"
		}
		fmt.Fprintf(w, "| %s | %s/%s | %s |\n", status, r.Kind, r.Name, ns)
	}

	// Show diffs for changed resources
	for _, r := range s.Results {
		if r.Status == diff.StatusChanged && r.Diff != "" {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "### %s/%s\n", r.Kind, r.Name)
			fmt.Fprintln(w)
			fmt.Fprintln(w, "```diff")
			fmt.Fprintln(w, r.Diff)
			fmt.Fprintln(w, "```")
		}
	}
}

// PrintSummaryOnly writes only the summary line to the writer.
func (s *Summary) PrintSummaryOnly(w io.Writer) {
	fmt.Fprintf(w, "Summary: %d resources — ", s.Total)
	parts := []string{}
	if s.Changed > 0 {
		parts = append(parts, fmt.Sprintf("\033[33m%d changed\033[0m", s.Changed))
	}
	if s.New > 0 {
		parts = append(parts, fmt.Sprintf("\033[32m%d new\033[0m", s.New))
	}
	if s.Deleted > 0 {
		parts = append(parts, fmt.Sprintf("\033[31m%d deleted\033[0m", s.Deleted))
	}
	if s.Unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", s.Unchanged))
	}
	fmt.Fprintln(w, strings.Join(parts, ", "))
}

// PrintTable writes a table-formatted report to the writer.
func (s *Summary) PrintTable(w io.Writer) {
	// Header
	fmt.Fprintf(w, "%-10s %-20s %-30s %-15s\n", "STATUS", "KIND", "NAME", "NAMESPACE")
	fmt.Fprintf(w, "%-10s %-20s %-30s %-15s\n", "------", "----", "----", "---------")

	for _, r := range s.Results {
		status := ""
		switch r.Status {
		case diff.StatusNew:
			status = "NEW"
		case diff.StatusChanged:
			status = "CHANGED"
		case diff.StatusUnchanged:
			status = "OK"
		case diff.StatusDeleted:
			status = "DELETED"
		}
		ns := r.Namespace
		if ns == "" {
			ns = "-"
		}
		fmt.Fprintf(w, "%-10s %-20s %-30s %-15s\n", status, r.Kind, r.Name, ns)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "Total: %d | Changed: %d | New: %d | Deleted: %d | Unchanged: %d\n",
		s.Total, s.Changed, s.New, s.Deleted, s.Unchanged)
}

func colorizeDiff(text string) string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		switch {
		case strings.HasPrefix(line, "+"):
			lines = append(lines, fmt.Sprintf("\033[32m%s\033[0m", line))
		case strings.HasPrefix(line, "-"):
			lines = append(lines, fmt.Sprintf("\033[31m%s\033[0m", line))
		case strings.HasPrefix(line, "@@"):
			lines = append(lines, fmt.Sprintf("\033[36m%s\033[0m", line))
		default:
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
