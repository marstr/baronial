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
	"github.com/marstr/envelopes"
	"math/big"
	"testing"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	// defaultTimeout = 4 * 24 * time.Hour
)

func TestLoadBudget(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	result, err := LoadBudget(ctx, "./testdata/test1/budget")
	if err != nil {
		t.Error(err)
		return
	}

	want := envelopes.Balance{"USD":big.NewRat(1234, 100)}

	if got := result.Balance; !got.Equal(want) {
		t.Logf("Raw Balance:\n\tgot:  %v\n\twant: %v", got, want)
		t.Fail()
	}

	if got := result.RecursiveBalance(); !got.Equal(want) {
		t.Logf("Recursive Balance:\n\tgot:  %v\n\twant: %v", got, want)
		t.Fail()
	}
}

func TestLoadAccounts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	testCases := []struct {
		location string
		expected map[string]envelopes.Balance
	}{
		{
			"./testdata/test1/accounts",
			map[string]envelopes.Balance{
				"citi/checking": {"USD": big.NewRat(1234, 100)},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			result, err := LoadAccounts(ctx, tc.location)
			if err != nil {
				t.Error(err)
				return
			}

			for _, name := range result.Names() {
				got, _ := result[name]
				want, ok := tc.expected[name]
				if !ok {
					t.Logf("unexpected budget: %s -> %v", name, got)
					t.Fail()
					continue
				}

				if !got.Equal(want) {
					t.Logf("%s\n\tgot:  %v\n\twant: %v", name, got, want)
					t.Fail()
				}

				delete(tc.expected, name)
			}

			for account, want := range tc.expected {
				if !result.HasAccount(account) {
					t.Logf("missing account: %s -> %v", account, want)
					t.Fail()
				}
			}
		})
	}
}
