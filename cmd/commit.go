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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/marstr/baronial/internal/index"
)

const (
	amountFlag      = "amount"
	amountShorthand = "a"
	amountDefault   = "USD 0.00"
	amountUsage     = "The magnitude of the transaction that should be displayed in logs."
)

const (
	merchantFlag      = "merchant"
	merchantShorthand = "m"
	merchantDefault   = "Unknown"
	merchantUsage     = "The party receiving funds as part of this transaction."
)

const (
	commentFlag      = "comment"
	commentShorthand = "c"
	commentDefault   = ""
	commentUsage     = "Notes that may be helpful later when identifying this transaction."
)

const (
	timeFlag      = "time"
	timeShorthand = "t"
	timeDefault   = "<current date/time>"
	timeUsage     = "The time and date when this transaction occurred."
)

const (
	forceFlag      = "force"
	forceShorthand = "f"
	forceDefault   = false
	forceUsage     = "Ignore warnings, commit the transaction anyway."
)

var commitConfig = viper.New()

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Create a transaction with the current impacts in the index.",
	Args: func(cmd *cobra.Command, args []string) error {
		if commitConfig.IsSet(timeFlag) {
			if currentValue := commitConfig.GetString(timeFlag); currentValue == timeDefault {
				commitConfig.SetDefault(timeFlag, time.Now())
			}
		}

		if finalTimeValue := commitConfig.GetTime(timeFlag); finalTimeValue.Equal(time.Time{}) {
			return fmt.Errorf("unable to parse time from %q", commitConfig.GetString(timeFlag))
		}

		if !commitConfig.IsSet(amountFlag) {
			return errors.New(`missing flag "` + amountFlag + `"`)
		}

		return cobra.NoArgs(cmd, args)
	},
	Run: func(_ *cobra.Command, _ []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		amount, err := envelopes.ParseBalance(commitConfig.GetString(amountFlag))
		if err != nil {
			logrus.Fatal(err)
		}

		targetDir, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		accountsDir := filepath.Join(targetDir, index.AccountsDir)
		accounts, err := index.LoadAccounts(ctx, accountsDir)
		if err != nil {
			logrus.Fatal(err)
		}

		budgetDir := filepath.Join(targetDir, index.BudgetDir)
		budget, err := index.LoadBudget(ctx, budgetDir)
		if err != nil {
			logrus.Fatal(err)
		}

		budgetBal := budget.RecursiveBalance()
		var accountsBal envelopes.Balance
		for _, entry := range accounts {
			accountsBal += entry
		}

		if budgetBal != accountsBal {
			logrus.Warnf(
				"accounts (%s) and budget (%s) balance are not equal by %s.",
				accountsBal,
				budgetBal,
				accountsBal-budgetBal)

			if !commitConfig.GetBool(forceFlag) {
				shouldContinue, err := promptToContinue(
					ctx,
					"proceed despite imbalance?",
					os.Stdout,
					os.Stdin)
				if err != nil {
					logrus.Fatal(err)
				}

				if !shouldContinue {
					return
				}
			}
		}

		persister := persist.FileSystem{
			Root: filepath.Join(targetDir, index.RepoName),
		}

		writer := persist.DefaultWriter{
			Stasher: persister,
		}

		parent, err := persister.Current(ctx)
		if err != nil {
			logrus.Fatal(err)
		}

		currentTransaction := envelopes.Transaction{
			Amount:   amount,
			Merchant: commitConfig.GetString(merchantFlag),
			Comment:  commitConfig.GetString(commentFlag),
			State: &envelopes.State{
				Accounts: accounts,
				Budget:   budget,
			},
			Time:   commitConfig.GetTime(timeFlag),
			Parent: parent,
		}

		err = writer.Write(ctx, currentTransaction)
		if err != nil {
			logrus.Fatal(err)
		}

		err = persister.WriteCurrent(ctx, &currentTransaction)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if defaultAmount, err := findDefaultAmount(ctx, "."); err == nil {
		commitConfig.SetDefault(amountFlag, defaultAmount.String())
	} else {
		commitConfig.SetDefault(amountFlag, amountDefault)
	}

	commitCmd.PersistentFlags().StringP(merchantFlag, merchantShorthand, merchantDefault, merchantUsage)
	commitCmd.PersistentFlags().StringP(commentFlag, commentShorthand, commentDefault, commentUsage)
	commitCmd.PersistentFlags().StringP(timeFlag, timeShorthand, timeDefault, timeUsage)
	commitCmd.PersistentFlags().StringP(amountFlag, amountShorthand, commitConfig.GetString(amountFlag), amountUsage)
	commitCmd.PersistentFlags().BoolP(forceFlag, forceShorthand, forceDefault, forceUsage)

	err := commitConfig.BindPFlags(commitCmd.PersistentFlags())
	if err != nil {
		logrus.Fatal(err)
	}
}

