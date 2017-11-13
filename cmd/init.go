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
	"path/filepath"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new ledger directory.",
	Long: `Creates a ledger directory in the directory specified, or in the current directory
if none is specified.`,
	Run: func(cmd *cobra.Command, args []string) {
		targetLocation := "."

		if argCount := len(args); argCount == 1 {
			targetLocation = args[0]
			err := os.MkdirAll(targetLocation, os.ModePerm)
			if err != nil {
				abs, _ := filepath.Abs(targetLocation)
				fmt.Fprintln(os.Stderr, "Couldn't create directory ", abs)
				os.Exit(1)
			}
		} else if argCount > 1 {
			fmt.Fprintln(os.Stderr, "Too many parameters passed to `ledger init`.")
			os.Exit(1)
		}

		firstBudget := envelopes.Budget{}
		firstAccounts := envelopes.Accounts{}
		firstState := envelopes.State{}.WithAccounts(firstAccounts.ID()).WithBudget(firstBudget.ID())

		root := envelopes.Transaction{}.WithComment("Ledger Created.").WithState(firstState.ID()).WithTime(time.Now())

		persister := persist.FileSystem{Root: targetLocation}

		err := persister.WriteBudget(context.Background(), firstBudget)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to write initial budget.")
			os.Exit(1)
		}

		err = persister.WriteAccounts(context.Background(), firstAccounts)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to write initial account list.")
			os.Exit(1)
		}

		err = persister.WriteState(context.Background(), firstState)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to write initial state.")
			os.Exit(1)
		}

		err = persister.WriteTransaction(context.Background(), root)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to write initial transaction.")
			os.Exit(1)
		}

		err = persister.WriteCurrent(context.Background(), root)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to updated HEAD.")
			os.Exit(1)
		}

		fmt.Printf("Initialized empty ledger in %s\n", targetLocation)
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
