package git

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/dev-report/dev-report/internal/types"
)

// FetchOptions controls what commits git log will return.
type FetchOptions struct {
	Author  string
	Since   time.Time
	Until   time.Time
	LastN   int
	WorkDir string
}

// Fetch runs git log with the given options and returns parsed commits.
func Fetch(opts FetchOptions) ([]*types.Commit, error) {
	args := []string{
		"log",
		"--pretty=format:==COMMIT==%n%H%n%an%n%ai%n%s",
		"--numstat",
		"--diff-filter=ACMRT",
	}

	if opts.Author != "" {
		args = append(args, fmt.Sprintf("--author=%s", opts.Author))
	}

	if !opts.Since.IsZero() {
		args = append(args, fmt.Sprintf("--since=%s", opts.Since.Format(time.RFC3339)))
	}

	if !opts.Until.IsZero() {
		args = append(args, fmt.Sprintf("--until=%s", opts.Until.Format(time.RFC3339)))
	}

	if opts.LastN > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", opts.LastN))
	}

	cmd := exec.Command("git", args...)
	if opts.WorkDir != "" {
		cmd.Dir = opts.WorkDir
	}

	out, err := cmd.Output()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			return nil, fmt.Errorf("git log failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return Parse(string(out))
}

// BuildOptions constructs FetchOptions from a ReportInput.
// It resolves the date and time range to concrete Since/Until values.
func BuildOptions(input *types.ReportInput) (FetchOptions, error) {
	opts := FetchOptions{
		Author:  input.User,
		LastN:   input.LastN,
		WorkDir: input.WorkDir,
	}

	// Determine the reference date
	refDate := input.Date
	if refDate == "" {
		refDate = time.Now().Format("2006-01-02")
	}

	// If checkin/checkout provided, build time range
	if input.CheckIn != "" && input.CheckOut != "" {
		since, err := parseDateTime(refDate, input.CheckIn)
		if err != nil {
			return opts, fmt.Errorf("invalid check-in time: %w", err)
		}
		until, err := parseDateTime(refDate, input.CheckOut)
		if err != nil {
			return opts, fmt.Errorf("invalid check-out time: %w", err)
		}
		opts.Since = since
		opts.Until = until
	} else if input.Date != "" {
		// Date only: full day
		since, err := parseDateTime(refDate, "00:00")
		if err != nil {
			return opts, err
		}
		until, err := parseDateTime(refDate, "23:59")
		if err != nil {
			return opts, err
		}
		opts.Since = since
		opts.Until = until
	}
	// If only LastN is set, Since/Until stay zero (no time filter)

	return opts, nil
}

// parseDateTime combines a date string (YYYY-MM-DD) and time string (HH:MM)
// into a time.Time using the local timezone.
func parseDateTime(date, t string) (time.Time, error) {
	combined := fmt.Sprintf("%s %s", date, t)
	return time.ParseInLocation("2006-01-02 15:04", combined, time.Local)
}
