package mario

import (
	"fmt"
	"strconv"

	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "summarize pipeline runs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		nDays, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("invalid number of days:", args[0])
			return
		}

		mario.Summarize(nDays)
	},
}

func init() {
	rootCmd.AddCommand(summarizeCmd)
}
