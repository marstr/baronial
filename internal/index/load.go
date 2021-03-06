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
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/marstr/envelopes"
	"github.com/marstr/units/data"
)

const (
	cashName    = "cash.txt"
	cashFileMax = int64(2 * data.Kilobyte)
)

// LoadState hydrates both accounts and the budget from the repository in the folder presented.
func LoadState(ctx context.Context, dirname string) (*envelopes.State, error) {
	var retval envelopes.State
	var err error

	dirname, err = RootDirectory(dirname)
	if err != nil {
		return nil, err
	}

	retval.Accounts, err = LoadAccounts(ctx, path.Join(dirname, AccountsDir))
	if err != nil {
		return nil, err
	}

	retval.Budget, err = LoadBudget(ctx, path.Join(dirname, BudgetDir))
	if err != nil {
		return nil, err
	}

	return &retval, nil
}

// LoadAccounts reads accounts balances from the current baronial index into memory.
func LoadAccounts(ctx context.Context, dirname string) (envelopes.Accounts, error) {
	var helper func(context.Context, string, envelopes.Accounts) (envelopes.Accounts, error)

	helper = func(ctx context.Context, dirname string, previous envelopes.Accounts) (envelopes.Accounts, error) {
		var entries []os.FileInfo
		var err error

		select {
		case <-ctx.Done():
			return envelopes.Accounts{}, ctx.Err()
		default:
			// Intentionally Left Blank
		}

		entries, err = ioutil.ReadDir(dirname)
		if err != nil {
			return envelopes.Accounts{}, err
		}

		found := previous

		for _, e := range entries {
			fullEntryName := filepath.Join(dirname, e.Name())

			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				// Look for accounts in this directory; skip directories that aren't accounts.
				// Note: In reality, this just means to skip the budget if the repository root is passed
				// into this function.
				found, err = helper(ctx, fullEntryName, found)
				if _, ok := err.(ErrNotAccount); ok {
					continue
				} else if err != nil {
					return envelopes.Accounts{}, err
				}
			} else if e.Name() == cashName {
				// If we've found a cash balance file in the accounts directory of a baronial repository, we've found an
				// account.
				var reader io.Reader
				var handle *os.File
				var contents []byte
				var bal envelopes.Balance

				// Determine the account name. If this is not an account, there is no need to continue.
				var accName string
				if name, err := AccountName(fullEntryName); err == nil {
					accName = name
				} else {
					return envelopes.Accounts{}, err
				}

				// Read the contents of the account
				handle, err = os.Open(fullEntryName)
				if err != nil {
					return envelopes.Accounts{}, err
				}
				reader = io.LimitReader(handle, cashFileMax)

				contents, err = ioutil.ReadAll(reader)
				handle.Close()
				if err != nil {
					return envelopes.Accounts{}, err
				}

				trimmed := strings.TrimSpace(string(contents))

				bal, err = envelopes.ParseBalance([]byte(trimmed))
				if err != nil {
					return envelopes.Accounts{}, err
				}

				found[accName] = bal
			}
		}

		return found, nil
	}

	return helper(ctx, dirname, make(envelopes.Accounts, 0))
}

// LoadBudget reads the budget portion of the current baronial index into memory.
func LoadBudget(ctx context.Context, dirname string) (retval *envelopes.Budget, err error) {
	var entries []os.FileInfo

	select {
	case <-ctx.Done():
		err = ctx.Err()
		return
	default:
		// Intentionally Left Blank
	}

	entries, err = ioutil.ReadDir(dirname)
	if err != nil {
		return
	}

	retval = new(envelopes.Budget)

	// Walk directory tree looking for files relevant to determining the balance of budgets.
	for _, e := range entries {
		fullEntryName := filepath.Join(dirname, e.Name())

		if e.IsDir() {
			// If the entry is a directory, load it as a child budget.

			var child *envelopes.Budget

			if strings.HasPrefix(e.Name(), ".") {
				continue
			}

			child, err = LoadBudget(ctx, fullEntryName)
			if err != nil {
				return
			}

			if retval.Children == nil {
				retval.Children = make(map[string]*envelopes.Budget)
			}
			retval.Children[e.Name()] = child
		} else if e.Name() == cashName {
			// If the directory entry is a file with the expected name for a file holding information about a cash
			// balance, parse the amount it contains and set that as the balance of this folder (excluding sub-balances).

			var reader io.Reader
			var contents []byte
			var handle *os.File

			handle, err = os.Open(fullEntryName)
			if err != nil {
				return
			}
			reader = handle

			reader = io.LimitReader(reader, cashFileMax)

			contents, err = ioutil.ReadAll(reader)
			handle.Close()
			if err != nil {
				return
			}

			retval.Balance, err = envelopes.ParseBalance(contents)
			if err != nil {
				return
			}
		}
	}

	return retval, nil
}
