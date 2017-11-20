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

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	account  string
	amount   float64
	target   string
	comment  string
	merchant string
	tranTime time.Time
)

// transactionCmd represents the transaction command
var transactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		roundedAmount := int64(amount + .5)

		fs := persist.FileSystem{
			Root: viper.GetString("location"),
		}
		fmt.Println("FS Root: ", fs.Root)
		loader := persist.DefaultLoader{
			Fetcher: fs,
		}

		latestID, err := fs.LoadCurrent(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to determine the most recent Transaction.")
			os.Exit(1)
		}

		ltsTran, ltsState, ltsAcc, ltsBudg, err := loader.LoadAll(context.Background(), latestID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to load the latest Transaction.")
			os.Exit(1)
		}

		updatedAcc, ok := ltsAcc.AdjustBalance(account, roundedAmount)
		if !ok {
			fmt.Fprintf(os.Stderr, "Couldn't find the Account %q", account)
			os.Exit(1)
		}

		// TODO modify this to an updated Budget that has been impacted by the transaction.
		updatedBudg := ltsBudg

		updatedState := ltsState.WithBudget(updatedBudg.ID())
		updatedState = updatedState.WithAccounts(updatedAcc.ID())

		created := envelopes.Transaction{}.WithParent(ltsTran.ID())
		created = created.WithComment(comment)
		created = created.WithMerchant(merchant)
		created = created.WithTime(time.Now())
		created = created.WithState(updatedState.ID())

		err = persist.WriteAll(context.Background(), fs, created, updatedState, ltsAcc, updatedBudg)

		err = fs.WriteCurrent(context.Background(), created)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unmable to set this transaction as the most recent one.")
			os.Exit(1)
		}
	},
}

func init() {
	addCmd.AddCommand(transactionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transactionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transactionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	transactionCmd.Flags().StringVarP(&merchant, "merchant", "m", "", "The merchant associated with this transaction.")
	transactionCmd.Flags().StringVarP(&comment, "comment", "c", "", "The comment associated with this comment.")
	transactionCmd.Flags().Float64VarP(&amount, "amount", "a", 0, "The amount of money transferred between accounts or budgets.")
	transactionCmd.Flags().StringVarP(&account, "account", "i", "", "The account(s) that were impacted by this account.")
	transactionCmd.Flags().StringVarP(&target, "budget", "b", "#", "The budget(s) that should be impacted by the added transaction.")
}
