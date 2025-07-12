/*
 * Copyright © 2019 Martin Strobel
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/format"
	"github.com/marstr/baronial/internal/index"
)

var logCmd = &cobra.Command{
	Use:   "log [{account | budget}...]",
	Short: "Lists an overview of each transaction.",
	Long:  "",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	PreRunE: setPagedCobraOutput,
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

		var root string
		root, err = index.RootDirectory(".")
		if err != nil {
			logrus.Error(err)
			return
		}

		var repo persist.RepositoryReader
		repo, err = filesystem.OpenRepositoryWithCache(ctx, filepath.Join(root, index.RepoName), 10000)
		if err != nil {
			logrus.Fatal(err)
		}

		currentRef, err := repo.Current(ctx)
		if err != nil {
			logrus.Error(err)
			return
		}

		currentID, err := persist.Resolve(ctx, repo, currentRef)
		if err != nil {
			logrus.Error(err)
			return
		}

		walker := persist.Walker{Loader: repo}
		err = walker.Walk(ctx, func(ctx context.Context, id envelopes.ID, transaction envelopes.Transaction) error {
			impact, err := persist.LoadImpact(ctx, repo, transaction)
			if err != nil {
				return err
			}

			if len(args) == 0 || containsEntity(impact, args...) {
				err = format.ConcisePrintTransaction(ctx, cmd.OutOrStdout(), transaction)
				if err != nil {
					if cast, ok := err.(*os.PathError); ok {
						if cast.Path == "|1" {
							return nil
						}
					}
					return err
				}
			}
			return nil
		}, currentID)

		if err != nil {
			logrus.Error(err)
			return
		}
	},
}

// containsEntity inspects an Impact, to see if any of the entities provided were impacted by a transaction.
func containsEntity(diff envelopes.Impact, entities ...string) bool {
	for _, entity := range entities {
		entity = strings.TrimLeft(entity, ".")
		entity = strings.TrimPrefix(entity, "/")
		entity = strings.TrimPrefix(entity, "\\")

		if strings.HasPrefix(entity, index.BudgetDir) && containsBudget(diff, entity) {
			return true
		}

		if strings.HasPrefix(entity, index.AccountsDir) && containsAccount(diff, entity) {
			return true
		}
	}

	return false
}

func containsBudget(diff envelopes.Impact, budgetName string) bool {
	budgetName = strings.Replace(budgetName, "\\", "/", -1)
	budgetName = strings.Trim(budgetName, "/")
	splitName := strings.Split(budgetName, "/")
	if splitName[0] == "budget" {
		splitName = splitName[1:]
	}

	current := diff.Budget

	for _, entry := range splitName {
		if current == nil || current.Children == nil {
			return false
		}

		if child, ok := current.Children[entry]; ok {
			current = child
		} else {
			return false
		}
	}
	return true
}

func containsAccount(diff envelopes.Impact, accountName string) bool {
	accountName = strings.Replace(accountName, "\\", "/", -1)
	accountName = strings.Trim(accountName, "/")
	accountName = strings.TrimPrefix(accountName, "accounts/")
	_, ok := diff.Accounts[accountName]
	return ok
}

func isEmptyID(subject envelopes.ID) bool {
	for _, val := range subject {
		if val != 0 {
			return false
		}
	}
	return true
}

func init() {
	rootCmd.AddCommand(logCmd)
}
