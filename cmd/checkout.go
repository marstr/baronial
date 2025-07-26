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
	"path"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var checkoutCmd = &cobra.Command{
	Use:     "checkout {refspec}",
	Aliases: []string{"ch"},
	Short:   "Resets the index to show the balances at a particular transaction.",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := RootContext(cmd)
		defer cancel()

		var root string
		var err error
		root, err = index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}
		root = path.Join(root, index.RepoName)

		requested := persist.RefSpec(args[0])

		var repo persist.RepositoryReaderWriter
		repo, err = filesystem.OpenRepositoryWithCache(ctx, root, 10000)
		if err != nil {
			logrus.Fatal(err)
		}

		var targetID envelopes.ID
		var transactionID envelopes.ID
		if transactionID, err = repo.ReadBranch(ctx, (string)(requested)); err == nil {
			targetID = transactionID
		} else {
			logrus.Warn("checking out a RefSpec that isn't a branch can cause data loss")
			targetID, err = persist.Resolve(ctx, repo, requested)
			if err != nil {
				logrus.Fatal(err)
			}
		}

		var target envelopes.Transaction
		err = repo.LoadTransaction(ctx, targetID, &target)
		if err != nil {
			logrus.Fatal(err)
		}

		err = index.CheckoutTransaction(ctx, &target, root, 0660)
		if err != nil {
			logrus.Fatal(err)
		}

		err = repo.SetCurrent(ctx, requested)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
