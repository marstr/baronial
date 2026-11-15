package cmd

import (
	"bytes"
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/marstr/envelopes"
)

func Test_writeBudgetBalancesDepth(t *testing.T) {
	budget := envelopes.Budget{
		Balance: envelopes.Balance{"USD": big.NewRat(1, 1)},
		Children: map[string]*envelopes.Budget{
			"food": {
				Balance: envelopes.Balance{"USD": big.NewRat(2, 1)},
				Children: map[string]*envelopes.Budget{
					"groceries": {Balance: envelopes.Balance{"USD": big.NewRat(3, 1)}},
				},
			},
		},
	}

	t.Run("depth1", func(t *testing.T) {
		var out bytes.Buffer
		if err := writeBudgetBalances(context.Background(), &out, budget, 1); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "\tfood: ") {
			t.Errorf("expected direct child listed, got:\n%s", out.String())
		}
		if strings.Contains(out.String(), "groceries") {
			t.Errorf("expected grandchild hidden at depth 1, got:\n%s", out.String())
		}
	})

	t.Run("depth2", func(t *testing.T) {
		var out bytes.Buffer
		if err := writeBudgetBalances(context.Background(), &out, budget, 2); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "\t\tgroceries: ") {
			t.Errorf("expected grandchild listed at depth 2, got:\n%s", out.String())
		}
	})
}
