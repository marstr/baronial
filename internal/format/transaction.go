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

package format

import (
	"context"
	"fmt"
	"io"
	"path"
	"sort"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
)

func ConcisePrintTransaction(_ context.Context, output io.Writer, subject envelopes.Transaction) (err error) {
	_, err = fmt.Fprintln(output, subject.ID())
	if err != nil {
		return
	}
	if !subject.ActualTime.Equal(time.Time{}) {
		_, err = fmt.Fprintf(output, "\tActual Time:    \t%v\n", subject.ActualTime)
		if err != nil {
			return
		}
	}
	if !subject.PostedTime.Equal(time.Time{}) {
		_, err = fmt.Fprintf(output, "\tPosted Time:    \t%v\n", subject.PostedTime)
		if err != nil {
			return
		}
	}
	if !subject.EnteredTime.Equal(time.Time{}) {
		_, err = fmt.Fprintf(output, "\tEntered Time:    \t%v\n", subject.EnteredTime)
		if err != nil {
			return
		}
	}
	_, err = fmt.Fprintf(output, "\tAmount:  \t%s\n", subject.Amount)
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(output, "\tMerchant:\t%s\n", subject.Merchant)
	if err != nil {
		return
	}
	if subject.RecordId != "" {
		_, err = fmt.Fprintf(output, "\tBank Record ID:\t%s\n", subject.RecordId)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(output, "\tComment: \t%s\n", subject.Comment)
	if err != nil {
		return
	}
	return nil
}

// prettyPrintTransaction serializes a transaction into text in as pretty of a fashion as it can muster. It requires a
// `persist.Loader` in order to fetch any budget related information in its pursuit. Most notably, it must fetch the
// parent of this transaction to figure out the differences to each budget/account.
func PrettyPrintTransaction(
	ctx context.Context,
	output io.Writer,
	loader persist.Loader,
	subject envelopes.Transaction) error {
	var err error
	var impacts envelopes.Impact

	if subject.Parent.Equal(envelopes.ID{}) {
		impacts = envelopes.Impact(*subject.State)
	} else {
		var parent envelopes.Transaction
		err := loader.Load(ctx, subject.Parent, &parent)
		if err != nil {
			return err
		}
		impacts = subject.State.Subtract(*parent.State)
	}

	if !subject.ActualTime.Equal(time.Time{}) {
		_, err = fmt.Fprintf(output, "Actual Time:    \t%s\n", subject.ActualTime)
		if err != nil {
			return err
		}
	}
	if !subject.PostedTime.Equal(time.Time{}) {
		_, err = fmt.Fprintf(output, "Posted Time:    \t%s\n", subject.PostedTime)
		if err != nil {
			return err
		}
	}
	if !subject.EnteredTime.Equal(time.Time{}) {
		_, err = fmt.Fprintf(output, "Entered Time:    \t%s\n", subject.EnteredTime)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(output, "Merchant:\t%s\n", subject.Merchant)
	if err != nil {
		return err
	}
	if subject.RecordId != "" {
		_, err = fmt.Fprintf(output, "Bank Record ID:\t%s\n", subject.RecordId)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(output, "Amount:  \t%s\n", subject.Amount)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Parent:  \t%s\n", subject.Parent)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Comment: \t%s\n", subject.Comment)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Impacts:\n")
	if err != nil {
		return err
	}

	return PrettyPrintImpact(output, impacts)
}

func PrettyPrintImpact(output io.Writer, impacts envelopes.Impact) (err error) {
	_ , err = fmt.Fprintf(output, "\tAccounts:\n")
	if err != nil {
		return
	}
	for acc, delta := range impacts.Accounts {
		_ , err = fmt.Fprintf(output, "\t\t%s: %s\n", acc, delta)
		if err != nil {
			return
		}
	}

	flattened := flattenBudgets(impacts)

	sortedBudgetNames := make([]string, 0, len(flattened))
	for name := range flattened {
		sortedBudgetNames = append(sortedBudgetNames, name)
	}
	sort.Strings(sortedBudgetNames)

	_ , err = fmt.Fprintf(output, "\tBudgets:\n")
	for _, name := range sortedBudgetNames {
		_ , err = fmt.Fprintf(output, "\t\t%s: %s\n", name, flattened[name])
		if err != nil {
			return
		}
	}

	return
}

func flattenBudgets(diff envelopes.Impact) map[string]envelopes.Balance {
	retval := make(map[string]envelopes.Balance)
	var helper func(*envelopes.Budget, string, string)
	helper = func(current *envelopes.Budget, running, name string) {
		if current == nil {
			return
		}

		fullName := path.Join(running, name)

		if !current.Balance.Equal(envelopes.Balance{}) {
			retval[fullName] = current.Balance
		}

		for childName, child := range current.Children {
			helper(child, fullName, childName)
		}
	}

	helper(diff.Budget, "", "")

	return retval
}
