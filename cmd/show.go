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
	"os"
	"path"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/format"
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

		err = format.PrettyPrintTransaction(ctx, os.Stdout, loader, target)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
