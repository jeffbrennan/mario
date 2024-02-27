package auth

import (
	"fmt"

	cmd "github.com/jeffbrennan/mario/cmd"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "add azure authentication to the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: add azure authentication")
	},
}

func init() {
	cmd.RootCmd.AddCommand(authCmd)
}
