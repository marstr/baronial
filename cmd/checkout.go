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
	"github.com/marstr/baronial/internal/diagnostics"
	"path"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var checkoutCmd = &cobra.Command{
	Use:     "checkout {refspec}",
	Aliases: []string{"ch"},
	Short:   "Resets the index to show the balances at a particular transaction.",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel, err := RootContext(cmd)
		if err != nil {
			return err
		}
		defer cancel()

		var root string
		root, err = index.RootDirectory(".")
		if err != nil {
			return err
		}
		root = path.Join(root, index.RepoName)

		requested := persist.RefSpec(args[0])

		var repo persist.RepositoryReaderWriter
		repo, err = filesystem.OpenRepositoryWithCache(ctx, root, 10000, filesystem.RepositoryLogger(diagnostics.Logger(ctx).With("component", "envelopes")))
		if err != nil {
			return err
		}

		var targetID envelopes.ID
		var transactionID envelopes.ID
		if transactionID, err = repo.ReadBranch(ctx, (string)(requested)); err == nil {
			targetID = transactionID
		} else {
			warnCommand(cmd, "checking out a RefSpec that isn't a branch can cause data loss")
			targetID, err = persist.Resolve(ctx, repo, requested)
			if err != nil {
				return err
			}
		}

		var target envelopes.Transaction
		err = repo.LoadTransaction(ctx, targetID, &target)
		if err != nil {
			return err
		}

		err = index.CheckoutTransaction(ctx, &target, root, 0660)
		if err != nil {
			return err
		}

		err = repo.SetCurrent(ctx, requested)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
