// Copyright © 2019 Martin Strobel
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
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/marstr/envelopes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

const (
	balanceDepthFlag      = "depth"
	balanceDepthShorthand = "d"
	balanceDepthDefault   = 1
)

// balanceCmd represents the balance command
var balanceCmd = &cobra.Command{
	Use:     "balance [index]",
	Aliases: []string{"bal", "b"},
	Short:   "Scours a baronial directory (or subdirectory) for balance information.",
	Run: func(cmd *cobra.Command, args []string) {
		var timeout time.Duration
		var err error
		timeout, err = cmd.Flags().GetDuration(timeoutFlag)
		if err != nil {
			logrus.Fatal(err)
		}

		var ctx context.Context
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(context.Background(), timeout)
			defer cancel()

		} else {
			ctx = context.Background()
		}

		var targetDir string
		if len(args) > 0 {
			targetDir = args[0]
		} else {
			targetDir = "."
		}

		targetDir, err = filepath.Abs(targetDir)
		if err != nil {
			logrus.Fatal(err)
		}

		root, err := index.RootDirectory(targetDir)
		if err != nil {
			logrus.Fatal(err)
		}

		root, err = filepath.Abs(root)
		if err != nil {
			logrus.Fatal(err)
		}

		var budgetDir string
		if targetDir == root {
			budgetDir = path.Join(targetDir, index.BudgetDir)
		} else if _, err = index.BudgetName(targetDir); err == nil {
			budgetDir = targetDir
		} else {
			logrus.Info(err)
		}

		var accountsDir string
		if targetDir == root {
			accountsDir = path.Join(targetDir, index.AccountsDir)
		} else if _, err = index.AccountName(targetDir); err == nil {
			accountsDir = targetDir
		} else {
			logrus.Info(err)
		}

		if accountsDir != "" {
			accs, err := index.LoadAccounts(ctx, accountsDir)
			if err == nil {
				err = writeAccountBalances(ctx, os.Stdout, accs)
				if err != nil {
					logrus.Fatal(err)
				}
			} else {
				logrus.Error(err)
			}
		}

		if budgetDir != "" {
			bdg, err := index.LoadBudget(ctx, budgetDir)
			if err == nil {
				err = writeBudgetBalances(ctx, os.Stdout, *bdg)
				if err != nil {
					logrus.Fatal(err)
				}
			} else {
				logrus.Error(err)
			}
		}

	},
	Args: cobra.MaximumNArgs(1),
}

func init() {
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
		balanceDepthDefault,
		"How recursively deep the balance tree should be shown before being truncated.",
	)
}

func writeBudgetBalances(_ context.Context, output io.Writer, budget envelopes.Budget) (err error) {
	_, err = fmt.Fprintln(output, "Total: ", budget.RecursiveBalance())
	if err != nil {
		return
	}
	_, err = fmt.Fprintln(output, "Balance: ", budget.Balance)
	if err != nil {
		return
	}

	if len(budget.Children) > 0 {
		_, err = fmt.Fprintln(output, "Children:")
		if err != nil {
			return
		}

		childNames := make([]string, 0, len(budget.Children))
		for name := range budget.Children {
			childNames = append(childNames, name)
		}
		sort.Strings(childNames)

		for _, name := range childNames {
			child := budget.Children[name]
			_, err = fmt.Fprintf(output, "\t%s: %s\n", name, child.RecursiveBalance())
			if err != nil {
				return
			}
		}
	}
	return
}

func writeAccountBalances(_ context.Context, output io.Writer, accounts envelopes.Accounts) (err error) {
	_, err = fmt.Fprintln(output, "Accounts:")
	if err != nil {
		return
	}

	for _, name := range accounts.Names() {
		_, err = fmt.Fprintf(output, "\t%s: %v\n", name, accounts[name])
		if err != nil {
			return
		}
	}

	return
}
