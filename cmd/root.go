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
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"time"

	"github.com/marstr/baronial/internal/diagnostics"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	logMaxBytesFlag = "log-max-bytes"
	timeoutFlag     = "timeout"
	timeoutDefault  = 0
	timeoutUsage    = `The maximum time to allow the command to run. The default is '0', which will let the command run indefinitely.

The value should be a positive decimal number, with an optional fraction and a unit suffix. For example, "300s", "1.5h", or "2h45m". Valid time units are "ns", "us", "ms", "s", "m", or "h".`
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "baronial",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "Manage your personal finances like a baron!",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// RootContext derives a cancellable command context with the configured timeout.
func RootContext(cmd *cobra.Command) (context.Context, context.CancelFunc, error) {
	timeout, err := cmd.Flags().GetDuration(timeoutFlag)
	if err != nil {
		return nil, nil, err
	}
	if timeout < 0 {
		return nil, nil, errors.New("timeout must not be negative")
	}
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		return ctx, cancel, nil
	}
	ctx, cancel := context.WithCancel(cmd.Context())
	return ctx, cancel, nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return executeCommand(rootCmd)
}

type executionKey struct{}

type execution struct {
	logger   *slog.Logger
	logFile  io.Closer
	cleanups []func() error
	started  time.Time
	command  *cobra.Command
	cancel   context.CancelFunc
}

func (run *execution) prepare(cmd *cobra.Command) error {
	levelName, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return err
	}
	level, enabled, err := diagnostics.ParseLevel(levelName)
	if err != nil {
		return err
	}
	maxBytes, err := cmd.Flags().GetInt64(logMaxBytesFlag)
	if err != nil {
		return err
	}
	if maxBytes <= 0 {
		return fmt.Errorf("--%s must be positive", logMaxBytesFlag)
	}
	timeout, err := cmd.Flags().GetDuration(timeoutFlag)
	if err != nil {
		return err
	}
	if timeout < 0 {
		return errors.New("timeout must not be negative")
	}
	path, err := cmd.Flags().GetString("log-file")
	if err != nil {
		return err
	}
	warn := func(err error) {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: diagnostic logging unavailable: %v\n", err)
	}
	ctx := context.WithValue(cmd.Context(), executionKey{}, run)
	run.started = time.Now()
	run.command = cmd
	if enabled {
		if path == "" {
			path, err = diagnostics.DefaultPath()
			if err != nil {
				warn(err)
				enabled = false
			}
		}
		if enabled {
			run.logger, run.logFile = diagnostics.Open(path, level, maxBytes, warn)
			run.logger = run.logger.With(
				"command", cmd.Name(), "invocation", rand.Text(),
				"version", version, "revision", revision,
				"platform", runtime.GOOS+"/"+runtime.GOARCH,
			)
			ctx = diagnostics.WithLogger(ctx, run.logger)
		}
	}
	if timeout > 0 {
		ctx, run.cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, run.cancel = context.WithCancel(ctx)
	}
	cmd.SetContext(ctx)
	diagnostics.Logger(ctx).InfoContext(ctx, "command started")
	return initConfig(cmd)
}

func executeCommand(root *cobra.Command) (err error) {
	run := &execution{}
	var restore []func()
	var wrap func(*cobra.Command)
	wrap = func(cmd *cobra.Command) {
		switch cmd.Name() {
		case "help", "completion", "__complete", "__completeNoDesc":
			return
		}
		previousArgs := cmd.Args
		previousContext := cmd.Context()
		if cmd.Run != nil || cmd.RunE != nil {
			cmd.Args = func(cmd *cobra.Command, args []string) error {
				if err := run.prepare(cmd); err != nil {
					return err
				}
				if previousArgs != nil {
					return previousArgs(cmd, args)
				}
				return nil
			}
			restore = append(restore, func() {
				cmd.Args = previousArgs
				cmd.SetContext(previousContext)
			})
		}
		for _, child := range cmd.Commands() {
			wrap(child)
		}
	}
	wrap(root)
	defer func() {
		for _, reset := range restore {
			reset()
		}
	}()
	_, err = root.ExecuteC()
	if errors.Is(err, errPagerClosed) {
		err = nil
	}
	for n := len(run.cleanups) - 1; n >= 0; n-- {
		err = errors.Join(err, run.cleanups[n]())
	}
	if run.command != nil {
		ctx := run.command.Context()
		logger := diagnostics.Logger(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "command failed", "error", err)
		}
		logger.InfoContext(ctx, "command finished", "success", err == nil, "duration", time.Since(run.started))
	}
	if run.logFile != nil {
		_ = run.logFile.Close() // The diagnostic writer reports failures through its warning callback.
	}
	if run.cancel != nil {
		run.cancel()
	}
	if err != nil {
		fmt.Fprintln(root.ErrOrStderr(), err)
	}
	return err
}

func warnCommand(cmd *cobra.Command, message string) {
	fmt.Fprintln(cmd.ErrOrStderr(), message)
	diagnostics.Logger(cmd.Context()).WarnContext(cmd.Context(), "command warning")
}

func init() {

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.baronial.yaml)")

	rootCmd.PersistentFlags().Duration(timeoutFlag, timeoutDefault, timeoutUsage)
	rootCmd.PersistentFlags().String("log-level", "info", "Diagnostic level: debug, info, warn, error, or off.")
	rootCmd.PersistentFlags().String("log-file", "", "Rotated diagnostic file path (default: per-user Baronial logs).")
	rootCmd.PersistentFlags().Int64(logMaxBytesFlag, int64(diagnostics.DefaultMaxBytes), "Maximum bytes per diagnostic file (default: 5 MiB).")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			return err
		}

		// Search config in home directory with name ".baronial" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".baronial")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		diagnostics.Logger(cmd.Context()).DebugContext(cmd.Context(), "using configuration file", "path", viper.ConfigFileUsed())
	} else {
		var missing viper.ConfigFileNotFoundError
		if !errors.As(err, &missing) {
			return fmt.Errorf("unable to read configuration: %w", err)
		}
	}
	return nil
}
