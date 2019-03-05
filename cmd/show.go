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
	"sort"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var showCmd = &cobra.Command{
	Use:   `show {transaction id}`,
	Short: "Display transaction details.",
	Long: `Looking through log, an amount of a transaction and some other metadata is 
displayed. However, the particular impacts to accounts and budgets are hidden
for the sake of brevity. This command shows all known details of a transaction.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		root, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}
		root = path.Join(root, index.RepoName)

		var targetID envelopes.ID
		err = targetID.UnmarshalText([]byte(args[0]))
		if err != nil {
			logrus.Fatal(err)
		}

		persister := persist.FileSystem{
			Root: root,
		}

		loader := persist.DefaultLoader{
			Fetcher: persister,
		}

		var target envelopes.Transaction
		err = loader.Load(ctx, targetID, &target)
		if err != nil {
			logrus.Fatal(err)
		}

		err = prettyPrintTransaction(ctx, os.Stdout, loader, target)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}

// prettyPrintTransaction serializes a transaction into text in as pretty of a fashion as it can muster. It requires a
// `persist.Loader` in order to fetch any budget related information in its pursuit. Most notably, it must fetch the
// parent of this transaction to figure out the differences to each budget/account.
func prettyPrintTransaction(
	ctx context.Context,
	output io.Writer,
	loader persist.Loader,
	subject envelopes.Transaction) error {

	var parent envelopes.Transaction
	err := loader.Load(ctx, subject.Parent, &parent)
	if err != nil {
		return err
	}

	impacts := subject.State.Subtract(*parent.State)

	fmt.Fprintf(output, "Time:    \t%s\n", subject.Time)
	fmt.Fprintf(output, "Merchant:\t%s\n", subject.Merchant)
	fmt.Fprintf(output, "Amount:  \t%s\n", subject.Amount)
	fmt.Fprintf(output, "Parent:  \t%s\n", subject.Parent)
	fmt.Fprintf(output, "Comment: \t%s\n", subject.Comment)
	fmt.Fprintf(output, "Impacts:\n")

	fmt.Fprintf(output, "\tAccounts:\n")
	for acc, delta := range impacts.Accounts {
		fmt.Fprintf(output, "\t\t%s: %s\n", acc, delta)
	}

	flattened := flattenBudgets(impacts)

	sortedBudgetNames := make([]string, 0, len(flattened))
	for name := range flattened {
		sortedBudgetNames = append(sortedBudgetNames, name)
	}
	sort.Strings(sortedBudgetNames)

	fmt.Fprintf(output, "\tBudgets:\n")
	for _, name := range sortedBudgetNames {
		fmt.Fprintf(output, "\t\t%s: %s\n", name, flattened[name])
	}

	return nil
}

func flattenBudgets(diff envelopes.Impact) map[string]envelopes.Balance {
	retval := make(map[string]envelopes.Balance)
	var helper func(*envelopes.Budget, string, string)
	helper = func(current *envelopes.Budget, running, name string) {
		if current == nil {
			return
		}

		fullName := path.Join(running, name)

		if current.Balance != 0 {
			retval[fullName] = current.Balance
		}

		for childName, child := range current.Children {
			helper(child, fullName, childName)
		}
	}

	helper(diff.Budget, "", "")

	return retval
}
