package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "mario",
	Short: "Mario - an ADF monitoring tool",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Whoops. There was an error while executing your CLI '%s'",
			err,
		)
		os.Exit(1)
	}
}
