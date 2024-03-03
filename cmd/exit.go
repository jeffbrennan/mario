package cmd

import (
	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var exitCmd = &cobra.Command{
	Use:   "exit",
	Short: "exit Mario",
	Run: func(cmd *cobra.Command, args []string) {

		mario.Exit()
	},
}

func init() {
	RootCmd.AddCommand(exitCmd)
}
