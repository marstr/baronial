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
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
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
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if n := len(args); n != 2 {
			return fmt.Errorf("unexpected number of arguments %d", n)
		}

		if _, err = parseAmount(args[0]); err != nil {
			return fmt.Errorf("unable to find an amount in %q", args[0])
		}

		location := viper.GetString("location")
		if location, err = homedir.Expand(location); err != nil {
			return fmt.Errorf("no ledger found in the provided location")
		}

		fs := persist.FileSystem{
			Root: location,
		}
		id, err := fs.LoadCurrent(context.Background())

		loader := persist.DefaultLoader{
			Fetcher: fs,
		}

		latestTran, err := loader.LoadTransaction(context.Background(), id)
		if err != nil {
			return fmt.Errorf("ledger at the provided location is in disrepair")
		}

		latestState, err := loader.LoadState(context.Background(), latestTran.State())
		if err != nil {
			return fmt.Errorf("ledger at the provided location is in disrepair")
		}

		latestAccounts, err := loader.LoadAccounts(context.Background(), latestState.Accounts())
		if err != nil {
			return fmt.Errorf("ledger at the provided location is in disrepair")
		}

		if contender := args[1]; !latestAccounts.HasAccount(contender) {
			return fmt.Errorf("unrecognized account %q", contender)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		exitStatus := 1
		defer func() {
			os.Exit(exitStatus)
		}()

		amount, err := parseAmount(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to find an amount in %q\n", args[0])
			return
		}

		expandedLocation, err := homedir.Expand(viper.GetString("location"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Couldn't find the ledger to act upon.")
			return
		}
		fs := persist.FileSystem{
			Root: expandedLocation,
		}

		loader := persist.DefaultLoader{
			Fetcher: fs,
		}

		latestID, err := fs.LoadCurrent(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to determine the most recent Transaction.")
			return
		}

		ltsTran, ltsState, ltsAcc, ltsBudg, err := loader.LoadAll(context.Background(), latestID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to load the latest Transaction.")
			return
		}

		accountName := args[1]
		updatedAcc, ok := ltsAcc.AdjustBalance(accountName, amount)
		if !ok {
			fmt.Fprintf(os.Stderr, "Couldn't find the Account %q", accountName)
			return
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

		err = fs.WriteBudget(context.Background(), updatedBudg)
		err = fs.WriteAccounts(context.Background(), updatedAcc)
		err = fs.WriteState(context.Background(), updatedState)
		err = fs.WriteTransaction(context.Background(), created)

		err = fs.WriteCurrent(context.Background(), created)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to set this transaction as the most recent one.")
			return
		}
		exitStatus = 0
	},
}

func parseAmount(raw string) (result int64, err error) {
	raw = strings.TrimPrefix(raw, "$")
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return
	}
	result = int64(parsed*100 + .5)
	return
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
	transactionCmd.Flags().StringVarP(&merchant, "merchant", "i", "", "The merchant associated with this transaction.")
	transactionCmd.Flags().StringVarP(&comment, "comment", "m", "", "The comment associated with this comment.")
	transactionCmd.Flags().StringVarP(&target, "budget", "e", "#", "The budget(s) that should be impacted by the added transaction.")
}
