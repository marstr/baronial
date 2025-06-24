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
	"github.com/marstr/envelopes/persist/filesystem"
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
	postedTimeFlag      = "posted-time"
	postedTimeShorthand = "p"
	postedTimeDefault   = "<current date/time>"
	postedTimeUsage     = "The time and date when the transaction was posted by your financial institution."
)

const (
	actualTimeFlag      = "actual-time"
	actualTimeShorthand = "t"
	actualTimeDefault   = "<empty>"
	actualTimeUsage     = "The time and date when this transaction occurred."
)

const (
	forceFlag      = "force"
	forceShorthand = "f"
	forceDefault   = false
	forceUsage     = "Ignore warnings, commit the transaction anyway."
)

const (
	bankRecordIDFlag      = "bank-record-id"
	bankRecordIDShorthand = "b"
	bankRecordIDDefault   = ""
	bankRecordIDUsage     = "A unique ID assigned to this transaction by a financial institution."
)

var commitConfig = viper.New()

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Create a transaction with the current impacts in the index.",
	Args: func(cmd *cobra.Command, args []string) error {
		if commitConfig.IsSet(postedTimeFlag) {
			if currentValue := commitConfig.GetString(postedTimeFlag); currentValue == postedTimeDefault {
				commitConfig.SetDefault(postedTimeFlag, time.Now())
			}
		}
		if finalTimeValue := commitConfig.GetTime(postedTimeFlag); finalTimeValue.Equal(time.Time{}) {
			return fmt.Errorf("unable to parse time from %q", commitConfig.GetString(postedTimeFlag))
		}

		if commitConfig.IsSet(actualTimeFlag) {
			if currentValue := commitConfig.GetString(actualTimeFlag); currentValue == actualTimeDefault {
				commitConfig.SetDefault(actualTimeFlag, "")
			}
		}
		if finalTimeValue := commitConfig.GetTime(actualTimeFlag); commitConfig.GetString(actualTimeFlag) != "" && finalTimeValue.Equal(time.Time{}) {
			return fmt.Errorf("unable to parse time from %q", commitConfig.GetString(actualTimeFlag))
		}

		if !commitConfig.IsSet(amountFlag) {
			return errors.New(`missing flag "` + amountFlag + `"`)
		}

		return cobra.NoArgs(cmd, args)
	},
	Run: func(_ *cobra.Command, _ []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		amount, err := envelopes.ParseBalance([]byte(commitConfig.GetString(amountFlag)))
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
			accountsBal = accountsBal.Add(entry)
		}

		if !budgetBal.Equal(accountsBal) {
			logrus.Warnf(
				"accounts (%s) and budget (%s) balance are not equal by %s.",
				accountsBal,
				budgetBal,
				accountsBal.Sub(budgetBal))

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

		var repo persist.RepositoryReaderWriter
		repo, err = filesystem.OpenRepositoryWithCache(ctx, filepath.Join(targetDir, index.RepoName), 10000)
		if err != nil {
			logrus.Fatal(err)
		}

		currentTransaction := envelopes.Transaction{
			Amount:   amount,
			Merchant: commitConfig.GetString(merchantFlag),
			Comment:  commitConfig.GetString(commentFlag),
			RecordID: envelopes.BankRecordID(commitConfig.GetString(bankRecordIDFlag)),
			State: &envelopes.State{
				Accounts: accounts,
				Budget:   budget,
			},
			ActualTime:  commitConfig.GetTime(actualTimeFlag),
			PostedTime:  commitConfig.GetTime(postedTimeFlag),
			EnteredTime: time.Now(),
		}

		err = persist.Commit(ctx, repo, currentTransaction)
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
	commitCmd.PersistentFlags().StringP(postedTimeFlag, postedTimeShorthand, postedTimeDefault, postedTimeUsage)
	commitCmd.PersistentFlags().StringP(actualTimeFlag, actualTimeShorthand, actualTimeDefault, actualTimeUsage)
	commitCmd.PersistentFlags().StringP(amountFlag, amountShorthand, commitConfig.GetString(amountFlag), amountUsage)
	commitCmd.PersistentFlags().StringP(bankRecordIDFlag, bankRecordIDShorthand, bankRecordIDDefault, bankRecordIDUsage)
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
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	targetDir, err := index.RootDirectory(targetDir)
	if err != nil {
		return envelopes.Balance{}, err
	}

	accountsDir := filepath.Join(targetDir, index.AccountsDir)
	accounts, err := index.LoadAccounts(ctx, accountsDir)
	if err != nil {
		return envelopes.Balance{}, err
	}

	budgetDir := filepath.Join(targetDir, index.BudgetDir)
	budget, err := index.LoadBudget(ctx, budgetDir)
	if err != nil {
		return envelopes.Balance{}, err
	}

	updated := envelopes.State{
		Accounts: accounts,
		Budget:   budget,
	}

	var repo persist.RepositoryReader
	repo, err = filesystem.OpenRepositoryWithCache(ctx, filepath.Join(targetDir, index.RepoName), 10000)

	current, err := repo.Current(ctx)
	if err != nil {
		return envelopes.Balance{}, err
	}

	id, err := persist.Resolve(ctx, repo, current)
	if err != nil {
		return envelopes.Balance{}, err
	}

	var head envelopes.Transaction
	err = repo.LoadTransaction(ctx, id, &head)
	if err != nil {
		return envelopes.Balance{}, err
	}

	return envelopes.CalculateAmount(*head.State, updated), nil
}
