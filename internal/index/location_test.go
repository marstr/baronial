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
	"path/filepath"
	"testing"
)

func TestRootDirectory(t *testing.T) {
	t.Run("expected", func(t *testing.T) {
		testCases := []struct {
			Path     string
			Expected string
		}{
			{
				"./testdata/test1/budget",
				"testdata/test1",
			},
			{
				"./testdata/test1",
				"testdata/test1",
			},
			{
				"./testdata/test1/accounts/citi",
				"testdata/test1",
			},
			{
				"./testdata/test1/accounts/citi/checking/cash.txt",
				"testdata/test1",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.Path, func(t *testing.T) {
				result, err := RootDirectory(tc.Path)
				if err != nil {
					t.Error(err)
					return
				}

				want, err := filepath.Abs(tc.Expected)
				if err != nil {
					t.Error(err)
					return
				}
				if result != want {
					t.Logf("\n\tgot:  %q\n\twant: %q", result, want)
					t.Fail()
				}
			})
		}
	})
}

func TestAccountName(t *testing.T) {
	t.Run("expected", func(t *testing.T) {
		testCases := []struct {
			Path     string
			Expected string
		}{
			{
				"./testdata/test1/accounts/citi/checking",
				"citi/checking",
			},
			{
				"./testdata/test1/accounts/citi",
				"citi",
			},
			{
				"./testdata/test1/accounts/citi/checking/cash.txt",
				"citi/checking",
			},
		}

		for _, tc := range testCases {
			t.Run("", func(t *testing.T) {
				got, err := AccountName(tc.Path)
				if err != nil {
					t.Error(err)
					return
				}

				if got != tc.Expected {
					t.Logf("\n\tgot:  %q\n\twant: %q", got, tc.Expected)
					t.Fail()
				}
			})
		}
	})

	t.Run("errored", func(t *testing.T) {
		testCases := []struct {
			Path     string
			Expected string
		}{
			{
				"./testdata/test1/budget/",
				"testdata/test1/budget",
			},
		}

		for _, tc := range testCases {
			got, err := AccountName(tc.Path)
			if got != "" {
				t.Logf("\n\tgot:  %q\n\twant: %q", got, "")
				t.Fail()
			}

			var want ErrNotAccount
			if rawWant, err := filepath.Abs(tc.Expected); err == nil {
				want = ErrNotAccount(rawWant)
			} else {
				t.Error(err)
				continue
			}

			if err != want {
				t.Logf("\n\tgot:  %T %v\n\twant: %T %v", err, err, want, want)
				t.Fail()
			}
		}
	})
}

func TestBudgetName(t *testing.T) {
	t.Run("expected", testBudgetNameExpected)
}

func testBudgetNameExpected(t *testing.T) {
	testCases := []struct {
		Path     string
		Expected string
	}{
		{
			"./testdata/test1/budget",
			"",
		},
		{
			"./testdata/test2/budget/rent",
			"rent",
		},
		{
			"./testdata/test3/budget/martin/bicycling",
			"martin/bicycling",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Path, func(t *testing.T) {
			got, err := BudgetName(tc.Path)
			if err != nil {
				t.Error(err)
				return
			}

			if got != tc.Expected {
				t.Logf("\n\tgot:  %q\n\twant: %q", got, tc.Expected)
				t.Fail()
			}
		})
	}
}
