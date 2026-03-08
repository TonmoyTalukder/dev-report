package report

import (
	"strings"
	"testing"

	"github.com/dev-report/dev-report/internal/types"
)

func TestTextIsCopyPasteFriendly(t *testing.T) {
	out := &types.ReportOutput{
		Date:        "2026-03-08",
		Developer:   "TonmoyTalukder",
		CheckIn:     "09:00",
		CheckOut:    "18:00",
		CommitCount: 2,
		TaskCount:   1,
		TotalTime:   "30m",
		Tasks: []*types.Task{
			{
				Number:      1,
				Title:       "Improved package download",
				Module:      "Npm",
				Description: "Made the package download process more reliable for Windows installs.",
				TimeSpent:   "30m",
				Status:      "Completed",
			},
		},
	}

	got := Text(out)

	if strings.Contains(got, "+---") || strings.Contains(got, "| # |") {
		t.Fatalf("expected text output without boxed table borders, got %q", got)
	}
	if !strings.Contains(got, "1. Improved package download") {
		t.Fatalf("expected numbered task heading, got %q", got)
	}
	if !strings.Contains(got, "Module: Npm  |  Time: 30m  |  Status: Completed") {
		t.Fatalf("expected compact meta line, got %q", got)
	}
	if !strings.Contains(got, "Check-in: 09:00  ->  Check-out: 18:00") {
		t.Fatalf("expected ASCII-safe check-in/check-out separator, got %q", got)
	}
	if !strings.Contains(got, "Total: 30m") {
		t.Fatalf("expected total time in output, got %q", got)
	}
}
