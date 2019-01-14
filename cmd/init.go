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
	"os"
	"path/filepath"

	"github.com/marstr/baronial/internal/index"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new Baronial repository.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		const initCmdFailurePrefix = "unable to initialize repository: "

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
			if err != nil {
				logrus.Fatal(initCmdFailurePrefix, err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
