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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/marstr/envelopes"
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
	briOutputDefault = briOutputOptionNone
	briOutputUsage = "How relevant transaction ID should be relayed."
	briOutputOptionNone = "none"
	briOutputOptionAny = "any"
)

var briSupportedOutputFormats = map[string]struct{}{
	briOutputOptionNone: {},
	briOutputOptionAny: {},
}

type ErrBriUnrecognizedOutputFormat string

func (err ErrBriUnrecognizedOutputFormat) Error() string {
	return fmt.Sprintf("unrecognized %s option provided: %q", briOutputFlag, string(err))
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
bank's identifier, the exit status will be 0.' In error conditions, a different
value will be returned.

Output Format:

none -> Only the exit status code will be set.
any -> Prints "true" if any transactions are associated with it, "false" otherwise.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		chosenOutput := bankRecordIdConfig.GetString(briOutputFlag)
		if _, ok := briSupportedOutputFormats[chosenOutput]; !ok {
			return ErrBriUnrecognizedOutputFormat(chosenOutput)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		subject := envelopes.BankRecordID(args[0])

		var rootDir string
		var err error
		rootDir, err = index.RootDirectory(".")
		if err != nil {
			logrus.Error(err)
		}

		recordFetcher := persist.FilesystemBankRecordIDIndex{
			Root: filepath.Join(rootDir, index.RepoName, repository.BankRecordIDIndexDirectory),
		}

		var exitCode int8
		switch bankRecordIdConfig.GetString(briOutputFlag) {
		case briOutputOptionNone:
			exitCode, err = processAnyRequest(ctx, ioutil.Discard, recordFetcher, subject)
		case briOutputOptionAny:
			exitCode, err = processAnyRequest(ctx, os.Stdout, recordFetcher, subject)
		default:
			logrus.Error(err)
			exitCode = 2
		}

		os.Exit(int(exitCode))
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

func processAnyRequest(ctx context.Context, output io.Writer, indexReader persist.FilesystemBankRecordIDIndex, bri envelopes.BankRecordID) (int8, error) {
	var exitCode int8
	var any bool
	var err error
	any, err = indexReader.HasBankRecordId(bri)
	if err != nil {
		return -1, err
	}
	if any {
		exitCode = 0
	} else {
		exitCode = 1
	}

	_, err = fmt.Fprintln(output, any)
	if err != nil{
		return exitCode, err
	}

	return exitCode, nil
}
