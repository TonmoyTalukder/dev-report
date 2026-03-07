package types

import "time"

// Commit represents a single parsed git commit.
type Commit struct {
	Hash      string
	Author    string
	Timestamp time.Time
	Message   string
	Files     []FileChange
}

// TotalLines returns the sum of added + deleted lines across all changed files.
func (c *Commit) TotalLines() int {
	total := 0
	for _, f := range c.Files {
		total += f.Added + f.Deleted
	}
	return total
}

// FileChange holds stats for one changed file in a commit.
type FileChange struct {
	Path    string
	Added   int
	Deleted int
}

// TaskGroup is a set of related commits that form one report task row.
type TaskGroup struct {
	Commits   []*Commit
	Module    string
	Weight    float64
	TimeSpent time.Duration
}

// EarliestTime returns the oldest commit timestamp in the group.
func (g *TaskGroup) EarliestTime() time.Time {
	if len(g.Commits) == 0 {
		return time.Time{}
	}
	earliest := g.Commits[0].Timestamp
	for _, c := range g.Commits[1:] {
		if c.Timestamp.Before(earliest) {
			earliest = c.Timestamp
		}
	}
	return earliest
}

// LatestTime returns the newest commit timestamp in the group.
func (g *TaskGroup) LatestTime() time.Time {
	if len(g.Commits) == 0 {
		return time.Time{}
	}
	latest := g.Commits[0].Timestamp
	for _, c := range g.Commits[1:] {
		if c.Timestamp.After(latest) {
			latest = c.Timestamp
		}
	}
	return latest
}

// TotalFiles returns the total number of unique file changes across all commits.
func (g *TaskGroup) TotalFiles() int {
	seen := map[string]bool{}
	for _, c := range g.Commits {
		for _, f := range c.Files {
			seen[f.Path] = true
		}
	}
	return len(seen)
}

// TotalLines returns the total lines changed across all commits in the group.
func (g *TaskGroup) TotalLines() int {
	total := 0
	for _, c := range g.Commits {
		total += c.TotalLines()
	}
	return total
}

// Task is one final formatted row in the work report.
type Task struct {
	Number      int
	Title       string
	Module      string
	Description string
	TimeSpent   string
	Status      string
}

// ReportInput holds all parameters provided by the user.
type ReportInput struct {
	User       string
	Date       string // YYYY-MM-DD; empty = today
	CheckIn    string // HH:MM
	CheckOut   string // HH:MM
	LastN      int
	Adjust     string // e.g. "35min", "1h40m"
	AIProvider string // groq, gemini, ollama, openrouter
	Output     string // markdown, table, excel, json
	OutputFile string // path for excel/file output
	WorkDir    string // git repo directory
}

// ReportOutput is the completed report ready for formatting.
type ReportOutput struct {
	Input     *ReportInput
	Tasks     []*Task
	TotalTime string
	Date      string
	Developer string
	CheckIn   string
	CheckOut  string
	Adjusted  string
}
