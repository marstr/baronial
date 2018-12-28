// Copyright © 2018 Martin Strobel
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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	// RepoName is the directory that contains metadata about a repository.
	RepoName = ".baronial"

	// BudgetDir is the directory that contains the budget definition of a repository.
	BudgetDir = "budget"

	// AccountsDir is the directory that contains the account listing of a repository.
	AccountsDir = "accounts"
)

type (
	// ErrNoRootDir is returned when a location-based function is called on a directory that is not in a baronial
	// repository.
	ErrNoRootDir string

	// ErrNotAccount is returned when AccountName is called with a directory that is not inside the "accounts" folder in
	// baronial repository.
	ErrNotAccount string

	// ErrNotBudget is returned when BudgetName is called with a directory that is not inside the "budget" folder in a
	// baronial repository.
	ErrNotBudget string
)

func (e ErrNoRootDir) Error() string {
	return fmt.Sprintf("not a baronial repository (or any of the parent directories): %s", string(e))
}

func (e ErrNotAccount) Error() string {
	return fmt.Sprintf("%q is not in the accounts directory", string(e))
}

func (e ErrNotBudget) Error() string {
	return fmt.Sprintf("%q is not in the budget directory", string(e))
}

// RootDirectory walks a directory structure from a given starting point looking for the root of a baronial
// repository.
func RootDirectory(dirname string) (string, error) {
	var err error
	var entry os.FileInfo

	dirname, err = filepath.Abs(dirname)
	if err != nil {
		return "", err
	}

	entry, err = os.Stat(dirname)
	if err != nil {
		return "", err
	}

	if !entry.IsDir() {
		dirname = path.Dir(dirname)
	}

	contents, err := ioutil.ReadDir(dirname)
	if err != nil {
		return "", err
	}

	for _, entry := range contents {
		if entry.IsDir() && entry.Name() == RepoName {
			return path.Clean(dirname), nil
		}
	}

	parent := path.Dir(dirname)
	if parent == dirname {
		return "", ErrNoRootDir(dirname)
	}

	parent, err = RootDirectory(parent)
	if err == nil {
		return parent, nil
	} else if _, ok := err.(ErrNoRootDir); ok {
		return "", ErrNoRootDir(dirname)
	} else {
		return "", err
	}
}

// AccountName finds the name of an Account as it exists in a baronial repository index.
func AccountName(dirname string) (string, error) {
	root, err := RootDirectory(dirname)
	if err != nil {
		return "", err
	}

	dirname, err = filepath.Abs(dirname)
	if err != nil {
		return "", nil
	}

	accountPrefix := path.Join(root, AccountsDir)
	if !strings.HasPrefix(dirname, accountPrefix) {
		return "", ErrNotAccount(dirname)
	}

	if info, err := os.Stat(dirname); err != nil {
		return "", err
	} else if !info.IsDir() {
		dirname = path.Dir(dirname)
	}

	return strings.TrimLeft(strings.TrimPrefix(dirname, accountPrefix), "/"), nil
}

// BudgetName finds the name of a Budget as it exists in a baronial repository index.
func BudgetName(dirname string) (string, error) {
	root, err := RootDirectory(dirname)
	if err != nil {
		return "", err
	}

	dirname, err = filepath.Abs(dirname)
	if err != nil {
		return "", nil
	}

	budgetPrefix := path.Join(root, BudgetDir)
	if !strings.HasPrefix(dirname, budgetPrefix) {
		return "", ErrNotBudget(dirname)
	}

	return strings.TrimPrefix(dirname, budgetPrefix), nil
}
