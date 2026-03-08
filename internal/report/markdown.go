package report

import (
	"fmt"
	"strings"

	"github.com/dev-report/dev-report/internal/types"
)

// Markdown renders the report as a Markdown document string.
func Markdown(out *types.ReportOutput) string {
	var sb strings.Builder

	sb.WriteString("# Daily Work Report\n\n")
	sb.WriteString(fmt.Sprintf("**Date:** %s", out.Date))
	if out.Developer != "" {
		sb.WriteString(fmt.Sprintf("  |  **Developer:** %s", out.Developer))
	}
	if out.CheckIn != "" && out.CheckOut != "" {
		sb.WriteString(fmt.Sprintf("  |  **Check-in:** %s  |  **Check-out:** %s", out.CheckIn, out.CheckOut))
	}
	if out.Adjusted != "" {
		sb.WriteString(fmt.Sprintf("  |  **Adjusted:** %s", out.Adjusted))
	}
	if out.TotalTime != "" {
		sb.WriteString(fmt.Sprintf("  |  **Total task time:** %s", out.TotalTime))
	}
	sb.WriteString("\n\n")
	if out.CommitCount > 0 || out.TaskCount > 0 {
		sb.WriteString(fmt.Sprintf("**Commits:** %d  |  **Tasks:** %d\n\n", out.CommitCount, out.TaskCount))
	}

	sb.WriteString("| # | Task | Project | Description | Time Spent | Status |\n")
	sb.WriteString("|---|------|--------|-------------|------------|--------|\n")
	for _, t := range out.Tasks {
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s |\n",
			t.Number,
			escapeMD(t.Title),
			escapeMD(t.Project),
			escapeMD(t.Description),
			t.TimeSpent,
			t.Status,
		))
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %s\n", out.TotalTime))
	return sb.String()
}

// escapeMD escapes pipe characters in markdown table cells.
func escapeMD(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}
