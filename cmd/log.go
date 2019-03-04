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
	"path/filepath"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var logCmd = &cobra.Command{
	Use:   "log [{account | budget}...]",
	Short: "Lists an overview of each transaction.",
	Long:  "",
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
		reader := persist.DefaultLoader{
			Fetcher: persister,
		}

		currentID, err := persister.Current(ctx)
		if err != nil {
			logrus.Fatal(err)
		}

		for !isEmptyID(currentID) {
			var current envelopes.Transaction
			err = reader.Load(ctx, currentID, &current)
			if err != nil {
				logrus.Fatal(err)
			}

			var parent envelopes.Transaction
			err = reader.Load(ctx, current.Parent, &parent)
			if !isEmptyID(current.Parent) && err != nil {

			}

			err = outputTransaction(ctx, os.Stdout, current)
			if err != nil {
				logrus.Fatal(err)
			}
			currentID = current.Parent
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

func outputTransaction(_ context.Context, output io.Writer, subject envelopes.Transaction) (err error) {
	_, err = fmt.Fprintln(output, subject.ID())
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(output, "\tTime:    \t%v\n", subject.Time)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(output, "\tAmount:  \t%s\n", subject.Amount)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(output, "\tMerchant:\t%s\n", subject.Merchant)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(output, "\tComment: \t%s\n", subject.Comment)
	if err != nil {
		return
	}
	return nil
}

func init() {
	rootCmd.AddCommand(logCmd)
}
