package processor

import (
	"strings"
	"time"

	"github.com/dev-report/dev-report/internal/constants"
	"github.com/dev-report/dev-report/internal/types"
)

// stopWords are words that carry no grouping signal.
var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "in": true, "on": true,
	"for": true, "to": true, "of": true, "and": true, "or": true,
	"is": true, "it": true, "with": true, "from": true, "at": true,
	"by": true, "as": true, "be": true, "are": true, "was": true,
	"been": true, "will": true, "this": true, "that": true, "not": true,
}

type groupCandidate struct {
	group    *types.TaskGroup
	keywords map[string]bool
	files    map[string]bool
	module   string
	latest   time.Time
}

// GroupCommits groups related commits into TaskGroups.
// Grouping rules:
//  1. Same dominant module AND shared keyword(s) → merge
//  2. Single commit with no match → its own group
func GroupCommits(commits []*types.Commit, taskMode string) []*types.TaskGroup {
	if len(commits) == 0 {
		return nil
	}

	if taskMode == "" {
		taskMode = constants.DefaultTaskGranularity
	}
	if taskMode == constants.TaskGranularityGranular {
		result := make([]*types.TaskGroup, len(commits))
		for i, commit := range commits {
			result[i] = &types.TaskGroup{
				Commits: []*types.Commit{commit},
				Module:  dominantModuleForCommit(commit),
			}
		}
		return result
	}

	var groups []*groupCandidate

	for _, commit := range commits {
		module := dominantModuleForCommit(commit)
		keywords := extractKeywords(commit.Message)
		files := extractFiles(commit)

		bestIdx := -1
		bestScore := -1

		for i, g := range groups {
			score := groupingScore(commit, module, keywords, files, g)
			if score > bestScore {
				bestScore = score
				bestIdx = i
			}
		}

		if bestIdx >= 0 && bestScore >= mergeThreshold(taskMode) {
			groups[bestIdx].group.Commits = append(groups[bestIdx].group.Commits, commit)
			for k := range keywords {
				groups[bestIdx].keywords[k] = true
			}
			for path := range files {
				groups[bestIdx].files[path] = true
			}
			if commit.Timestamp.After(groups[bestIdx].latest) {
				groups[bestIdx].latest = commit.Timestamp
			}
		} else {
			g := &types.TaskGroup{
				Commits: []*types.Commit{commit},
				Module:  module,
			}
			groups = append(groups, &groupCandidate{
				group:    g,
				keywords: keywords,
				files:    files,
				module:   module,
				latest:   commit.Timestamp,
			})
		}
	}

	result := make([]*types.TaskGroup, len(groups))
	for i, c := range groups {
		result[i] = c.group
	}
	return result
}

func mergeThreshold(taskMode string) int {
	if taskMode == constants.TaskGranularityDetailed {
		return 5
	}
	return 4
}

func groupingScore(commit *types.Commit, module string, keywords map[string]bool, files map[string]bool, candidate *groupCandidate) int {
	if candidate.module != module {
		return -1
	}

	score := keywordOverlap(keywords, candidate.keywords) * 2
	score += keywordOverlap(files, candidate.files) * 3

	if !candidate.latest.IsZero() && !commit.Timestamp.IsZero() {
		gap := commit.Timestamp.Sub(candidate.latest)
		if gap < 0 {
			gap = -gap
		}
		if gap <= 2*time.Hour {
			score += 2
		} else if gap <= 6*time.Hour {
			score++
		}
	}

	return score
}

func extractFiles(commit *types.Commit) map[string]bool {
	files := map[string]bool{}
	for _, file := range commit.Files {
		files[file.Path] = true
	}
	return files
}

// dominantModuleForCommit returns the module detected from the commit's changed files.
func dominantModuleForCommit(c *types.Commit) string {
	paths := make([]string, len(c.Files))
	for i, f := range c.Files {
		paths[i] = f.Path
	}
	return DominantModule(paths)
}

// extractKeywords returns a set of meaningful words from a commit message.
func extractKeywords(message string) map[string]bool {
	// Strip git prefixes
	msg := strings.TrimSpace(strings.ToLower(message))
	for _, kind := range constants.GitCommitTypes {
		if strings.HasPrefix(msg, kind+":") {
			msg = strings.TrimSpace(msg[len(kind)+1:])
			break
		}
		if strings.HasPrefix(msg, kind+"(") {
			if idx := strings.Index(msg, "):"); idx >= 0 {
				msg = strings.TrimSpace(msg[idx+2:])
				break
			}
		}
	}

	// Tokenize
	words := strings.FieldsFunc(msg, func(r rune) bool {
		return r == ' ' || r == ',' || r == '.' || r == ':' ||
			r == '(' || r == ')' || r == '[' || r == ']' ||
			r == '\'' || r == '"' || r == '/' || r == '_' || r == '-'
	})

	keywords := map[string]bool{}
	for _, w := range words {
		if len(w) < 3 {
			continue
		}
		if stopWords[w] {
			continue
		}
		keywords[w] = true
	}
	return keywords
}

// keywordOverlap counts how many keywords two sets share.
func keywordOverlap(a, b map[string]bool) int {
	count := 0
	for k := range a {
		if b[k] {
			count++
		}
	}
	return count
}
