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
	"time"

	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var transferCmd = &cobra.Command{
	Use:     "transfer {src} {dest} {amount}",
	Aliases: []string{"t", "tran"},
	Short:   "Moves funds from one category of spending to another.",
	Args:    cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		rawSrc := args[0]
		rawDest := args[1]
		rawMagnitude := args[2]
		magnitude, err := envelopes.ParseBalance(rawMagnitude)
		if err != nil {
			logrus.Fatal(err)
		}

		src, err := index.LoadBudget(ctx, rawSrc)
		if err != nil {
			logrus.Fatal(err)
		}

		dest, err := index.LoadBudget(ctx, rawDest)
		if err != nil {
			logrus.Fatal(err)
		}

		src.Balance -= magnitude
		dest.Balance += magnitude

		err = index.WriteBudget(ctx, rawSrc, *src)
		if err != nil {
			logrus.Error(err)
		}

		err = index.WriteBudget(ctx, rawDest, *dest)
		if err != nil {
			logrus.Error(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)
}
