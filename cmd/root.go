// Copyright © 2019 Martin Strobel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	timeoutFlag    = "timeout"
	timeoutDefault = 0
	timeoutUsage   = `The maximum time to allow the command to run. The default is '0', which will let the command run indefinitely.

The value should be a positive decimal number, with an optional fraction and a unit suffix. For example, "300s", "1.5h", or "2h45m". Valid time units are "ns", "us", "ms", "s", "m", or "h".`
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "baronial",
	Short: "Manage your personal finances like a baron!",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

var (
	parseTimeout sync.Once
	rootContext  context.Context
	rootCancel   context.CancelFunc
)

func RootContext(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	parseTimeout.Do(func() {
		var timeout time.Duration
		var err error
		timeout, err = cmd.Flags().GetDuration(timeoutFlag)
		if err != nil {
			logrus.Fatal(err)
		}

		if timeout > 0 {
			rootContext, rootCancel = context.WithTimeout(context.Background(), timeout)

		} else {
			rootContext, rootCancel = context.WithCancel(context.Background())
		}
	})

	return rootContext, rootCancel
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.baronial.yaml)")

	rootCmd.PersistentFlags().Duration(timeoutFlag, timeoutDefault, timeoutUsage)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".baronial" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".baronial")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
