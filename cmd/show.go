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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		root, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}
		root = path.Join(root, index.RepoName)

		persister := persist.FileSystem{
			Root: root,
		}

		loader := persist.DefaultLoader{
			Fetcher: persister,
		}

		resolver := persist.RefSpecResolver{
			Loader: loader,
			Brancher: persister,
			CurrentReader: persister,
		}

		var targetID envelopes.ID

		targetID, err = resolver.Resolve(ctx, persist.RefSpec(args[0]))
		if err != nil {
			logrus.Fatal(err)
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
	var err error
	var impacts envelopes.Impact

	if subject.Parent.Equal(envelopes.ID{}) {
		impacts = envelopes.Impact(*subject.State)
	} else {
		var parent envelopes.Transaction
		err := loader.Load(ctx, subject.Parent, &parent)
		if err != nil {
			return err
		}
		impacts = subject.State.Subtract(*parent.State)
	}

	if !subject.ActualTime.Equal(time.Time{}) {
		_, err = fmt.Fprintf(output, "Actual Time:    \t%s\n", subject.ActualTime)
		if err != nil {
			return err
		}
	}
	if !subject.PostedTime.Equal(time.Time{}) {
		_, err = fmt.Fprintf(output, "Posted Time:    \t%s\n", subject.PostedTime)
		if err != nil {
			return err
		}
	}
	if !subject.EnteredTime.Equal(time.Time{}) {
		_, err = fmt.Fprintf(output, "Entered Time:    \t%s\n", subject.EnteredTime)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(output, "Merchant:\t%s\n", subject.Merchant)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Amount:  \t%s\n", subject.Amount)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Parent:  \t%s\n", subject.Parent)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Comment: \t%s\n", subject.Comment)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Impacts:\n")
	if err != nil {
		return err
	}

	return prettyPrintImpact(output, impacts)
}

func prettyPrintImpact(output io.Writer, impacts envelopes.Impact) (err error) {
	_ , err = fmt.Fprintf(output, "\tAccounts:\n")
	if err != nil {
		return
	}
	for acc, delta := range impacts.Accounts {
		_ , err = fmt.Fprintf(output, "\t\t%s: %s\n", acc, delta)
		if err != nil {
			return
		}
	}

	flattened := flattenBudgets(impacts)

	sortedBudgetNames := make([]string, 0, len(flattened))
	for name := range flattened {
		sortedBudgetNames = append(sortedBudgetNames, name)
	}
	sort.Strings(sortedBudgetNames)

	_ , err = fmt.Fprintf(output, "\tBudgets:\n")
	for _, name := range sortedBudgetNames {
		_ , err = fmt.Fprintf(output, "\t\t%s: %s\n", name, flattened[name])
		if err != nil {
			return
		}
	}

	return
}

func flattenBudgets(diff envelopes.Impact) map[string]envelopes.Balance {
	retval := make(map[string]envelopes.Balance)
	var helper func(*envelopes.Budget, string, string)
	helper = func(current *envelopes.Budget, running, name string) {
		if current == nil {
			return
		}

		fullName := path.Join(running, name)

		if !current.Balance.Equal(envelopes.Balance{}) {
			retval[fullName] = current.Balance
		}

		for childName, child := range current.Children {
			helper(child, fullName, childName)
		}
	}

	helper(diff.Budget, "", "")

	return retval
}
