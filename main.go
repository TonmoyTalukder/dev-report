package main

import (
	"os"

	"github.com/dev-report/dev-report/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
