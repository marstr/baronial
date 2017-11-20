// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/marstr/envelopes/persist"
	"github.com/marstr/randname"

	"github.com/spf13/cobra"
)

var (
	initialBalance float64
	name           string
	accComment     string
)

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		exitStatus := 1
		defer func() {
			os.Exit(exitStatus)
		}()

		roundedAmount := int64(amount + .5)

		if name == "" {
			name = randname.Generate()
		}

		fs := persist.FileSystem{
			Root: repoLocation,
		}

		l := persist.DefaultLoader{
			Fetcher: fs,
		}

		currentID, err := fs.LoadCurrent(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not find the most recent transaction.\n")
			return
		}

		ltsTran, ltsState, ltsAcc, ltsBudg, err := l.LoadAll(context.Background(), currentID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not load latest transaction because: %v\n", err)
			return
		}

		updatedAcc, ok := ltsAcc.AddAccount(name, roundedAmount)
		if !ok {
			fmt.Fprintf(os.Stderr, "Couldn't add account because there is already an account named %q\n", name)
			return
		}

		// TODO allow for distrubtion of the new funds accross budgets. (logic should be the same as a deposit)
		updatedBudg := ltsBudg

		updatedState := ltsState.WithAccounts(updatedAcc.ID())
		updatedState = updatedState.WithBudget(updatedBudg.ID())

		updatedTransaction := ltsTran.WithTime(time.Now())
		updatedTransaction = updatedTransaction.WithAmount(roundedAmount)
		updatedTransaction = updatedTransaction.WithState(updatedState.ID())
		updatedTransaction = updatedTransaction.WithComment(accComment)

		fs.WriteAccounts(context.Background(), updatedAcc)
		fs.WriteBudget(context.Background(), updatedBudg)
		fs.WriteState(context.Background(), updatedState)
		fs.WriteTransaction(context.Background(), updatedTransaction)
		fs.WriteCurrent(context.Background(), updatedTransaction)

		exitStatus = 0
	},
}

func init() {
	addCmd.AddCommand(accountCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// accountCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// accountCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	accountCmd.Flags().Float64VarP(&initialBalance, "balance", "b", 0, "The initial balance of the account. (default is $0.00)")
	accountCmd.Flags().StringVarP(&name, "name", "n", "", "The name that should be used to reference this account. If the name is left empty, the default will be assumed. (default is a random name)")
	accountCmd.Flags().StringVar(&accComment, "comment", "", "A comment to associate with the adding of this account.")
}
