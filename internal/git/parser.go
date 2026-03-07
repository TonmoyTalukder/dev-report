package git

import (
	"strconv"
	"strings"
	"time"

	"github.com/dev-report/dev-report/internal/types"
)

// Parse converts raw git log output (using the ==COMMIT== format) into Commit structs.
//
// Expected format produced by:
//
//	git log --pretty=format:"==COMMIT==%n%H%n%an%n%ai%n%s" --numstat
func Parse(raw string) ([]*types.Commit, error) {
	var commits []*types.Commit

	// Split on the commit marker — keep the marker line itself
	sections := strings.Split(raw, "==COMMIT==")

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		lines := strings.Split(section, "\n")
		if len(lines) < 4 {
			continue
		}

		hash := strings.TrimSpace(lines[0])
		author := strings.TrimSpace(lines[1])
		dateStr := strings.TrimSpace(lines[2])
		message := strings.TrimSpace(lines[3])

		if hash == "" {
			continue
		}

		ts, err := parseGitDate(dateStr)
		if err != nil {
			// Skip unparseable timestamps rather than failing the whole run
			continue
		}

		commit := &types.Commit{
			Hash:      hash,
			Author:    author,
			Timestamp: ts,
			Message:   message,
		}

		// Parse numstat lines (lines[4:])
		for _, line := range lines[4:] {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fc := parseNumstatLine(line)
			if fc != nil {
				commit.Files = append(commit.Files, *fc)
			}
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// parseGitDate parses the date format produced by git's %ai format:
// "2026-03-07 10:25:33 +0600"
func parseGitDate(s string) (time.Time, error) {
	// Try the git %ai format first
	layouts := []string{
		"2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05 +0700",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	// Fallback: try without seconds
	return time.Parse("2006-01-02 15:04 -0700", s)
}

// parseNumstatLine parses one line of --numstat output.
// Format: "10\t5\tpath/to/file"
// Binary files show: "-\t-\tpath/to/file"
func parseNumstatLine(line string) *types.FileChange {
	parts := strings.SplitN(line, "\t", 3)
	if len(parts) != 3 {
		return nil
	}
	added, _ := strconv.Atoi(parts[0])   // "-" → 0 for binaries
	deleted, _ := strconv.Atoi(parts[1]) // "-" → 0 for binaries
	path := strings.TrimSpace(parts[2])
	if path == "" {
		return nil
	}
	return &types.FileChange{
		Path:    path,
		Added:   added,
		Deleted: deleted,
	}
}
