package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/somaz94/kube-diff/internal/diff"
)

// ANSI color codes
const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorReset  = "\033[0m"
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

// statusLabel returns a display label for the given status and format.
type statusFormat struct {
	newLabel       string
	changedLabel   string
	unchangedLabel string
	deletedLabel   string
}

var (
	colorStatusFormat = statusFormat{
		newLabel:       colorGreen + "★ NEW    " + colorReset,
		changedLabel:   colorYellow + "~ CHANGED " + colorReset,
		unchangedLabel: colorGray + "✓ OK     " + colorReset,
		deletedLabel:   colorRed + "✗ DELETED " + colorReset,
	}
	plainStatusFormat = statusFormat{
		newLabel:       "* NEW    ",
		changedLabel:   "~ CHANGED ",
		unchangedLabel: "  OK     ",
		deletedLabel:   "x DELETED ",
	}
)

func (sf statusFormat) label(status diff.DiffStatus) string {
	switch status {
	case diff.StatusNew:
		return sf.newLabel
	case diff.StatusChanged:
		return sf.changedLabel
	case diff.StatusUnchanged:
		return sf.unchangedLabel
	case diff.StatusDeleted:
		return sf.deletedLabel
	default:
		return ""
	}
}

// buildSummaryParts builds the summary count parts.
func (s *Summary) buildSummaryParts(colorize bool) []string {
	var parts []string
	if s.Changed > 0 {
		if colorize {
			parts = append(parts, fmt.Sprintf(colorYellow+"%d changed"+colorReset, s.Changed))
		} else {
			parts = append(parts, fmt.Sprintf("%d changed", s.Changed))
		}
	}
	if s.New > 0 {
		if colorize {
			parts = append(parts, fmt.Sprintf(colorGreen+"%d new"+colorReset, s.New))
		} else {
			parts = append(parts, fmt.Sprintf("%d new", s.New))
		}
	}
	if s.Deleted > 0 {
		if colorize {
			parts = append(parts, fmt.Sprintf(colorRed+"%d deleted"+colorReset, s.Deleted))
		} else {
			parts = append(parts, fmt.Sprintf("%d deleted", s.Deleted))
		}
	}
	if s.Unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", s.Unchanged))
	}
	return parts
}

// writeSummaryLine writes the "Summary: N resources — ..." line.
func (s *Summary) writeSummaryLine(w io.Writer, colorize bool) {
	fmt.Fprintf(w, "Summary: %d resources — ", s.Total)
	fmt.Fprintln(w, strings.Join(s.buildSummaryParts(colorize), ", "))
}

// printResults iterates results and prints each with the given format.
func (s *Summary) printResults(w io.Writer, sf statusFormat, showDiff func(string) string) {
	for _, r := range s.Results {
		fmt.Fprintf(w, "%s%s\n", sf.label(r.Status), r.ResourceKey())
		if r.Status == diff.StatusChanged && r.Diff != "" {
			fmt.Fprintln(w, showDiff(r.Diff))
		}
	}
}

// PrintColor writes a colorized report to the writer.
func (s *Summary) PrintColor(w io.Writer) {
	s.printResults(w, colorStatusFormat, colorizeDiff)
	fmt.Fprintln(w)
	s.writeSummaryLine(w, true)
}

// PrintPlain writes a plain-text report (no ANSI colors) to the writer.
func (s *Summary) PrintPlain(w io.Writer) {
	s.printResults(w, plainStatusFormat, func(d string) string { return d })
	fmt.Fprintln(w)
	s.writeSummaryLine(w, false)
}

// PrintSummaryOnly writes only the summary line to the writer.
func (s *Summary) PrintSummaryOnly(w io.Writer) {
	s.writeSummaryLine(w, true)
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
		Total     int          `json:"total"`
		Changed   int          `json:"changed"`
		New       int          `json:"new"`
		Deleted   int          `json:"deleted"`
		Unchanged int          `json:"unchanged"`
		Resources []jsonResult `json:"resources"`
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

// PrintMarkdown writes a markdown-formatted report to the writer.
func (s *Summary) PrintMarkdown(w io.Writer) {
	fmt.Fprintln(w, "## kube-diff Report")
	fmt.Fprintln(w)

	// Summary with bold counts
	fmt.Fprintf(w, "**%d** resources — ", s.Total)
	var parts []string
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

	// Status table
	mdStatus := map[diff.DiffStatus]string{
		diff.StatusNew:       "🟢 NEW",
		diff.StatusChanged:   "🟡 CHANGED",
		diff.StatusUnchanged: "⚪ OK",
		diff.StatusDeleted:   "🔴 DELETED",
	}

	fmt.Fprintln(w, "| Status | Resource | Namespace |")
	fmt.Fprintln(w, "|--------|----------|-----------|")
	for _, r := range s.Results {
		ns := r.Namespace
		if ns == "" {
			ns = "-"
		}
		fmt.Fprintf(w, "| %s | %s/%s | %s |\n", mdStatus[r.Status], r.Kind, r.Name, ns)
	}

	// Diffs for changed resources
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

// PrintTable writes a table-formatted report to the writer.
func (s *Summary) PrintTable(w io.Writer) {
	tableStatus := map[diff.DiffStatus]string{
		diff.StatusNew:       "NEW",
		diff.StatusChanged:   "CHANGED",
		diff.StatusUnchanged: "OK",
		diff.StatusDeleted:   "DELETED",
	}

	fmt.Fprintf(w, "%-10s %-20s %-30s %-15s\n", "STATUS", "KIND", "NAME", "NAMESPACE")
	fmt.Fprintf(w, "%-10s %-20s %-30s %-15s\n", "------", "----", "----", "---------")

	for _, r := range s.Results {
		ns := r.Namespace
		if ns == "" {
			ns = "-"
		}
		fmt.Fprintf(w, "%-10s %-20s %-30s %-15s\n", tableStatus[r.Status], r.Kind, r.Name, ns)
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
			lines = append(lines, colorGreen+line+colorReset)
		case strings.HasPrefix(line, "-"):
			lines = append(lines, colorRed+line+colorReset)
		case strings.HasPrefix(line, "@@"):
			lines = append(lines, colorCyan+line+colorReset)
		default:
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