func promptToContinue(ctx context.Context, message string, output io.Writer, input io.Reader) (bool, error) {
	results := make(chan bool, 1)
	errs := make(chan error, 1)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			default:
				// Intentionally Left Blank
			}

			_, err := fmt.Fprintf(output, "%s (y/N): ", message)
			if err != nil {
				errs <- err
				return
			}

			// If `ctx` expires while we're waiting for user response here, this goroutine will leak. There are a lot of
			// different ways to organize around this problem, but until there is a Reader that allows for
			// cancellation in the standard library (or something we're willing to take a dependency on) there's not
			// actually anyway to get around this leak.
			//
			// Given that this function is expected be executed in very short-lived programs, and realistically this
			// will leak one or zero times for any-given process before immediately being terminated, I'm not worried
			// about it.
			reader := bufio.NewReader(input)
			response, err := reader.ReadString('\n')
			if err != nil {
				errs <- err
				return
			}

			response = strings.TrimSpace(response)

			switch {
			case strings.EqualFold(response, "yes"):
				fallthrough
			case strings.EqualFold(response, "y"):
				results <- true
				return
			case strings.EqualFold(response, "quit"):
				fallthrough
			case strings.EqualFold(response, "q"):
				fallthrough
			case strings.EqualFold(response, ""):
				fallthrough
			case strings.EqualFold(response, "no"):
				fallthrough
			case strings.EqualFold(response, "n"):
				results <- false
				return
			default:
				// Intentionally Left Blank
				// The loop should be re-executed until an answer in an expected format is provided.
			}
		}
	}(ctx)

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case err := <-errs:
		return false, err
	case result := <-results:
		return result, nil
	}
}

func findDefaultAmount(ctx context.Context, targetDir string) (envelopes.Balance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	targetDir, err := index.RootDirectory(targetDir)
	if err != nil {
		return 0, err
	}

	accountsDir := filepath.Join(targetDir, index.AccountsDir)
	accounts, err := index.LoadAccounts(ctx, accountsDir)
	if err != nil {
		return 0, err
	}

	budgetDir := filepath.Join(targetDir, index.BudgetDir)
	budget, err := index.LoadBudget(ctx, budgetDir)
	if err != nil {
		return 0, err
	}

	updated := envelopes.State{
		Accounts: accounts,
		Budget:   budget,
	}

	persister := persist.FileSystem{
		Root: filepath.Join(targetDir, index.RepoName),
	}

	id, err := persister.Current(ctx)
	if err != nil {
		return 0, err
	}

	loader := persist.DefaultLoader{
		Fetcher: persister,
	}

	var head envelopes.Transaction
	err = loader.Load(ctx, id, &head)
	if err != nil {
		return 0, err
	}

	return findAmount(*head.State, updated), nil

}

func findAmount(original, updated envelopes.State) envelopes.Balance {
	if changed := findAccountAmount(original, updated); changed != 0 {
		return changed
	}

	return findBudgetAmount(original, updated)
}

