package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dev-report/dev-report/internal/constants"
)

var rootCmd = &cobra.Command{
	Use:   "dev-report",
	Short: "Generate a daily work report from your Git commits",
	Long: `dev-report — AI-powered developer work report generator.

Reads your Git commits and produces a structured work report with
	task names, projects, descriptions, and time spent.

Supported AI providers (all free): ` + strings.Join(constants.SupportedProviders, ", ") + `

Examples:
  dev-report generate --user=john --hours=9h --adjust=35min
  dev-report generate --user=john --date=2026-03-07 --checkin=09:00 --checkout=18:00
  dev-report generate --user=john --last=10 --ai=gemini
  dev-report init`,
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}
