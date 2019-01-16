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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use: "log",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		root, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		persister := persist.FileSystem{Root: filepath.Join(root, index.RepoName)}

		currentID, err := persister.Current(ctx)
		if err != nil {
			logrus.Fatal(err)
		}

		for !isEmptyID(currentID) {
			result, err := persister.Fetch(ctx, currentID)
			if err != nil {
				logrus.Fatal(err)
			}

			var current envelopes.Transaction
			err = json.Unmarshal(result, &current)
			if err != nil {
				logrus.Fatal(err)
			}

			outputTransaction(ctx, os.Stdout, current)
			currentID = current.Parent()
		}
	},
}

func isEmptyID(subject envelopes.ID) bool {
	for _, val := range subject {
		if val != 0 {
			return false
		}
	}
	return true
}

func outputTransaction(_ context.Context, output io.Writer, subject envelopes.Transaction) error {
	fmt.Fprintln(output, subject.ID())
	fmt.Fprintf(output, "\tTime:    \t%v\n", subject.Time())
	fmt.Fprintf(output, "\tAmount:  \t%s\n", envelopes.FormatAmount(subject.Amount()))
	fmt.Fprintf(output, "\tMerchant:\t%s\n", subject.Merchant())
	fmt.Fprintf(output, "\tComment: \t%s\n", subject.Comment())
	return nil
}

func init() {
	rootCmd.AddCommand(logCmd)
}
