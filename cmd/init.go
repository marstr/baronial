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
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new Baronial repository in the current working directory.",
	Args:  cobra.NoArgs,
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

		const initCmdFailurePrefix = "unable to initialize repository: "

		const initialBranch = persist.DefaultBranch

		dirsToCreate := []string{
			index.RepoName,
			index.AccountsDir,
			index.BudgetDir,
		}

		for _, dir := range dirsToCreate {
			const dirCreationPermissions = 0750
			err := os.Mkdir(dir, os.FileMode(dirCreationPermissions))
			if os.IsExist(err) {
				// Intentionally Left Blank
			} else if err != nil {
				logrus.Fatal(initCmdFailurePrefix, err)
			}
		}

		var repo persist.RepositoryReaderWriter
		repo, err = filesystem.OpenRepositoryWithCache(ctx, index.RepoName, 10000)
		if err != nil {
			logrus.Fatal(err)
		}

		err = repo.WriteBranch(ctx, initialBranch, envelopes.ID{})
		if err != nil {
			logrus.Fatal(err)
		}

		err = repo.SetCurrent(ctx, initialBranch)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
