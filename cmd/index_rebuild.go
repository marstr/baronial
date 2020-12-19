/*
Copyright © 2020 Martin Strobel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	deleteFirstFlag = "unsafe-pre-delete"
	deleteFirstUsage = `Forgo the safety measure of rebuilding each index
in a new file before overwriting the old copy.`
)

var rebuildConfig = viper.New()

// rebuildCmd represents the rebuild command
var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Delete and reconstruct indices.",
	Long: `Each index is a collection of files. These files list the transactions
associated with some aspect of your transaction data. Like the index in a book,
they offer to show you which page to turn to, not the actual data. In this way,
they are fundamentally a redundant structure, and do not intrinsically store
your data.

This operation will re-traverse every transaction present in this repository,
and rebuild from scratch each list of transactions. Once the operation is
complete, the original index will be deleted. Depending on the number of
transactions in your repository, this could be a time consuming endeavor.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("rebuild called")
	},
}

func init() {
	indexCmd.AddCommand(rebuildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rebuildCmd.PersistentFlags().String("foo", "", "A help for foo")

	rebuildCmd.PersistentFlags().Bool(deleteFirstFlag, false, deleteFirstUsage)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rebuildCmd.Flags().BoolP("toggle", "t", false, "Help messasge for toggle")

	err := rebuildConfig.BindPFlags(rebuildCmd.PersistentFlags())
	if err != nil {
		logrus.Fatal(err)
	}
}
