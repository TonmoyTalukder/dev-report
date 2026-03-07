package processor

import (
	"path/filepath"
	"strings"
	"unicode"
)

// DetectModule returns a human-readable module name from a file path.
// It uses the top-level directory name, title-cased.
// e.g. "hospital/doctorSummary.tsx" → "Hospital"
//
//	"src/components/Button.tsx"  → "Components"
//	"main.go"                    → "Root"
func DetectModule(filePath string) string {
	filePath = filepath.ToSlash(filePath)
	parts := strings.Split(filePath, "/")

	// Skip common non-meaningful top-level dirs to get to the real module
	skip := map[string]bool{
		"src": true, "lib": true, "app": true, "pkg": true,
		"internal": true, "cmd": true, "api": true,
	}

	for _, part := range parts[:max(len(parts)-1, 1)] {
		if part == "" {
			continue
		}
		if skip[strings.ToLower(part)] {
			continue
		}
		return titleCase(part)
	}

	// Only a filename at root level
	return "Root"
}

// DominantModule picks the most common module across a list of file paths.
func DominantModule(files []string) string {
	if len(files) == 0 {
		return "General"
	}
	counts := map[string]int{}
	for _, f := range files {
		m := DetectModule(f)
		counts[m]++
	}
	best := ""
	bestCount := 0
	for m, c := range counts {
		if c > bestCount || (c == bestCount && m < best) {
			best = m
			bestCount = c
		}
	}
	if best == "" {
		return "General"
	}
	return best
}

// titleCase converts a string like "doctorSummary" or "doctor_summary" or
// "doctor-summary" to "Doctor Summary".
func titleCase(s string) string {
	// Replace separators
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Split camelCase
	s = splitCamelCase(s)

	// Title-case each word
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// splitCamelCase inserts spaces before uppercase letters in camelCase strings.
func splitCamelCase(s string) string {
	var result strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(runes[i-1]) {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	return result.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
