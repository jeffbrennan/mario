package cmd

import (
	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "add azure environment details to the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		mario.HelloConfig()
	},
}

func init() {
	RootCmd.AddCommand(configCmd)

}
