/*
 * Copyright © 2019 Martin Strobel
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

package index

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/marstr/collection"
	"github.com/marstr/envelopes"
)

// BranchLocation gives the location of a branch relative to a repository.
func BranchLocation(repoRoot, name string) string {
	return path.Join(repoRoot, "refs", "heads", name)
}

// ReadBranch fetches the contents of a file holding a Transaction ID.
func ReadBranch(repoRoot, name string) (retval envelopes.ID, err error) {
	branchLoc := BranchLocation(repoRoot, name)
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

// WriteBranch saves a Transaction ID to a file.
func WriteBranch(repoRoot, name string, transaction envelopes.ID) error {
	branchLoc := BranchLocation(repoRoot, name)
	handle, err := os.Create(branchLoc)
	if err != nil {
		return err
	}

	_, err = handle.WriteString(transaction.String())
	return err
}

// ReadCurrent fetches the content of the file pointing at the current HEAD.
func ReadCurrent(repoRoot string) (retval RefSpec, err error) {
	fileLoc := path.Join(repoRoot, "current.txt")

	contents, err := ioutil.ReadFile(fileLoc)
	if err != nil {
		return "", err
	}

	return RefSpec(strings.TrimSpace(string(contents))), nil
}

// ListBranches enumerates all of the branches named in a particular repository.
func ListBranches(ctx context.Context, repoRoot string) <-chan RefSpec {
	branchFiles := collection.Directory{
		Location: path.Join(repoRoot, "refs", "heads"),
		Options: collection.DirectoryOptionsExcludeDirectories | collection.DirectoryOptionsRecursive,
	}

	output := make(chan RefSpec)
	go func(results chan RefSpec) error {
		defer close(output)
		branches := branchFiles.Enumerate(ctx.Done()).Select(func(x interface{}) interface{}{
			return path.Base(x.(string))
		})
		for branch := range branches {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case output <- RefSpec(branch.(string)):
				// Intentionally Left Blank
			}
		}
		return nil
	}(output)

	return output
}