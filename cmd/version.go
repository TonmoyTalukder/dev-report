package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is injected at build time via ldflags:
//
//	go build -ldflags "-X github.com/dev-report/dev-report/cmd.Version=1.0.0"
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the dev-report version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dev-report %s\n", Version)
	},
}
