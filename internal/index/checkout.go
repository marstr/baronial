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
	"os"
	"path"

	"github.com/marstr/envelopes"
)

// CheckoutState offers a shortcut to calling Checkout, should you know that the ID that has been
// handed to you points to an envelopes.State.
func CheckoutState(ctx context.Context, state *envelopes.State, targetDir string, perm os.FileMode) error {
	targetDir, err := RootDirectory(targetDir)
	if err != nil {
		return err
	}

	// Delete all existing contents to prevent inadvertent merge.

	accountsDir := path.Join(targetDir, AccountsDir)

	err = os.RemoveAll(accountsDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = os.Mkdir(accountsDir, perm|os.ModeDir|0110)
	if err != nil {
		return err
	}

	budgetDir := path.Join(targetDir, BudgetDir)
	err = os.RemoveAll(budgetDir)
	if err != nil {
		return err
	}

	err = os.Mkdir(budgetDir, perm|os.ModeDir|0110)
	if err != nil {
		return err
	}

	// Write accounts

	for accName, accBal := range state.Accounts {
		dirName := path.Join(accountsDir, accName)
		err = os.MkdirAll(dirName, perm|os.ModeDir|0110)
		if err != nil {
			return err
		}

		fileName := path.Join(dirName, cashName)
		err = os.WriteFile(fileName, []byte(accBal.String()), perm)
		if err != nil {
			return err
		}
	}

	// Write Budgets

	var processBudget func(context.Context, string, *envelopes.Budget) error

	processBudget = func(ctx context.Context, location string, budget *envelopes.Budget) error {
		err = os.MkdirAll(location, perm|os.ModeDir|0110)
		if err != nil {
			return err
		}

		err = WriteBudget(ctx, location, *budget)
		if err != nil {
			return err
		}

		for childName, child := range budget.Children {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Intentionally Left Blank
			}

			err = processBudget(ctx, path.Join(location, childName), child)
			if err != nil {
				return err
			}
		}

		return nil
	}
	if state.Budget != nil {
		return processBudget(ctx, budgetDir, state.Budget)
	}
	return nil
}

// CheckoutTransaction offers a shortcut to calling Checkout, should you know that the ID that has
// been handed to you points to an envelopes.Transaction.
func CheckoutTransaction(ctx context.Context, transaction *envelopes.Transaction, targetDir string, perm os.FileMode) error {
	return CheckoutState(ctx, transaction.State, targetDir, perm)
}
