package cmd

import (
	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var summarizePipelinesCmd = &cobra.Command{
	Use:   "pipelines",
	Short: "summarize pipeline information",
	Run: func(cmd *cobra.Command, args []string) {
		mario.SummarizePipelines()
	},
}

func init() {
	summarizeCmd.AddCommand(summarizePipelinesCmd)
}
