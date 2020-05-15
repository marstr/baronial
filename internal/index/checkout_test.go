package index

import (
	"context"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"testing"
	"time"

	"github.com/marstr/envelopes"
)

func TestCheckoutState_roundtrip(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()
	testCases := []*envelopes.State{
		{
			Accounts: map[string]envelopes.Balance{
				"checking": {"USD": big.NewRat(10096, 100)},
				"savings":  {"USD": big.NewRat(478302, 100)},
			},
			Budget: &envelopes.Budget{
				Balance: envelopes.Balance{"USD": big.NewRat(488398, 100)},
			},
		},
		{
			Accounts: map[string]envelopes.Balance{
				"checking": {"USD": big.NewRat(10096, 100)},
				"savings": {"USD": big.NewRat(478302, 100)},
			},
			Budget: &envelopes.Budget{
				Children: map[string]*envelopes.Budget{
					"foo": {Balance: envelopes.Balance{"USD": big.NewRat(10096, 100)}},
					"bar": {Balance: envelopes.Balance{"USD": big.NewRat(478302, 100)}},
				},
			},
		},
		{},
	}

	repoLocation, err := ioutil.TempDir("", "baronial_index_checkout_roundtrip_")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(repoLocation)

	err = os.Mkdir(path.Join(repoLocation, RepoName), os.ModePerm)
	if err != nil {
		t.Error(err)
		return
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			err = CheckoutState(ctx, tc, repoLocation, os.ModePerm)
			if err != nil {
				t.Error(err)
				return
			}

			got, err := LoadState(ctx, repoLocation)
			if err != nil {
				t.Error(err)
				return
			}

			diff := got.Subtract(*tc)

			if len(diff.Accounts) > 0 {
				t.Logf("Account balances didn't match.")
				t.Fail()
			}

			if diff.Budget != nil {
				if !diff.Budget.Balance.Equal(envelopes.Balance{}) || len(diff.Budget.Children) > 0 {
					t.Logf("Budget balances didn't match.")
					t.Fail()
				}
			}
		})
	}
}
