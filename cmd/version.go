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
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version  string
	revision string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version information about baronial.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: ", version)
		fmt.Printf("System: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("Go: ", runtime.Version())
		fmt.Println("Source Revision: ", revision)
	},
}

func init() {
	if version == "" {
		version = "unknown"
	}

	if revision == "" {
		revision = "unknown"
	}

	rootCmd.AddCommand(versionCmd)
}
