package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "summarize factory info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pick a subcommand")
	},
}

func init() {
	RootCmd.AddCommand(summarizeCmd)
}
