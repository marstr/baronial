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
	"path"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var checkoutCmd = &cobra.Command{
	Use: "checkout {transaction id}",
	Aliases: []string{"ch"},
	Short: "Resets the index to show the balances at a particular transaction.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
		defer cancel()

		root, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}
		root = path.Join(root, index.RepoName)

		requested := index.RefSpec(args[0])

		var targetID envelopes.ID

		if transactionID, err := index.ReadBranch(root, requested); err == nil {
			targetID = transactionID
		} else {
			logrus.Warn("checking out a RefSpec that isn't a branch can cause data loss")
			targetID, err = requested.Transaction(ctx, root)
			if err != nil {
				logrus.Fatal(err)
			}
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

		err = index.CheckoutTransaction(ctx, &target, root, 0777)
		if err != nil {
			logrus.Fatal(err)
		}

		err = index.WriteCurrent(root, requested)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}