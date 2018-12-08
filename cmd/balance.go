// Copyright © 2018 Martin Strobel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	balanceDepthFlag      = "depth"
	balanceDepthShorthand = "d"
	balanceDepthDefault   = 1
)

var balanceConfig *viper.Viper

// balanceCmd represents the balance command
var balanceCmd = &cobra.Command{
	Use:     "balance [budget]",
	Aliases: []string{"bal", "b"},
	Short:   "Scours a ledger directory (or subdirectory) for balance information.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("balance called")
	},
	Args: cobra.MaximumNArgs(1),
}

func init() {
	balanceConfig = viper.New()
	balanceConfig.SetDefault(balanceDepthFlag, balanceDepthDefault)

	rootCmd.AddCommand(balanceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// balanceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// balanceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	balanceCmd.Flags().Uint8P(
		balanceDepthFlag,
		balanceDepthShorthand,
		uint8(balanceConfig.GetInt(balanceDepthFlag)),
		`How recursively deep the balance tree should be shown before being truncated.`)
}
