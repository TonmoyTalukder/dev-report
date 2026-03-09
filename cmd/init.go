package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dev-report/dev-report/internal/config"
	"github.com/dev-report/dev-report/internal/constants"
)

var promptInput io.Reader = os.Stdin

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a dev-report.config.json in the current directory",
	Long:  `Interactively creates a dev-report.config.json with your preferred settings.`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	promptInput = os.Stdin

	fmt.Println("\n  dev-report — setup wizard")
	fmt.Println("  ─────────────────────────────────────")

	cfg := config.Defaults()

	cfg.User = prompt("Git author name (leave blank for all authors)", "")
	cfg.GitHubUsername = prompt("GitHub username (for repo and package publishing)", "")
	cfg.AIProvider = promptChoice("AI provider", constants.SupportedProviders, constants.DefaultAIProvider)

	switch cfg.AIProvider {
	case constants.ProviderGroq:
		cfg.GroqAPIKey = prompt("Groq API key (get free key at "+constants.GroqConsoleURL+")", "")
		fmt.Println("  Free models: llama-3.3-70b-versatile, llama3-8b-8192, gemma2-9b-it")
		cfg.GroqModel = prompt("Groq model", constants.DefaultGroqModel)
	case constants.ProviderGemini:
		cfg.GeminiAPIKey = prompt("Gemini API key (get free key at "+constants.GeminiConsoleURL+")", "")
	case constants.ProviderOpenRouter:
		cfg.OpenRouterKey = prompt("OpenRouter API key (get free key at "+constants.OpenRouterKeysURL+")", "")
	case constants.ProviderOllama:
		cfg.OllamaURL = prompt("Ollama URL", constants.DefaultOllamaURL)
		cfg.OllamaModel = prompt("Ollama model", constants.DefaultOllamaModel)
	}

	cfg.DefaultOutput = promptChoice("Default output format", constants.OutputFormats, constants.DefaultOutput)

	if err := config.Write(workDir, cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("\n  ✅ Config saved to dev-report.config.json\n")
	fmt.Printf("  Run: dev-report generate --hours=9h --adjust=35min\n")
	fmt.Printf("       or dev-report generate --checkin=09:00 --checkout=18:00\n\n")
	return nil
}

func prompt(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("  %s: ", label)
	}
	val := readPromptLine()
	if val == "" {
		return defaultVal
	}
	return val
}

func promptChoice(label string, choices []string, defaultVal string) string {
	fmt.Printf("  %s %v [%s]: ", label, choices, defaultVal)
	val := readPromptLine()
	for _, c := range choices {
		if val == c {
			return val
		}
	}
	return defaultVal
}

func readPromptLine() string {
	reader := bufio.NewReader(promptInput)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return ""
	}
	promptInput = reader
	return strings.TrimRight(line, "\r\n")
}
