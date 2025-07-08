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

	"github.com/marstr/envelopes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

var debitCmd = &cobra.Command{
	Use:     `debit {amount} {budget | account} [{budget | account}...]`,
	Aliases: []string{"d"},
	Short:   `Removes funds from a category of spending.`,
	Args:    creditDebitArgValidation,
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

		rawMagnitude := args[0]
		magnitude, err := envelopes.ParseBalance([]byte(rawMagnitude))
		if err != nil {
			logrus.Fatal(err)
		}

		for _, targetDir := range args[1:] {
			bdg, err := index.LoadBudget(ctx, targetDir)
			if err != nil {
				logrus.Fatal(err)
			}

			bdg.Balance = bdg.Balance.Sub(magnitude)
			err = index.WriteBudget(ctx, targetDir, *bdg)
			if err != nil {
				logrus.Fatal(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(debitCmd)
}
