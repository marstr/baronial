/*
 * Copyright © 2020 Martin Strobel
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
	"errors"
	"fmt"
	"path"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/marstr/baronial/internal/index"
)

type diffCmd struct {
	cobra.Command
	Context context.Context
	Resolver persist.RefSpecResolver
}

func init() {
	diff := newDiffCmd()

	rootCmd.AddCommand(&diff.Command)
}

func newDiffCmd() *diffCmd {
	return newDiffCmdWithContext(context.Background())
}

func newDiffCmdWithContext(ctx context.Context) *diffCmd {
	retval := &diffCmd{
		Context: ctx,
	}

	retval.Command = cobra.Command{
		Use: "diff [refspec] [refspec]",
		Short: "Finds the difference between two states, be they from the index or two transactions.",
		Long: ``,
		Args: retval.processArgs,
		PreRunE: setPagedCobraOutput,
		Run: retval.run,
	}

	return retval
}

func (dc diffCmd) processArgs(cmd *cobra.Command, arg []string) error {
	if err := cobra.MaximumNArgs(2)(cmd, arg); err != nil {
		return err
	}

	for _, rs := range arg {
		id, err := dc.Resolver.Resolve(dc.Context, persist.RefSpec(rs))
		if err != nil || id.Equal(envelopes.ID{}){
			return fmt.Errorf("%q is not a valid refspec", rs)
		}
	}
	return nil
}

func (dc diffCmd) run(cmd *cobra.Command, args []string) {
	var err error
	defer func(){
		if err != nil {
			logrus.Error(err)
			return
		}
	}()

	var repoRoot string
	repoRoot, err = index.RootDirectory(".")
	if err != nil {
		return
	}

	var left, right *envelopes.State
	left, right, err = getDiffStates(dc.Context, args, repoRoot)
	if err != nil {
		return
	}

	diff := left.Subtract(*right)

	err = prettyPrintImpact(cmd.OutOrStdout(), diff)
	if err != nil {
		return
	}
}

func getDiffStates(ctx context.Context, args []string, indexRoot string) (*envelopes.State, *envelopes.State, error) {
	var err error
	var left, right *envelopes.State
	fs := persist.FileSystem{Root: path.Join(indexRoot, index.RepoName)}
	loader := persist.DefaultLoader{
		Fetcher: fs,
	}
	resolver := persist.RefSpecResolver{
		Loader:   loader,
		Brancher: fs,
		CurrentReader: fs,
	}

	loadFromRepository := func(ctx context.Context, rs persist.RefSpec) (*envelopes.State, error) {
		var targetID envelopes.ID
		targetID, err = resolver.Resolve(ctx, rs)
		if err != nil {
			return nil, err
		}
		var target envelopes.Transaction
		err = loader.Load(ctx, targetID, &target)
		if err != nil {
			return nil, err
		}
		return target.State, nil
	}

	index.LoadState(ctx, indexRoot)
	switch len(args) {
	case 0: // Compare current index against HEAD
		right, err = loadFromRepository(ctx, persist.MostRecentTransactionAlias)
		if err != nil {
			return nil, nil, err
		}

		left, err = index.LoadState(ctx, indexRoot)
		if err != nil{
			return nil, nil, err
		}
	case 1: // Compare current index against specified refspec
		right, err = loadFromRepository(ctx, persist.RefSpec(args[0]))
		if err != nil {
			return nil, nil, err
		}

		left, err = index.LoadState(ctx, indexRoot)
		if err != nil {
			return nil, nil, err
		}
	case 2: // Compare the two arbitrary refspecs
		right, err = loadFromRepository(ctx, persist.RefSpec(args[0]))
		if err != nil {
			return nil, nil, err
		}

		left, err = loadFromRepository(ctx, persist.RefSpec(args[0]))
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, errors.New("too many arguments")
	}

	return left, right, nil
}
