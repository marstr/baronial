// Copyright © 2017 Martin Strobel
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
	"sort"

	"github.com/marstr/envelopes/persist"

	"github.com/spf13/cobra"
)

var useDebug bool

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "balance",
	Short: "Displays the current balance of each Budget and Account",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		persister := persist.FileSystem{
			Root: ".",
		}
		loader := persist.DefaultLoader{
			Fetcher: persister,
		}

		currentID, err := persister.LoadCurrent(context.Background())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not find current ledger status.")
			os.Exit(1)
		}

		latestTransaction, err := loader.LoadTransaction(context.Background(), *currentID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Not able to read latest transaction from disk.")
			os.Exit(1)
		}

		latestState, err := loader.LoadState(context.Background(), latestTransaction.State())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Not able to read latest state from disk.")
			os.Exit(1)
		}

		latestBudget, err := loader.LoadBudget(context.Background(), latestState.Budget())
		if err != nil {
			fmt.Fprintln(os.Stderr, "Not able to parse latest budget.")
			os.Exit(1)
		}

		fmt.Printf("Total Balance: %d\n", latestBudget.RecursiveBalance())

		fmt.Printf("Balance: %d\n", latestBudget.Balance())

		latestChildren := latestBudget.Children()
		if len(latestChildren) > 0 {
			childNames := make([]string, 0, len(latestChildren))
			for name := range latestChildren {
				childNames = append(childNames, name)
			}

			sort.Strings(childNames)

			fmt.Println("Children:")
			for _, currentName := range childNames {
				fmt.Printf("\t%s: $%0.2f\n", currentName, float64(latestChildren[currentName].RecursiveBalance())/100)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
