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

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
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
		ctx := context.Background()
		var err error

		var root string
		root, err = index.RootDirectory(".")
		if err != nil {
			logrus.Error(err)
			return
		}

		persister := persist.FileSystem{Root: filepath.Join(root, index.RepoName)}
		reader := persist.DefaultLoader{
			Fetcher: persister,
		}

		currentRef, err := persister.Current(ctx)
		if err != nil {
			logrus.Error(err)
			return
		}

		resolver := persist.RefSpecResolver{
			Loader:        reader,
			Brancher:      persister,
			CurrentReader: persister,
		}

		currentID, err := resolver.Resolve(ctx, currentRef)
		if err != nil {
			logrus.Error(err)
			return
		}

		for !isEmptyID(currentID) {
			// TODO: refactor so that if parent was already loaded below, current re-uses that pre-loaded instance.
			var current envelopes.Transaction
			err = reader.Load(ctx, currentID, &current)
			if err != nil {
				logrus.Error(err)
				return
			}

			var diff envelopes.Impact

			if isEmptyID(current.Parent) {
				diff = envelopes.Impact(*current.State)
			} else {
				var parent envelopes.Transaction
				err = reader.Load(ctx, current.Parent, &parent)
				if err != nil {
					logrus.Error(err)
					return
				}

				diff = current.State.Subtract(*parent.State)
			}

			if len(args) == 0 || containsEntity(diff, args...) {
				err = format.ConcisePrintTransaction(ctx, cmd.OutOrStdout(), current)
				if err != nil {
					if cast, ok := err.(*os.PathError); ok {
						if cast.Path == "|1" {
							return
						}
					}
					logrus.Error(err)
					return
				}
			}
			currentID = current.Parent
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
