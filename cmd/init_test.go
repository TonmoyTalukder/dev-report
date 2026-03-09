package cmd

import (
	"strings"
	"testing"
)

func TestPromptAllowsBlankInputWithoutShiftingNextAnswer(t *testing.T) {
	original := promptInput
	defer func() { promptInput = original }()

	promptInput = strings.NewReader("\nTonmoyTalukder\n\n")

	if got := prompt("Git author name", "default-user"); got != "default-user" {
		t.Fatalf("expected default on blank line, got %q", got)
	}
	if got := prompt("GitHub username", ""); got != "TonmoyTalukder" {
		t.Fatalf("expected second answer to remain aligned, got %q", got)
	}
	if got := prompt("Groq API key", "default-key"); got != "default-key" {
		t.Fatalf("expected third blank line to keep default, got %q", got)
	}
}

func TestPromptChoiceReturnsDefaultForBlankInput(t *testing.T) {
	original := promptInput
	defer func() { promptInput = original }()

	promptInput = strings.NewReader("\n")

	got := promptChoice("AI provider", []string{"groq", "gemini"}, "groq")
	if got != "groq" {
		t.Fatalf("expected default choice, got %q", got)
	}
}
