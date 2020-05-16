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
	"fmt"
	"path"
	"time"

	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var revParseCmd = &cobra.Command{
	Use: "rev-parse {refspec}",
	Short: "Prints a realized transaction ID.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()

		root, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		fs := &persist.FileSystem{Root: path.Join(root, index.RepoName)}
		resolver := persist.RefSpecResolver{
			Loader: persist.DefaultLoader{Fetcher: fs},
			Brancher: fs,
			CurrentReader: fs,
		}

		id, err := resolver.Resolve(ctx, persist.RefSpec(args[0]))
		if err != nil {
			logrus.Fatal(err)
		}

		fmt.Println(id.String())
	},
}

func init() {
	rootCmd.AddCommand(revParseCmd)
}
