package summarize

import (
	"github.com/jeffbrennan/mario/cmd"
	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "summarize pipeline runs",
	Run: func(cmd *cobra.Command, args []string) {

		nDays, _ := cmd.Flags().GetInt("days")
		name, _ := cmd.Flags().GetString("name")
		mario.Summarize(nDays, name)
	},
}

func init() {
	cmd.RootCmd.AddCommand(summarizeCmd)
	summarizeCmd.PersistentFlags().Int("days", 7, "number of days to summarize")
	summarizeCmd.PersistentFlags().
		String("name", "", "substring of the pipeline to summarize")
}
