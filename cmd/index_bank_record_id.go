/*
Copyright © 2020 Martin Strobel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/baronial/internal/repository"
)

const (
	briOutputFlag = "output-format"
	briOutputShorthand = "f"
	briOutputDefault = "none"
	briOutputUsage = "How relevant transaction ID should be relayed."
)

var briSupportedOutputFormats = map[string]struct{}{
	"none": {},
}

var bankRecordIdConfig = viper.New()

// bankRecordIdCmd represents the bankRecordId command
var bankRecordIdCmd = &cobra.Command{
	Use:   "bank-record-id",
	Aliases: []string{"bri"},
	Short: "Find Transactions by bank assigned Record ID",
	Long: `Looks up transactions associated with a unique identifier
associated with a unique identifier assigned by a financial institution.

Status Code:

When no transactions are associated with the given bank record ID, the exit
status code will be set to 1. If even one transaction is associated with the
bank's identifier, the exit status will be 0.'

Output Format:

none -> Only the exit status code will be set.
any -> Prints "true" if any transactions are associated with it, "false" otherwise.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		chosenOutput := bankRecordIdConfig.GetString(briOutputFlag)
		if _, ok := briSupportedOutputFormats[chosenOutput]; !ok {
			return fmt.Errorf("unrecognized %s option provided: %q", briOutputFlag, chosenOutput)
		}
		return cobra.MaximumNArgs(1)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		var rootDir string
		var err error
		rootDir, err = index.RootDirectory(".")
		if err != nil {
			logrus.Error(err)
		}

		recordFetcher := persist.FilesystemBankRecordIDIndex{
			Root: filepath.Join(rootDir, index.RepoName, repository.BankRecordIDIndexDirectory),
		}

		fmt.Printf("ready to read from: %s", recordFetcher.Root)
	},
}

func init() {
	indexCmd.AddCommand(bankRecordIdCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bankRecordIdCmd.PersistentFlags().String("foo", "", "A help for foo")
	bankRecordIdCmd.PersistentFlags().StringP(briOutputFlag, briOutputShorthand, briOutputDefault, briOutputUsage)


	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bankRecordIdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	bankRecordIdConfig.BindPFlags(bankRecordIdCmd.PersistentFlags())
}
