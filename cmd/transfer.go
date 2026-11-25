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
	"errors"
	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/spf13/cobra"
)

var transferCmd = &cobra.Command{
	Use:     "transfer {amount} {src} {dest}",
	Aliases: []string{"t", "tran"},
	Short:   "Moves funds from one category of spending to another.",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel, err := RootContext(cmd)
		if err != nil {
			return err
		}
		defer cancel()

		rawSrc := args[1]
		rawDest := args[2]
		rawMagnitude := args[0]
		magnitude, err := envelopes.ParseBalance([]byte(rawMagnitude))
		if err != nil {
			return err
		}

		src, err := index.LoadBudget(ctx, rawSrc)
		if err != nil {
			return err
		}

		dest, err := index.LoadBudget(ctx, rawDest)
		if err != nil {
			return err
		}

		src.Balance = src.Balance.Sub(magnitude)
		dest.Balance = dest.Balance.Add(magnitude)

		srcErr := index.WriteBudget(ctx, rawSrc, *src)
		destErr := index.WriteBudget(ctx, rawDest, *dest)
		return errors.Join(srcErr, destErr)
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)
}
