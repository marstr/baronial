/*
Copyright © 2025 Martin Strobel

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// revertCmd represents the revert command
var revertCmd = &cobra.Command{
	Use:   "revert {ref-spec}",
	Short: "Deletes a previous transaction.",
	Long: `Logs a new transaction with the opposite effect as a previous transaction. The
new transaction also has a reference to the old one, so that any tooling will
know to skip the effects of both. 

It is not advised to revert a revert. Even if this tool handles it well, it
complicates all future tools and many may not do a good job. If you 
accidentally reverted a transaction, just commit a new transaction that is
identical to the original.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := RootContext(cmd)
		defer cancel()

		root, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		var repo persist.RepositoryReaderWriter
		repo, err = filesystem.OpenRepositoryWithCache(ctx, path.Join(root, index.RepoName), 10000)
		if err != nil {
			logrus.Fatal(err)
		}

		id, err := persist.Resolve(ctx, repo, persist.RefSpec(args[0]))
		if err != nil {
			logrus.Fatal(err)
		}

		var toRevert envelopes.Transaction
		err = repo.LoadTransaction(ctx, id, &toRevert)
		if err != nil {
			logrus.Fatal(err)
		}

		var delta envelopes.Impact
		delta, err = persist.LoadImpact(ctx, repo, toRevert)
		if err != nil {
			logrus.Fatal(err)
		}

		var headID envelopes.ID
		headID, err = persist.Resolve(ctx, repo, persist.MostRecentTransactionAlias)
		if err != nil {
			logrus.Fatal(err)
		}

		var head envelopes.Transaction
		err = repo.LoadTransaction(ctx, headID, &head)
		if err != nil {
			logrus.Fatal(err)
		}

		updated := envelopes.State(head.State.Add(envelopes.State(delta.Negate())))

		err = index.CheckoutState(ctx, &updated, root, os.ModePerm)
		if err != nil {
			logrus.Fatal(err)
		}

		fmt.Printf("Undid the effects of transaction %s. Please check current balances for accuracy, make any necessary edits, then commit.", id)
	},
}

func init() {
	rootCmd.AddCommand(revertCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// revertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// revertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
