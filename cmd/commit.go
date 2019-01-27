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
	"path/filepath"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/marstr/baronial/internal/index"
)

const (
	merchantFlag      = "merchant"
	merchantShorthand = "m"
	merchantDefault   = "Unknown"
	merchantUsage     = "The party receiving funds as part of this transaction."
)

const (
	commentFlag      = "comment"
	commentShorthand = "c"
	commentDefault   = ""
	commentUsage     = "Notes that may be helpful later when identifying this transaction."
)

const (
	timeFlag      = "time"
	timeShorthand = "t"
	timeDefault   = "<current date/time>"
	timeUsage     = "The time and date when this transaction occurred."
)

var commitConfig = viper.New()

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Create a transaction with the current impacts in the index.",
	Args:  cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		targetDir, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		accountsDir := filepath.Join(targetDir, index.AccountsDir)
		accounts, err := index.LoadAccounts(ctx, accountsDir)
		if err != nil {
			logrus.Fatal(err)
		}

		budgetDir := filepath.Join(targetDir, index.BudgetDir)
		budget, err := index.LoadBudget(ctx, budgetDir)
		if err != nil {
			logrus.Fatal(err)
		}

		persister := persist.FileSystem{
			Root: filepath.Join(targetDir, index.RepoName),
		}

		writer := persist.DefaultWriter{
			Stasher: persister,
		}

		parent, err := persister.Current(ctx)
		if err != nil {
			logrus.Fatal(err)
		}

		currentTransaction := envelopes.Transaction{
			Merchant: commitConfig.GetString(merchantFlag),
			Comment:  commitConfig.GetString(commentFlag),
			State: &envelopes.State{
				Accounts: accounts,
				Budget:   budget,
			},
			Time:   commitConfig.GetTime(timeFlag),
			Parent: parent,
		}

		err = writer.Write(ctx, currentTransaction)
		if err != nil {
			logrus.Fatal(err)
		}

		err = persister.WriteCurrent(ctx, &currentTransaction)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)

	commitConfig.SetDefault(timeFlag, time.Now().String())

	commitCmd.PersistentFlags().StringP(merchantFlag, merchantShorthand, merchantDefault, merchantUsage)
	commitCmd.PersistentFlags().StringP(commentFlag, commentShorthand, commentDefault, commentUsage)
	commitCmd.PersistentFlags().StringP(timeFlag, timeShorthand, timeDefault, timeUsage)

	err := commitConfig.BindPFlags(commitCmd.PersistentFlags())
	if err != nil {
		logrus.Fatal(err)
	}
}
