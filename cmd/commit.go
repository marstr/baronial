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
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

const (
	amountFlag      = "amount"
	amountShorthand = "a"
	amountDefault   = "<calculated>"
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

var commitTransactionFromFlags envelopes.Transaction

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Create a transaction with the current impacts in the index.",
	Args: func(cmd *cobra.Command, args []string) error {
		currentTime := time.Now()

		if cmd.Flags().Changed(postedTimeFlag) {
			if postedTime, err := cmd.Flags().GetString(postedTimeFlag); err == nil {
				commitTransactionFromFlags.PostedTime, err = cast.ToTimeE(postedTime)
				if err != nil {
					return fmt.Errorf("unable to parse time from %q because: %v", postedTime, err)
				}
			}
		} else {
			commitTransactionFromFlags.PostedTime = currentTime
		}

		if cmd.Flags().Changed(actualTimeFlag) {
			if actualTime, err := cmd.Flags().GetString(actualTimeFlag); err == nil {
				commitTransactionFromFlags.ActualTime, err = cast.ToTimeE(actualTime)
				if err != nil {
					return fmt.Errorf("unable to parse time from %q because: %v", actualTime, err)
				}
			}
		}

		if cmd.Flags().Changed(amountFlag) {
			rawAmount, err := cmd.Flags().GetString(amountFlag)
			if err != nil {
				return err
			}
			commitTransactionFromFlags.Amount, err = envelopes.ParseBalance([]byte(rawAmount))
			if err != nil {
				return err
			}
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			var err error
			commitTransactionFromFlags.Amount, err = calculateAmount(ctx, ".")
			if err != nil {
				logrus.Fatalf("Failed to calculate the amount from %q because of the following error: %s", amountDefault, err)
			}

		}

		commitTransactionFromFlags.EnteredTime = time.Now()

		return cobra.NoArgs(cmd, args)
	},
	Run: func(cmd *cobra.Command, _ []string) {
		ctx, cancel := RootContext(cmd)
		defer cancel()

		targetDir, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		commitTransactionFromFlags.State, err = index.LoadState(ctx, targetDir)
		if err != nil {
			logrus.Fatal(err)
		}

		budgetBal := commitTransactionFromFlags.State.Budget.RecursiveBalance()
		var accountsBal envelopes.Balance
		for _, entry := range commitTransactionFromFlags.State.Accounts {
			accountsBal = accountsBal.Add(entry)
		}

		var force bool
		force, err = cmd.Flags().GetBool(forceFlag)
		if err != nil {
			logrus.Fatal(err)
		}

		if !budgetBal.Equal(accountsBal) {
			logrus.Warnf(
				"accounts (%s) and budget (%s) balance are not equal by %s.",
				accountsBal,
				budgetBal,
				accountsBal.Sub(budgetBal))

			if !force {
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

		repoLoc := filepath.Join(targetDir, index.RepoName)
		var repo persist.RepositoryReaderWriter
		repo, err = filesystem.OpenRepositoryWithCache(ctx, repoLoc, 10000)
		if err != nil {
			logrus.Fatal(err)
		}

		commitTransactionFromFlags.Merchant, err = cmd.Flags().GetString(merchantFlag)
		if err != nil {
			logrus.Fatal(err)
		}

		commitTransactionFromFlags.Comment, err = cmd.Flags().GetString(commentFlag)
		if err != nil {
			logrus.Fatal(err)
		}

		var pendingRevert bool
		pendingRevert, err = RevertIsInProgress(ctx, repoLoc)
		if err != nil {
			logrus.Warn("unable to read if revert is staged, assuming not")
		}

		if pendingRevert {
			var revertParameters RevertParameters
			err = RevertUnstowProgress(ctx, repoLoc, &revertParameters)
			if err != nil {
				logrus.Fatal("unable to read pending revert")
			}

			commitTransactionFromFlags.Reverts = revertParameters.Reverts

			if commitTransactionFromFlags.Comment == "" {
				commitTransactionFromFlags.Comment = revertParameters.Comment
			}

			defer RevertResetProgress(ctx, repoLoc)
		}

		var rawRecordId string
		rawRecordId, err = cmd.Flags().GetString(bankRecordIDFlag)
		if err != nil {
			logrus.Fatal(err)
		}
		commitTransactionFromFlags.RecordID = envelopes.BankRecordID(rawRecordId)

		err = persist.Commit(ctx, repo, commitTransactionFromFlags)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)

	commitCmd.Flags().StringP(merchantFlag, merchantShorthand, merchantDefault, merchantUsage)
	commitCmd.Flags().StringP(commentFlag, commentShorthand, commentDefault, commentUsage)
	commitCmd.Flags().StringP(postedTimeFlag, postedTimeShorthand, postedTimeDefault, postedTimeUsage)
	commitCmd.Flags().StringP(actualTimeFlag, actualTimeShorthand, actualTimeDefault, actualTimeUsage)
	commitCmd.Flags().StringP(amountFlag, amountShorthand, amountDefault, amountUsage)
	commitCmd.Flags().StringP(bankRecordIDFlag, bankRecordIDShorthand, bankRecordIDDefault, bankRecordIDUsage)
	commitCmd.Flags().BoolP(forceFlag, forceShorthand, forceDefault, forceUsage)
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

func calculateAmount(ctx context.Context, targetDir string) (envelopes.Balance, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	targetDir, err := index.RootDirectory(targetDir)
	if err != nil {
		return envelopes.Balance{}, err
	}

	updated, err := index.LoadState(ctx, targetDir)
	if err != nil {
		return envelopes.Balance{}, err
	}

	var repo persist.RepositoryReader
	repo, err = filesystem.OpenRepositoryWithCache(ctx, filepath.Join(targetDir, index.RepoName), 10000)
	if err != nil {
		return envelopes.Balance{}, err
	}

	current, err := repo.Current(ctx)
	if err != nil {
		return envelopes.Balance{}, err
	}

	id, err := persist.Resolve(ctx, repo, current)
	if err != nil {
		return envelopes.Balance{}, err
	}

	var prevState envelopes.State
	// This happens when a repository is first initialized
	if !id.Equal(envelopes.ID{}) {
		var head envelopes.Transaction
		err = repo.LoadTransaction(ctx, id, &head)
		if err != nil {
			return envelopes.Balance{}, err
		}
		prevState = *head.State
	}

	return envelopes.CalculateAmount(prevState, *updated), nil
}
