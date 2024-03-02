package compare

import (
	"github.com/jeffbrennan/mario/cmd"
	"github.com/jeffbrennan/mario/pkg/mario"
	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "compare the contents of two pipelines",
	Run: func(cmd *cobra.Command, args []string) {
		name1, _ := cmd.Flags().GetString("name1")
		name2, _ := cmd.Flags().GetString("name2")
		mario.Compare(name1, name2)
	},
}

func init() {
	cmd.RootCmd.AddCommand(compareCmd)
	compareCmd.PersistentFlags().
		String("name1", "", "the first pipeline to compare")
	compareCmd.PersistentFlags().
		String("name2", "", "the second pipeline to compare")
}
