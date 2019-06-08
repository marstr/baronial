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
	"io"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var branchCmd = &cobra.Command{
	Use: "branch {name}",
	Short: "Creates a branch with a given name.",
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		// ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		// defer cancel()

		ctx := context.Background()

		indexRootDir, err  := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		repoDir := path.Join(indexRootDir, index.RepoName)

		head, err := index.ReadCurrent(repoDir)
		if err != nil {
			logrus.Fatal(err)
		}

		if len(args) > 1 {
			err = createBranch(ctx, head, repoDir, args[0])
			if err != nil {
				logrus.Fatal(err)
			}
		} else {
			err = printBranchList(ctx, os.Stdout, head, repoDir)
			if err != nil {
				logrus.Fatal(err)
			}
		}
	},
}

func createBranch(ctx context.Context, head index.RefSpec, repoDir, name string) error {
	transaction, err := head.Transaction(ctx, repoDir)
	if err != nil {
		return err
	}

	return index.WriteBranch(repoDir, name, transaction)
}

func printBranchList(ctx context.Context, output io.Writer, head index.RefSpec, repoDir string) (err error) {
	for branch := range index.ListBranches(ctx, repoDir) {
		_, err = fmt.Fprint(output, string(branch))
		if err != nil {
			return
		}
		if branch == head {
			fmt.Fprint(output, " *")
			if err != nil {
				return
			}
		}
		_, err = fmt.Fprintln(output)
		if err != nil {
			return
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
