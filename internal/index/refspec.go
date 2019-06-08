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

package index

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"sync"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
)

type (
	// RefSpec exposes operations on a string that is attempting to specify a particular Transaction ID.
	RefSpec string
	commitRefSpec RefSpec
	caretRefSpec RefSpec
	tildeRefSpec RefSpec
	branchRefSpec RefSpec
)

// ErrNoRefSpec indicates that a particular value was passed as if it could be interpreted as a RefSpec, but it could
// not.
type ErrNoRefSpec string

func (err ErrNoRefSpec) Error() string {
	return fmt.Sprintf("%s is not a valid RefSpec", string(err))
}

var (
	commitPattern = buildRegexpOnce(`^[0-9a-fA-F]{40}$`)
	caretPattern = buildRegexpOnce(`^(?P<parent>.+)\^$`)
	tildePattern = buildRegexpOnce(`^(?P<ancestor>.+)~(?P<jumps>\d+)$`)
)

// Transaction finds the ID of the transaction associated with this RefSpec.
func (rs RefSpec) Transaction(ctx context.Context, repoLocation string) (envelopes.ID, error) {
	root, err := RootDirectory(repoLocation)
	if err != nil {
		return envelopes.ID{}, err
	}

	root = path.Join(root, RepoName)

	return rs.transactionMux(ctx, root)
}

func (rs RefSpec) transactionMux(ctx context.Context, repoRoot string) (envelopes.ID, error) {
	if result, err := commitRefSpec(rs).Transaction(ctx, repoRoot); err == nil {
		return result, nil
	}

	if result, err := caretRefSpec(rs).Transaction(ctx, repoRoot); err == nil {
		return result, nil
	}

	if result, err := tildeRefSpec(rs).Transaction(ctx, repoRoot); err == nil {
		return result, nil
	}

	if result, err := branchRefSpec(rs).Transaction(ctx, repoRoot); err == nil {
		return result, nil
	}

	return envelopes.ID{}, ErrNoRefSpec(rs)
}

// buildRegexpOnce acts as a getter for regular expressions, lazily compiling patterns exactly once.
func buildRegexpOnce(pattern string) func() *regexp.Regexp {
	var guard sync.Once
	var built *regexp.Regexp
	return func() *regexp.Regexp {
		guard.Do(func(){
			built = regexp.MustCompile(pattern)
		})
		return built
	}
}

func (crs commitRefSpec) Transaction(ctx context.Context, repoRoot string) (envelopes.ID, error) {
	if !commitPattern().MatchString(string(crs)) {
		return envelopes.ID{}, ErrNoRefSpec(crs)
	}

	var result envelopes.ID

	err := result.UnmarshalText([]byte(crs))
	return result, err
}

func (crs caretRefSpec) Transaction(ctx context.Context, repoRoot string) (envelopes.ID, error) {
	matches := caretPattern().FindStringSubmatch(string(crs))
	if len(matches) < 2 {
		return envelopes.ID{}, ErrNoRefSpec(crs)
	}

	target, err := RefSpec(matches[1]).transactionMux(ctx, repoRoot)
	if err != nil {
		return envelopes.ID{}, err
	}

	fs := persist.FileSystem{Root: repoRoot}
	loader := persist.DefaultLoader{Fetcher: fs}

	loaded, err := persist.LoadAncestor(ctx, loader, target, 1)
	return loaded.ID(), nil
}

func (trs tildeRefSpec) Transaction(ctx context.Context, repoRoot string) (envelopes.ID, error) {
	matches := tildePattern().FindStringSubmatch(string(trs))
	if len(matches) < 3 {
		return envelopes.ID{}, ErrNoRefSpec(trs)
	}

	jumps, err := strconv.ParseInt(matches[2], 10, 32)
	if err != nil {
		return envelopes.ID{}, err
	}

	target, err := RefSpec(matches[1]).transactionMux(ctx, repoRoot)
	if err != nil {
		return envelopes.ID{}, err
	}

	fs := persist.FileSystem{Root: repoRoot}
	loader := persist.DefaultLoader{Fetcher: fs}

	loaded, err := persist.LoadAncestor(ctx, loader, target, uint(jumps))
	if err != nil {
		return envelopes.ID{}, err
	}
	return loaded.ID(), nil
}

func (brs branchRefSpec) Transaction(ctx context.Context, repoRoot string) (retval envelopes.ID, err error) {
	branchLoc := path.Join(repoRoot, "refs", "heads", string(brs))
	handle, err := os.Open(branchLoc)
	if err != nil {
		return
	}
	var contents [2 * cap(retval)]byte
	var n int
	n, err = handle.Read(contents[:])
	if err != nil {
		return
	}

	if n != cap(contents) {
		err = fmt.Errorf(
			"%s was not long enough to be a candidate for pointing to a Transaction ID (want: %v got: %v)",
			branchLoc,
			cap(contents),
			n)
		return
	}

	err = retval.UnmarshalText(contents[:])
	return
}
