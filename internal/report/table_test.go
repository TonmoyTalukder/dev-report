package report

import (
	"strings"
	"testing"

	"github.com/dev-report/dev-report/internal/types"
)

func TestTextIsCopyPasteFriendly(t *testing.T) {
	out := &types.ReportOutput{
		Date:      "2026-03-08",
		TotalTime: "30m",
		Tasks: []*types.Task{
			{
				Number:      1,
				Title:       "Improved package download",
				Project:     "dev-report",
				Description: "Made downloads more reliable",
				TimeSpent:   "30m",
				Status:      "Completed",
			},
		},
	}

	got := Text(out)

	if strings.Contains(got, "+---") || strings.Contains(got, "| # |") {
		t.Fatalf("expected text output without boxed table borders, got %q", got)
	}
	if !strings.Contains(got, "#  TASK | Project | Description | Time Spent | Status") {
		t.Fatalf("expected compact header row, got %q", got)
	}
	if !strings.Contains(got, "1. Improved package download | dev-report | Made downloads more reliable | 30m | Completed") {
		t.Fatalf("expected single-line task row, got %q", got)
	}
	if strings.Contains(got, "Daily Work Report") || strings.Contains(got, "Check-in:") {
		t.Fatalf("expected no extra report metadata in compact table output, got %q", got)
	}
	if !strings.Contains(got, "Total: 30m") {
		t.Fatalf("expected total time in output, got %q", got)
	}
}
