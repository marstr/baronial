package index

import (
	"context"
	"io/ioutil"
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
				"checking": 10096,
				"savings": 478302,
			},
			Budget: &envelopes.Budget{
				Balance: 488398,
			},
		},
		{
			Accounts: map[string]envelopes.Balance{
				"checking": 10096,
				"savings": 478302,
			},
			Budget: &envelopes.Budget{
				Children: map[string]*envelopes.Budget{
					"foo": {Balance: 10096},
					"bar": {Balance: 478302},
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
				if diff.Budget.Balance != 0 || len(diff.Budget.Children) > 0 {
					t.Logf("Budget balances didn't match.")
					t.Fail()
				}
			}
		})
	}
}
