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
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	balanceDepthFlag      = "depth"
	balanceDepthShorthand = "d"
	balanceDepthDefault   = 1
)

var balanceConfig = viper.New()

// balanceCmd represents the balance command
var balanceCmd = &cobra.Command{
	Use:     "balance [index]",
	Aliases: []string{"bal", "b"},
	Short:   "Scours a baronial directory (or subdirectory) for balance information.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var targetDir string
		if len(args) > 0 {
			targetDir = args[0]
		} else {
			targetDir = "."
		}

		bdg, err := index.Load(ctx, targetDir)
		if err != nil {
			logrus.Fatal(err)
		}

		writeBalances(ctx, os.Stdout, bdg)
	},
	Args: cobra.MaximumNArgs(1),
}

func init() {
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

func writeBalances(_ context.Context, output io.Writer, budget envelopes.Budget) (err error) {
	fmt.Fprintln(output, "Total: ", envelopes.FormatAmount(budget.RecursiveBalance()))
	fmt.Fprintln(output, "Balance: ", envelopes.FormatAmount(budget.Balance()))

	children := budget.Children()
	if len(children) > 0 {
		fmt.Fprintln(output, "Children:")
		for name, child := range budget.Children() {
			fmt.Fprintf(output, "\t%s: %s\n", name, envelopes.FormatAmount(child.RecursiveBalance()))
		}
	}
	return nil
}
