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

	"github.com/marstr/envelopes"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var creditCmd = &cobra.Command{
	Use:     "credit {amount} {budget | account} [{budget | account}...]",
	Aliases: []string{"c", "cr"},
	Short:   "Makes funds available for one or more category of spending.",
	Args:    creditDebitArgValidation,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel, err := RootContext(cmd)
		if err != nil {
			return err
		}
		defer cancel()

		rawMagnitude := args[0]
		magnitude, err := envelopes.ParseBalance([]byte(rawMagnitude))
		if err != nil {
			return err
		}

		for _, rawBudget := range args[1:] {
			bdg, err := index.LoadBudget(ctx, rawBudget)
			if err != nil {
				return err
			}

			bdg.Balance = bdg.Balance.Add(magnitude)
			err = index.WriteBudget(ctx, rawBudget, *bdg)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(creditCmd)
}

func creditDebitArgValidation(cmd *cobra.Command, args []string) error {
	ctx, cancel, err := RootContext(cmd)
	if err != nil {
		return err
	}
	defer cancel()

	if argCount := len(args); argCount < 2 {
		return fmt.Errorf("too few arguments (%d). %q requires at least a balance and one budget or account", argCount, cmd.Name())
	}

	_, err = envelopes.ParseBalance([]byte(args[0]))
	if err != nil {
		return fmt.Errorf("%q not recognized as an amount", args[0])
	}

	for _, arg := range args[1:] {
		_, err := index.LoadBudget(ctx, arg)
		if err == nil {
			continue
		}

		_, err = index.LoadAccounts(ctx, arg)
		if err == nil {
			continue
		}

		return fmt.Errorf("%q was recognized as neither a budget nor an account", arg)
	}

	return nil
}
