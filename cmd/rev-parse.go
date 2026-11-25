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
	"fmt"
	"github.com/marstr/baronial/internal/diagnostics"
	"path"

	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var revParseCmd = &cobra.Command{
	Use:   "rev-parse {refspec}",
	Short: "Prints a realized transaction ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel, err := RootContext(cmd)
		if err != nil {
			return err
		}
		defer cancel()

		root, err := index.RootDirectory(".")
		if err != nil {
			return err
		}

		var repo persist.RepositoryReader
		repo, err = filesystem.OpenRepositoryWithCache(ctx, path.Join(root, index.RepoName), 10000, filesystem.RepositoryLogger(diagnostics.Logger(ctx).With("component", "envelopes")))
		if err != nil {
			return err
		}

		id, err := persist.Resolve(ctx, repo, persist.RefSpec(args[0]))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), id.String())
		return err
	},
}

func init() {
	rootCmd.AddCommand(revParseCmd)
}
