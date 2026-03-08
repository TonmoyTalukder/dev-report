package report

import (
	"fmt"
	"strings"

	"github.com/dev-report/dev-report/internal/types"
)

// Table prints the report as a formatted table to stdout.
func Table(out *types.ReportOutput) {
	fmt.Print(Text(out))
}

func Text(out *types.ReportOutput) string {
	var sb strings.Builder

	sb.WriteString("\n#  TASK | Project | Description | Time Spent | Status\n")

	for _, task := range out.Tasks {
		sb.WriteString(fmt.Sprintf("%d. %s | %s | %s | %s | %s\n",
			task.Number,
			sanitizeCell(task.Title),
			sanitizeCell(task.Project),
			sanitizeCell(task.Description),
			sanitizeCell(task.TimeSpent),
			sanitizeCell(task.Status),
		))
	}

	if out.TotalTime != "" {
		sb.WriteString(fmt.Sprintf("\nTotal: %s\n", out.TotalTime))
	}

	return sb.String()
}

func sanitizeCell(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	value = strings.ReplaceAll(value, "|", "/")
	if value == "" {
		return "-"
	}
	return value
}
