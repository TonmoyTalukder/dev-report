package report

import (
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/dev-report/dev-report/internal/types"
)

// Excel writes the report to an .xlsx file at the given path.
func Excel(out *types.ReportOutput, filePath string) error {
	f := excelize.NewFile()
	sheet := "Work Report"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	// Styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "1F4E79"},
		Alignment: &excelize.Alignment{Horizontal: "left"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"2E75B6"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "thin", Color: "AAAAAA", Style: 1},
		},
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"},
		Border: []excelize.Border{
			{Type: "thin", Color: "DDDDDD", Style: 1},
		},
	})
	timeStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "2E75B6"},
		Alignment: &excelize.Alignment{Horizontal: "center", WrapText: true, Vertical: "top"},
		Border: []excelize.Border{
			{Type: "thin", Color: "DDDDDD", Style: 1},
		},
	})
	statusStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "375623"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E2EFDA"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "top"},
		Border: []excelize.Border{
			{Type: "thin", Color: "DDDDDD", Style: 1},
		},
	})
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "1F4E79"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D6E4F0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})

	// Row 1: Title
	f.SetCellValue(sheet, "A1", "Daily Work Report")
	f.SetCellStyle(sheet, "A1", "F1", titleStyle)
	f.MergeCell(sheet, "A1", "F1")

	// Row 2: Meta info
	meta := fmt.Sprintf("Date: %s", out.Date)
	if out.Developer != "" {
		meta += fmt.Sprintf("  |  Developer: %s", out.Developer)
	}
	if out.CheckIn != "" && out.CheckOut != "" {
		meta += fmt.Sprintf("  |  Check-in: %s  ->  Check-out: %s", out.CheckIn, out.CheckOut)
	}
	if out.Adjusted != "" {
		meta += fmt.Sprintf("  |  Adjusted: %s", out.Adjusted)
	}
	f.SetCellValue(sheet, "A2", meta)
	f.MergeCell(sheet, "A2", "F2")

	// Row 3: empty spacer
	// Row 4: Header
	headers := []string{"#", "Task", "Project", "Description", "Time Spent", "Status"}
	cols := []string{"A", "B", "C", "D", "E", "F"}
	for i, h := range headers {
		cell := cols[i] + "4"
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Column widths
	f.SetColWidth(sheet, "A", "A", 5)
	f.SetColWidth(sheet, "B", "B", 38)
	f.SetColWidth(sheet, "C", "C", 18)
	f.SetColWidth(sheet, "D", "D", 55)
	f.SetColWidth(sheet, "E", "E", 12)
	f.SetColWidth(sheet, "F", "F", 12)

	// Data rows
	for _, t := range out.Tasks {
		row := t.Number + 4 // offset: row 4 is header
		rowStr := fmt.Sprintf("%d", row)
		f.SetCellValue(sheet, "A"+rowStr, t.Number)
		f.SetCellValue(sheet, "B"+rowStr, t.Title)
		f.SetCellValue(sheet, "C"+rowStr, t.Project)
		f.SetCellValue(sheet, "D"+rowStr, t.Description)
		f.SetCellValue(sheet, "E"+rowStr, t.TimeSpent)
		f.SetCellValue(sheet, "F"+rowStr, t.Status)

		f.SetCellStyle(sheet, "A"+rowStr, "D"+rowStr, cellStyle)
		f.SetCellStyle(sheet, "E"+rowStr, "E"+rowStr, timeStyle)
		f.SetCellStyle(sheet, "F"+rowStr, "F"+rowStr, statusStyle)
		f.SetRowHeight(sheet, row, 30)
	}

	// Total row
	totalRow := len(out.Tasks) + 5
	totalRowStr := fmt.Sprintf("%d", totalRow)
	f.SetCellValue(sheet, "D"+totalRowStr, "Total:")
	f.SetCellValue(sheet, "E"+totalRowStr, out.TotalTime)
	f.SetCellStyle(sheet, "D"+totalRowStr, "E"+totalRowStr, totalStyle)
	f.MergeCell(sheet, "A"+totalRowStr, "D"+totalRowStr)

	return f.SaveAs(filePath)
}
