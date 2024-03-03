package cmd

import (
	"fmt"

	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var SummarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "summarize pipeline runs",
	Run: func(cmd *cobra.Command, args []string) {

		nDays, _ := cmd.Flags().GetInt("days")
		name, _ := cmd.Flags().GetString("name")

		fmt.Println("Summarizing pipeline runs...")
		fmt.Println("Days:", nDays)
		fmt.Println("Name:", name)
		mario.Summarize(nDays, name)
	},
}

func init() {
	RootCmd.AddCommand(SummarizeCmd)
	SummarizeCmd.PersistentFlags().Int("days", 7, "number of days to summarize")
	SummarizeCmd.PersistentFlags().
		String("name", "", "substring of the pipeline to summarize")
}
