package git

import (
	"testing"
	"time"

	"github.com/dev-report/dev-report/internal/types"
)

func TestBuildOptionsUsesFullDayWindowForWorkingHours(t *testing.T) {
	input := &types.ReportInput{
		Date:         "2026-03-09",
		WorkingHours: "8h30m",
	}

	opts, err := BuildOptions(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	wantSince := time.Date(2026, 3, 9, 0, 0, 0, 0, time.Local)
	wantUntil := time.Date(2026, 3, 9, 23, 59, 0, 0, time.Local)
	if !opts.Since.Equal(wantSince) {
		t.Fatalf("expected since %v, got %v", wantSince, opts.Since)
	}
	if !opts.Until.Equal(wantUntil) {
		t.Fatalf("expected until %v, got %v", wantUntil, opts.Until)
	}
}
