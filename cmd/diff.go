/*
 * Copyright © 2020 Martin Strobel
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
	"errors"
	"path"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/format"
	"github.com/marstr/baronial/internal/index"
)

var diffCmd = &cobra.Command{
	Use:     "diff [refspec] [refspec]",
	Short:   "Finds the difference between two states, be they from the index or two transactions.",
	Long:    ``,
	Args:    cobra.MaximumNArgs(2),
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

		var repoRoot string
		repoRoot, err = index.RootDirectory(".")
		if err != nil {
			return
		}

		var left, right *envelopes.State
		left, right, err = getDiffStates(ctx, args, repoRoot)
		if err != nil {
			return
		}

		diff := left.Subtract(*right)

		err = format.PrettyPrintImpact(cmd.OutOrStdout(), diff)
		if err != nil {
			return
		}
	},
}

func getDiffStates(ctx context.Context, args []string, indexRoot string) (*envelopes.State, *envelopes.State, error) {
	var err error
	var left, right *envelopes.State
	var repo persist.RepositoryReader

	repo, err = filesystem.OpenRepositoryWithCache(ctx, path.Join(indexRoot, index.RepoName), 10000)
	if err != nil {
		logrus.Fatal(err)
	}

	loadFromRepository := func(ctx context.Context, rs persist.RefSpec) (*envelopes.State, error) {
		var targetID envelopes.ID
		targetID, err = persist.Resolve(ctx, repo, rs)
		if err != nil {
			return nil, err
		}
		var target envelopes.Transaction
		err = repo.LoadTransaction(ctx, targetID, &target)
		if err != nil {
			return nil, err
		}
		return target.State, nil
	}

	index.LoadState(ctx, indexRoot)
	switch len(args) {
	case 0: // Compare current index against HEAD
		right, err = loadFromRepository(ctx, persist.MostRecentTransactionAlias)
		if err != nil {
			return nil, nil, err
		}

		left, err = index.LoadState(ctx, indexRoot)
		if err != nil {
			return nil, nil, err
		}
	case 1: // Compare current index against specified refspec
		right, err = loadFromRepository(ctx, persist.RefSpec(args[0]))
		if err != nil {
			return nil, nil, err
		}

		left, err = index.LoadState(ctx, indexRoot)
		if err != nil {
			return nil, nil, err
		}
	case 2: // Compare the two arbitrary refspecs
		right, err = loadFromRepository(ctx, persist.RefSpec(args[0]))
		if err != nil {
			return nil, nil, err
		}

		left, err = loadFromRepository(ctx, persist.RefSpec(args[0]))
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, errors.New("too many arguments")
	}

	return left, right, nil
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
