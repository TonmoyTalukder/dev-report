package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/dev-report/dev-report/internal/types"
)

// Table prints the report as a formatted table to stdout.
func Table(out *types.ReportOutput) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	// Header info
	cyan.Printf("\n  Daily Work Report\n")
	fmt.Printf("  Date: %s", out.Date)
	if out.Developer != "" {
		fmt.Printf("  |  Developer: %s", out.Developer)
	}
	if out.CheckIn != "" && out.CheckOut != "" {
		fmt.Printf("  |  Check-in: %s  →  Check-out: %s", out.CheckIn, out.CheckOut)
	}
	if out.Adjusted != "" {
		yellow.Printf("  |  Adjusted: %s", out.Adjusted)
	}
	fmt.Println()

	// Table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Task", "Module", "Description", "Time Spent", "Status"})
	table.SetBorder(true)
	table.SetRowLine(false)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold},
	)
	table.SetColumnColor(
		tablewriter.Colors{},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgYellowColor},
		tablewriter.Colors{},
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgGreenColor},
	)
	table.SetColWidth(40)

	for _, t := range out.Tasks {
		table.Append([]string{
			fmt.Sprintf("%d", t.Number),
			wrapText(t.Title, 35),
			t.Module,
			wrapText(t.Description, 45),
			t.TimeSpent,
			t.Status,
		})
	}

	table.Render()

	// Footer
	separator := strings.Repeat("─", 60)
	fmt.Println(" " + separator)
	green.Printf("  Total: %s\n\n", out.TotalTime)
}

// wrapText inserts newlines into long strings for terminal display.
func wrapText(s string, width int) string {
	if len(s) <= width {
		return s
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
	return strings.Join(lines, "\n")
}
