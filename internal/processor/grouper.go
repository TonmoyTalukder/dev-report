package processor

import (
	"strings"

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

// GroupCommits groups related commits into TaskGroups.
// Grouping rules:
//  1. Same dominant module AND shared keyword(s) → merge
//  2. Single commit with no match → its own group
func GroupCommits(commits []*types.Commit) []*types.TaskGroup {
	if len(commits) == 0 {
		return nil
	}

	type candidate struct {
		group    *types.TaskGroup
		keywords map[string]bool
		module   string
	}

	var groups []*candidate

	for _, commit := range commits {
		module := dominantModuleForCommit(commit)
		keywords := extractKeywords(commit.Message)

		// Try to find an existing group to merge into
		bestIdx := -1
		bestScore := 0

		for i, g := range groups {
			if g.module != module {
				continue
			}
			score := keywordOverlap(keywords, g.keywords)
			if score > bestScore {
				bestScore = score
				bestIdx = i
			}
		}

		if bestIdx >= 0 && bestScore >= 1 {
			// Merge into existing group
			groups[bestIdx].group.Commits = append(groups[bestIdx].group.Commits, commit)
			// Merge keywords
			for k := range keywords {
				groups[bestIdx].keywords[k] = true
			}
		} else {
			// Start a new group
			g := &types.TaskGroup{
				Commits: []*types.Commit{commit},
				Module:  module,
			}
			groups = append(groups, &candidate{
				group:    g,
				keywords: keywords,
				module:   module,
			})
		}
	}

	result := make([]*types.TaskGroup, len(groups))
	for i, c := range groups {
		result[i] = c.group
	}
	return result
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
