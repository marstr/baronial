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
	"fmt"
	"github.com/marstr/baronial/internal/diagnostics"
	"io"
	"path"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var branchCmd = &cobra.Command{
	Use:     "branch {name}",
	Aliases: []string{"br"},
	Short:   "Creates a branch with a given name.",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel, err := RootContext(cmd)
		if err != nil {
			return err
		}
		defer cancel()

		var indexRootDir string
		indexRootDir, err = index.RootDirectory(".")
		if err != nil {
			return err
		}

		var repo persist.RepositoryReaderWriter
		repo, err = filesystem.OpenRepositoryWithCache(ctx, path.Join(indexRootDir, index.RepoName), 10000, filesystem.RepositoryLogger(diagnostics.Logger(ctx).With("component", "envelopes")))
		if err != nil {
			return err
		}

		var head persist.RefSpec
		head, err = repo.Current(ctx)
		if err != nil {
			return err
		}

		if len(args) > 0 {
			branchName := args[0]

			var target envelopes.ID
			target, err = persist.Resolve(ctx, repo, head)
			if err != nil {
				return err
			}

			err = repo.WriteBranch(ctx, branchName, target)
			if err != nil {
				return err
			}
		} else {
			err = printBranchList(ctx, cmd.OutOrStdout(), repo, head)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func printBranchList(ctx context.Context, output io.Writer, lister persist.BranchLister, head persist.RefSpec) error {
	branches, err := lister.ListBranches(ctx)
	if err != nil {
		return err
	}

	for branch := range branches {
		_, err = fmt.Fprint(output, string(branch))
		if err != nil {
			return err
		}
		if branch == (string)(head) {
			_, err = fmt.Fprint(output, " *")
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(output)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
