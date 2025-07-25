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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type RevertParameters struct {
	Comment string         `json:"comment,omitempty"`
	Reverts []envelopes.ID `json:"reverts"`
}

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
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := RootContext(cmd)
		defer cancel()

		root, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		repoLoc := filepath.Join(root, index.RepoName)
		var repo persist.RepositoryReaderWriter
		repo, err = filesystem.OpenRepositoryWithCache(ctx, repoLoc, 10000)
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

		var inProg bool
		inProg, err = RevertIsInProgress(ctx, repoLoc)
		if err != nil {
			logrus.Warn("couldn't see if previous revert was in progress because: ", err)
		}

		var revertParams RevertParameters

		if inProg {
			err = RevertUnstowProgress(ctx, repoLoc, &revertParams)
			if err != nil {
				logrus.Fatal("couldn't read the currently in-progress revert because: ", err)
			}
		}

		revertParams.Reverts = append(revertParams.Reverts, id)
		revertParams.Comment = getRevertComment(revertParams.Reverts)

		err = RevertStowProgress(ctx, repoLoc, revertParams)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to stow revert information. Please reset to the last known good state.")
			logrus.Fatal(err)
		}

		fmt.Printf("Undid the effects of transaction %s. Please check current balances for accuracy, make any necessary edits, then commit.", id)
	},
}

func RevertIsInProgress(_ context.Context, repoLoc string) (bool, error) {
	_, err := os.Stat(getRevertParamsLoc(repoLoc))
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func RevertStowProgress(_ context.Context, repoLoc string, parameters RevertParameters) error {
	const filePermissions = 0660 // owner and group can read and write. The file is not executable. Other users cannot read it.
	toWrite, err := json.Marshal(parameters)
	if err != nil {
		return err
	}

	return os.WriteFile(getRevertParamsLoc(repoLoc), toWrite, filePermissions)
}

func RevertUnstowProgress(_ context.Context, repoLoc string, destination *RevertParameters) error {
	contents, err := os.ReadFile(getRevertParamsLoc(repoLoc))
	if err != nil {
		return err
	}

	return json.Unmarshal(contents, destination)
}

func RevertResetProgress(_ context.Context, repoLoc string) error {
	return os.Remove(getRevertParamsLoc(repoLoc))
}

func getRevertParamsLoc(repoLoc string) string {
	return filepath.Join(repoLoc, "revert.json")
}

func getRevertComment(reverts []envelopes.ID) string {
	var buf bytes.Buffer

	fmt.Fprint(&buf, "Reverts ")

	seenAny := false
	for _, id := range reverts {
		seenAny = true
		fmt.Fprintf(&buf, "%s, ", id)
	}

	if seenAny {
		buf.Truncate(buf.Len() - 2)
	}

	return buf.String()
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
