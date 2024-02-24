package mario

import (
	"fmt"

	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "summarize pipeline runs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		res := mario.Summarize(args[0])
		fmt.Println(res)
	},
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}
