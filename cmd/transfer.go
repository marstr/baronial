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
	"time"

	"github.com/marstr/envelopes/persist"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// transferCmd represents the transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer {amount} {src} {dst}",
	Short: "Moves funds between two envelopes or two accounts.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		fs := persist.FileSystem{
			Root: viper.GetString("location"),
		}

		loader := persist.DefaultLoader{
			Fetcher: fs,
		}

		currentID, err := fs.LoadCurrent(ctx)
		if err != nil {
			return
		}

		loader.

	},
}

func init() {
	RootCmd.AddCommand(transferCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transferCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transferCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
