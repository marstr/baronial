package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config {command}",
	Aliases: []string{"conf"},
	Short:   "Control the behavior of baronial.",
}

var accountConfigCmd = &cobra.Command{
	Use:   "account [name]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Get or set the current default account for relevant operations.",
	Long: `Certain baronial operations take an account to execute on (i.e. credit and 
debit.) Because it is common to process all transactions in a particular
account before moving to the next, instead of specifying an account every time,
one can use this command to set the account that is currently being processed.

Providing no argument will get the currently targeted account.
Providing an argument will set the currently targeted account.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	configCmd.AddCommand(accountConfigCmd)
	rootCmd.AddCommand(configCmd)

}
