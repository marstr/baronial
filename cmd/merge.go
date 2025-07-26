/*
 * Copyright © 2025 Martin Strobel
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type MergeParameters struct {
	Comment     string            `json:"comment,omitempty"`
	Parents     []envelopes.ID    `json:"parent_ids"`
	ParentNames []persist.RefSpec `json:"parent_names"`
}

var mergeCmd = &cobra.Command{
	Use:  "merge {refspec} [refspec]...",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := RootContext(cmd)
		defer cancel()

		root, err := index.RootDirectory(".")
		if err != nil {
			logrus.Fatal(err)
		}

		var repo persist.RepositoryReader
		repo, err = filesystem.OpenRepositoryWithCache(ctx, filepath.Join(root, index.RepoName), 10000)
		if err != nil {
			logrus.Fatal(err)
		}

		var currentHead persist.RefSpec
		currentHead, err = repo.Current(ctx)
		if err != nil {
			logrus.Fatal("couldn't read what's currently checked out because: ", err)
		}

		heads := append([]persist.RefSpec{currentHead}, stringsToRefSpecs(args)...)

		var inProg bool
		inProg, err = MergeIsInProgress(ctx, root)
		if err != nil {
			logrus.Warn("couldn't see if previous merge is in progress because: ", err)
		}

		var mergeParams MergeParameters

		if inProg {
			err = MergeUnstowProgress(ctx, root, &mergeParams)
			if err != nil {
				logrus.Fatal("couldn't read the currently in-progress merge because: ", err)
			}
		}

		mergeParams.ParentNames = append(mergeParams.ParentNames, heads...)
		for _, head := range heads {
			var id envelopes.ID
			id, err = persist.Resolve(ctx, repo, head)
			if err != nil {
				logrus.Fatalf("couldn't resolve head %q because: %v", head, err)
			}

			mergeParams.Parents = append(mergeParams.Parents, id)
		}

		mergeParams.Comment = fmt.Sprintf(
			"Merging %s into %s",
			strings.Join(refSpecsToStrings(mergeParams.ParentNames)[1:], ", "),
			string(mergeParams.ParentNames[0]),
		)

		err = MergeStowProgress(ctx, root, mergeParams)
		if err != nil {
			logrus.Fatal(err)
		}

		var merged envelopes.State
		merged, err = persist.Merge(ctx, repo, heads)
		if err != nil {
			logrus.Fatal(err)
		}

		err = index.CheckoutState(ctx, &merged, root, 0660)
		if err != nil {
			logrus.Fatal(err)
		}

		fmt.Println("Merge complete. Please check balances for accuracy, make any necessary reverts, and commit.")
	},
}

func MergeIsInProgress(_ context.Context, repoLoc string) (bool, error) {
	_, err := os.Stat(getMergeParamsLoc(repoLoc))
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func MergeStowProgress(_ context.Context, repoLoc string, parameters MergeParameters) error {
	const filePermissions = 0660
	toWrite, err := json.Marshal(parameters)
	if err != nil {
		return fmt.Errorf("couldn't marshal merge parameters: %w", err)
	}

	err = os.WriteFile(getMergeParamsLoc(repoLoc), toWrite, filePermissions)
	if err != nil {
		return fmt.Errorf("couldn't write merge parameter file: %w", err)
	}

	return nil
}

func MergeUnstowProgress(_ context.Context, repoLoc string, destination *MergeParameters) error {
	contents, err := os.ReadFile(getMergeParamsLoc(repoLoc))
	if err != nil {
		return fmt.Errorf("couldn't read merge parameter file: %w", err)
	}

	err = json.Unmarshal(contents, destination)
	if err != nil {
		return fmt.Errorf("couldn't parse the merge parameter json: %w", err)
	}
	return nil
}

func MergeResetProgress(_ context.Context, repoLoc string) error {
	return os.Remove(getRevertParamsLoc(repoLoc))
}

func getMergeParamsLoc(repoLoc string) string {
	return filepath.Join(repoLoc, "merge.json")
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}

func stringsToRefSpecs(before []string) []persist.RefSpec {
	after := make([]persist.RefSpec, len(before))
	for i := range before {
		after[i] = persist.RefSpec(before[i])
	}
	return after
}

func refSpecsToStrings(before []persist.RefSpec) []string {
	after := make([]string, len(before))
	for i := range before {
		after[i] = string(before[i])
	}
	return after
}
