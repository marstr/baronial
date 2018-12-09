// Copyright © 2018 Martin Strobel
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

	"github.com/marstr/baronial/internal/budget"
	"github.com/marstr/envelopes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var debitCmd = &cobra.Command{
	Use:     `debit {budget} {amount}`,
	Aliases: []string{"d"},
	Short:   `Removes funds from a category of spending.`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		targetDir := args[0]
		bdg, err := budget.Load(ctx, targetDir)
		if err != nil {
			logrus.Fatal(err)
		}

		rawMagnitude := args[1]
		magnitude, err := envelopes.ParseAmount(rawMagnitude)
		if err != nil {
			logrus.Fatal(err)
		}

		bdg = bdg.DecreaseBalance(magnitude)
		budget.Write(ctx, targetDir, bdg)
	},
}

func init() {
	rootCmd.AddCommand(debitCmd)
}
