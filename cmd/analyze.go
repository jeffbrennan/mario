package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "analyze a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pick a subcommand")
	},
}

func init() {
	RootCmd.AddCommand(analyzeCmd)
}
