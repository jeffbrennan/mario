package config

import (
	"fmt"

	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "add azure environment details to the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("This is the config setup command")
		mario.ConfigSetup()
	},
}

func init() {
	configCmd.AddCommand(configSetupCmd)
}
