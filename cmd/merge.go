/*
 * Copyright © 2025 Martin Strobel
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
	"fmt"
	"path/filepath"

	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:  "merge {refspec} [refspec]...",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := RootContext(cmd)
		defer cancel()

		root, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		var repo persist.RepositoryReader
		repo, err = filesystem.OpenRepositoryWithCache(ctx, filepath.Join(root, index.RepoName), 10000)
		if err != nil {
			logrus.Fatal(err)
		}

		var merged envelopes.State
		merged, err = persist.Merge(ctx, repo, append([]persist.RefSpec{persist.MostRecentTransactionAlias}, stringsToRefSpecs(args)...))
		if err != nil {
			logrus.Fatal(err)
		}

		err = index.CheckoutState(ctx, &merged, root, 0660)
		if err != nil {
			logrus.Fatal(err)
		}

		fmt.Println("Merge complete. Please check balances for accuracy, make any necessary reverts, and commit.")
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}

func stringsToRefSpecs(before []string) []persist.RefSpec {
	after := make([]persist.RefSpec, len(before))
	for i := range before {
		after[i] = persist.RefSpec(before[i])
	}
	return after
}
