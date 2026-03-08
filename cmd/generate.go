package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dev-report/dev-report/internal/config"
	"github.com/dev-report/dev-report/internal/constants"
	"github.com/dev-report/dev-report/internal/engine"
	"github.com/dev-report/dev-report/internal/report"
	"github.com/dev-report/dev-report/internal/types"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a work report from Git commits",
	Example: `  # Time range for today
  dev-report generate --user=john --checkin=09:00 --checkout=18:00

  # With break adjustment
  dev-report generate --user=john --checkin=09:00 --checkout=18:00 --adjust=35min

  # Specific date
  dev-report generate --user=john --date=2026-03-07 --checkin=09:00 --checkout=17:30

  # Last N commits
  dev-report generate --user=john --last=10

  # Use a specific AI provider
  dev-report generate --user=john --checkin=09:00 --checkout=18:00 --ai=gemini

  # Export to Excel
  dev-report generate --user=john --checkin=09:00 --checkout=18:00 --output=excel`,
	RunE: runGenerate,
}

// flags
var (
	flagUser       string
	flagDate       string
	flagCheckIn    string
	flagCheckOut   string
	flagLast       int
	flagAdjust     string
	flagTaskMode   string
	flagAI         string
	flagOutput     string
	flagOutputFile string
)

func init() {
	generateCmd.Flags().StringVar(&flagUser, "user", "", "Git author name to filter commits (leave empty for all authors)")
	generateCmd.Flags().StringVar(&flagDate, "date", "", "Date to generate report for, YYYY-MM-DD (default: today)")
	generateCmd.Flags().StringVar(&flagCheckIn, "checkin", "", "Work start time, HH:MM (e.g. 09:00)")
	generateCmd.Flags().StringVar(&flagCheckOut, "checkout", "", "Work end time, HH:MM (e.g. 18:00)")
	generateCmd.Flags().IntVar(&flagLast, "last", 0, "Use last N commits instead of date/time filter")
	generateCmd.Flags().StringVar(&flagAdjust, "adjust", "", "Non-task time to subtract from budget (e.g. 35min, 1h40m)")
	generateCmd.Flags().StringVar(&flagTaskMode, "task-mode", constants.DefaultTaskGranularity, "Task granularity: "+strings.Join(constants.TaskGranularities, ", "))
	generateCmd.Flags().StringVar(&flagAI, "ai", "", "AI provider: "+strings.Join(constants.SupportedProviders, ", ")+" (overrides config)")
	generateCmd.Flags().StringVar(&flagOutput, "output", "", "Output format: table (default), markdown, excel, json")
	generateCmd.Flags().StringVar(&flagOutputFile, "out", "", "Output file path (for markdown/excel/json)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	// Load config (file + env vars)
	cfg, err := config.Load(workDir)
	if err != nil {
		return fmt.Errorf("config load: %w", err)
	}

	// Build input, CLI flags override config
	user := flagUser
	if user == "" {
		user = cfg.User
	}

	outputFmt := flagOutput
	if outputFmt == "" {
		outputFmt = cfg.DefaultOutput
	}
	if outputFmt == "" {
		outputFmt = constants.DefaultOutput
	}

	taskMode := flagTaskMode
	if taskMode == "" {
		taskMode = constants.DefaultTaskGranularity
	}

	input := &types.ReportInput{
		User:        user,
		Date:        flagDate,
		CheckIn:     flagCheckIn,
		CheckOut:    flagCheckOut,
		LastN:       flagLast,
		Adjust:      flagAdjust,
		TaskMode:    taskMode,
		AIProvider:  flagAI,
		Output:      outputFmt,
		OutputFile:  flagOutputFile,
		ProjectName: filepath.Base(workDir),
		WorkDir:     workDir,
	}

	if (input.CheckIn == "") != (input.CheckOut == "") {
		return fmt.Errorf("both --checkin and --checkout must be provided together")
	}
	if input.LastN < 0 {
		return fmt.Errorf("--last must be zero or greater")
	}
	if !isSupportedOutput(input.Output) {
		return fmt.Errorf("unsupported --output %q (supported: %s)", input.Output, strings.Join(constants.OutputFormats, ", "))
	}
	if !isSupportedTaskMode(input.TaskMode) {
		return fmt.Errorf("unsupported --task-mode %q (supported: %s)", input.TaskMode, strings.Join(constants.TaskGranularities, ", "))
	}
	if input.AIProvider != "" && !isSupportedProvider(input.AIProvider) {
		return fmt.Errorf("unsupported --ai %q (supported: %s)", input.AIProvider, strings.Join(constants.SupportedProviders, ", "))
	}

	// Validate: need at least one commit selection method
	if input.User == "" && input.LastN == 0 && input.Date == "" && input.CheckIn == "" {
		fmt.Fprintln(os.Stderr, "ℹ  No filter specified — fetching all commits from today.")
		input.Date = time.Now().Format("2006-01-02")
	}

	fmt.Fprintf(os.Stderr, "\n🔍 %s — generating report…\n", constants.AppName)
	if input.CheckIn != "" && input.CheckOut != "" {
		fmt.Fprintf(os.Stderr, "   %s %s -> %s", input.Date, input.CheckIn, input.CheckOut)
		if input.Adjust != "" {
			fmt.Fprintf(os.Stderr, "  (adjusted -%s)", input.Adjust)
		}
		fmt.Fprintln(os.Stderr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	out, err := engine.Run(ctx, input, cfg)
	if err != nil {
		return err
	}

	// Output
	switch outputFmt {
	case constants.OutputMarkdown:
		md := report.Markdown(out)
		if flagOutputFile != "" {
			if err := os.WriteFile(flagOutputFile, []byte(md), 0644); err != nil {
				return fmt.Errorf("write markdown file: %w", err)
			}
			fmt.Fprintf(os.Stderr, "✅ Markdown saved to %s\n", flagOutputFile)
		} else {
			fmt.Println(md)
		}

	case constants.OutputExcel:
		outFile := flagOutputFile
		if outFile == "" {
			date := out.Date
			if date == "" {
				date = time.Now().Format("2006-01-02")
			}
			outFile = fmt.Sprintf("work_report_%s.xlsx", date)
		}
		if err := report.Excel(out, outFile); err != nil {
			return fmt.Errorf("excel export: %w", err)
		}
		abs, _ := filepath.Abs(outFile)
		fmt.Fprintf(os.Stderr, "✅ Excel saved to %s\n", abs)

	case constants.OutputJSON:
		printJSON(out, flagOutputFile)

	default: // table
		report.Table(out)

		// If user also wants a file output
		if flagOutputFile != "" {
			ext := filepath.Ext(flagOutputFile)
			switch ext {
			case ".md", ".markdown":
				md := report.Markdown(out)
				_ = os.WriteFile(flagOutputFile, []byte(md), 0644)
				fmt.Fprintf(os.Stderr, "✅ Markdown saved to %s\n", flagOutputFile)
			case ".xlsx":
				_ = report.Excel(out, flagOutputFile)
				fmt.Fprintf(os.Stderr, "✅ Excel saved to %s\n", flagOutputFile)
			}
		}
	}

	return nil
}

// printJSON outputs the report as JSON to stdout or a file.
func printJSON(out *types.ReportOutput, path string) {
	type jsonTask struct {
		Number      int    `json:"number"`
		Task        string `json:"task"`
		Project     string `json:"project"`
		Description string `json:"description"`
		TimeSpent   string `json:"timeSpent"`
		Status      string `json:"status"`
	}
	tasks := make([]jsonTask, len(out.Tasks))
	for i, t := range out.Tasks {
		tasks[i] = jsonTask{
			Number:      t.Number,
			Task:        t.Title,
			Project:     t.Project,
			Description: t.Description,
			TimeSpent:   t.TimeSpent,
			Status:      t.Status,
		}
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal error: %v\n", err)
		return
	}

	if path != "" {
		if err := os.WriteFile(path, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write file error: %v\n", err)
			return
		}
		fmt.Fprintf(os.Stderr, "✅ JSON saved to %s\n", path)
	} else {
		fmt.Println(string(data))
	}
}

func isSupportedProvider(provider string) bool {
	for _, candidate := range constants.SupportedProviders {
		if provider == candidate {
			return true
		}
	}
	return false
}

func isSupportedOutput(output string) bool {
	for _, candidate := range constants.OutputFormats {
		if output == candidate {
			return true
		}
	}
	return false
}

func isSupportedTaskMode(taskMode string) bool {
	for _, candidate := range constants.TaskGranularities {
		if taskMode == candidate {
			return true
		}
	}
	return false
}
