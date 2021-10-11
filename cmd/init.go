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
	"github.com/marstr/envelopes/persist/json"
	"os"
	"path/filepath"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new Baronial repository.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		const initCmdFailurePrefix = "unable to initialize repository: "

		const initialBranch = persist.DefaultBranch

		var targetDir string
		if len(args) > 0 {
			targetDir = args[0]
		} else {
			var err error
			targetDir, err = filepath.Abs(".")
			if err != nil {
				logrus.Fatal(initCmdFailurePrefix, err)
			}

			targetDir = filepath.Clean(targetDir)
		}

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
		var err error
		repo, err = json.NewFileSystemRepository(index.RepoName)
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
