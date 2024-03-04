package cmd

import (
	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var summarizeRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "summarize pipeline runs",
	Run: func(cmd *cobra.Command, args []string) {
		nDays, _ := cmd.Flags().GetInt("days")
		name, _ := cmd.Flags().GetString("name")
		mario.Summarize(nDays, name)
	},
}

func init() {
	summarizeCmd.AddCommand(summarizeRunsCmd)
	summarizeRunsCmd.PersistentFlags().Int("days", 7, "number of days to summarize")
	summarizeRunsCmd.PersistentFlags().
		String("name", "", "substring of the pipeline to summarize")
}
