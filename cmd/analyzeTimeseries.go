package cmd

import (
	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var analyzeTimeseriesCmd = &cobra.Command{
	Use:   "timeseries",
	Short: "print a timeseries of runs for a pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		nDays, _ := cmd.Flags().GetInt("days")
		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			panic("name is required")
		}

		mario.AnalyzeRuns(nDays, name)
	},
}

func init() {
	analyzeCmd.AddCommand(analyzeTimeseriesCmd)
	analyzeTimeseriesCmd.PersistentFlags().
		Int("days", 7, "number of days to analyze")
	analyzeTimeseriesCmd.PersistentFlags().
		String("name", "", "name of pipeline")
}