func findAccountAmount(original, updated envelopes.State) envelopes.Balance {
	modifiedAccounts := make(envelopes.Accounts, len(original.Accounts))

	addedAccountNames := make(map[string]struct{}, len(original.Accounts))
	for name := range updated.Accounts {
		addedAccountNames[name] = struct{}{}
	}

	for name, oldBalance := range original.Accounts {
		if _, ok := addedAccountNames[name]; ok {
			// Mark this account as not a new one.
			delete(addedAccountNames, name)
		}

		if newBalance, ok := updated.Accounts[name]; ok && newBalance == oldBalance {
			// Nothing has changed
			continue
		} else if !ok {
			// An account was removed
			modifiedAccounts[name] = -1 * oldBalance
		} else {
			// An account had its balance modified
			modifiedAccounts[name] = newBalance - oldBalance
		}
	}

	// Iterate over the accounts that weren't seen in the original, and mark them as new.
	for name := range addedAccountNames {
		modifiedAccounts[name] = updated.Accounts[name]
	}

	// If there was a transfer between two accounts, we don't want to mark it as amount $0.00, but rather that magnitude
	// of the transfer. For that reason, we'll figure out the total negative and positive change of the accounts
	// involved.
	//
	// If it was a transfer between budgets, we'll count the total deposited into the receiving accounts.
	// If it was a deposit or credit, the amount positive or negative will get reflected because the opposite will
	// register as a zero.
	var positiveAccountDifferences, negativeAccountDifferences envelopes.Balance
	for _, bal := range modifiedAccounts {
		if bal > 0 {
			positiveAccountDifferences += bal
		} else {
			negativeAccountDifferences += bal
		}
	}

	if positiveAccountDifferences > 0 {
		return positiveAccountDifferences
	}

	if negativeAccountDifferences < 0 {
		return negativeAccountDifferences
	}

	return 0
}

func findBudgetAmount(original, updated envelopes.State) envelopes.Balance {
	// Normalize the budgets into a flattened shape for easier comparison, more like Accounts
	const separator = string(os.PathSeparator)
	originalBudgets := make(map[string]envelopes.Balance)
	updatedBudgets := make(map[string]envelopes.Balance)

	var treeFlattener func(map[string]envelopes.Balance, string, *envelopes.Budget)
	treeFlattener = func(discovered map[string]envelopes.Balance, currentPath string, target *envelopes.Budget) {
		discovered[currentPath] = target.Balance

		for name, subTarget := range target.Children {
			treeFlattener(discovered, currentPath+separator+name, subTarget)
		}
	}

	treeFlattener(originalBudgets, separator, original.Budget)
	treeFlattener(updatedBudgets, separator, updated.Budget)

	// Make a list of all budget names in the updated state, so that we can find the ones which were added.
	addedBudgets := make(map[string]struct{}, len(updatedBudgets))
	for name := range updatedBudgets {
		addedBudgets[name] = struct{}{}
	}

	modifiedBudgets := make(map[string]envelopes.Balance, len(originalBudgets))

	for name, oldBalance := range originalBudgets {
		if _, ok := addedBudgets[name]; ok {
			// Mark this budget as not a new one.
			delete(addedBudgets, name)
		}

		if newBalance, ok := updatedBudgets[name]; ok && newBalance == oldBalance {
			// Nothing has changed here
			continue
		} else if !ok {
			modifiedBudgets[name] = -1 * oldBalance
		} else {
			modifiedBudgets[name] = newBalance - oldBalance
		}
	}

	for name := range addedBudgets {
		modifiedBudgets[name] = updatedBudgets[name]
	}

	var positiveBudgetDifferences, negativeBudgetDifferences envelopes.Balance
	for _, bal := range modifiedBudgets {
		if bal > 0 {
			positiveBudgetDifferences += bal
		} else {
			negativeBudgetDifferences += bal
		}
	}

	if positiveBudgetDifferences > 0 {
		return positiveBudgetDifferences
	}

	if negativeBudgetDifferences < 0 {
		return negativeBudgetDifferences
	}

	return 0
}
