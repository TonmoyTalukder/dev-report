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

	sb.WriteString("\nDaily Work Report\n")
	sb.WriteString(fmt.Sprintf("Date: %s", out.Date))
	if out.Developer != "" {
		sb.WriteString(fmt.Sprintf("  |  Developer: %s", out.Developer))
	}
	if out.CheckIn != "" && out.CheckOut != "" {
		sb.WriteString(fmt.Sprintf("  |  Check-in: %s  ->  Check-out: %s", out.CheckIn, out.CheckOut))
	}
	if out.Adjusted != "" {
		sb.WriteString(fmt.Sprintf("  |  Adjusted: %s", out.Adjusted))
	}
	sb.WriteString("\n")
	if out.CommitCount > 0 || out.TaskCount > 0 {
		sb.WriteString(fmt.Sprintf("Commits: %d  |  Tasks: %d\n", out.CommitCount, out.TaskCount))
	}

	for i, task := range out.Tasks {
		if i == 0 {
			sb.WriteString("\n")
		} else {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%d. %s\n", task.Number, task.Title))

		meta := make([]string, 0, 3)
		if task.Module != "" {
			meta = append(meta, fmt.Sprintf("Module: %s", task.Module))
		}
		if task.TimeSpent != "" {
			meta = append(meta, fmt.Sprintf("Time: %s", task.TimeSpent))
		}
		if task.Status != "" {
			meta = append(meta, fmt.Sprintf("Status: %s", task.Status))
		}
		if len(meta) > 0 {
			sb.WriteString("   ")
			sb.WriteString(strings.Join(meta, "  |  "))
			sb.WriteString("\n")
		}

		if task.Description != "" {
			lines := wrapLines(task.Description, 72)
			if len(lines) > 0 {
				sb.WriteString("   Description: ")
				sb.WriteString(lines[0])
				sb.WriteString("\n")
				for _, line := range lines[1:] {
					sb.WriteString("                ")
					sb.WriteString(line)
					sb.WriteString("\n")
				}
			}
		}
	}

	if out.TotalTime != "" {
		sb.WriteString(fmt.Sprintf("\nTotal: %s\n\n", out.TotalTime))
	}

	return sb.String()
}

// wrapText inserts newlines into long strings for terminal display.
func wrapText(s string, width int) string {
	return strings.Join(wrapLines(s, width), "\n")
}

func wrapLines(s string, width int) []string {
	if len(s) <= width {
		return []string{s}
	}
	words := strings.Fields(s)
	var lines []string
	current := ""
	for _, w := range words {
		if len(current)+len(w)+1 > width {
			if current != "" {
				lines = append(lines, current)
			}
			current = w
		} else {
			if current == "" {
				current = w
			} else {
				current += " " + w
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
