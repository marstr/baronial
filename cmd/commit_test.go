package cmd

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/marstr/envelopes"
)

func Test_promptToContinue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	t.Run("affirmative", getTestAffirmativePromptReponses(ctx))
	t.Run("negative", getTestNegativePromptResponses(ctx))
	t.Run("prompt", getTestPromptText(ctx))
}

func Test_findAmount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	t.Run("deposit", getTestDepositAmount(ctx))
	t.Run("credit", getTestCreditAmount(ctx))
	t.Run("account_transfer", getTestAccountTransferAmount(ctx))
	t.Run("budget_transfer", getTestBudgetTransferAmount(ctx))
}

func getTestAffirmativePromptReponses(ctx context.Context) func(*testing.T) {
	return func(t *testing.T) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		testCases := []string{
			"y",
			"Y",
			"yes",
			"Yes",
			"YES",
			"YEs",
			" yes",
			"yes\n",
			"yes",
			"yes\r\n",
		}

		output, input := &bytes.Buffer{}, &bytes.Buffer{}
		for _, tc := range testCases {
			output.Reset()
			input.Reset()

			_, err := fmt.Fprintln(input, tc)
			if err != nil {
				t.Error(err)
				continue
			}

			result, err := promptToContinue(ctx, "want to proceed?", output, input)
			if err != nil {
				t.Error(err)
			} else if !result {
				t.Logf("returned false for: %q", tc)
				t.Fail()
			}
		}
	}
}

func getTestNegativePromptResponses(ctx context.Context) func(t *testing.T) {
	return func(t *testing.T) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		testCases := []string{
			"n",
			"N",
			"no",
			"No",
			"q",
			"Q",
			"quit",
			"QuIt",
			"",
			" ",
			"\t",
			"\r\n",
		}

		output, input := &bytes.Buffer{}, &bytes.Buffer{}

		for _, tc := range testCases {
			output.Reset()
			input.Reset()

			_, err := fmt.Fprintln(input, tc)
			if err != nil {
				t.Error(err)
				continue
			}

			result, err := promptToContinue(ctx, "want to proceed?", output, input)
			if err != nil {
				t.Error(err)
			} else if result {
				t.Logf("returned true for %q", tc)
				t.Fail()
			}
		}
	}
}

func getTestPromptText(ctx context.Context) func(*testing.T) {
	return func(t *testing.T) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		testCases := []string{
			"want to proceed?",
		}

		input, output := &bytes.Buffer{}, &bytes.Buffer{}

		for _, tc := range testCases {
			input.Reset()
			output.Reset()

			_, err := fmt.Fprintln(input)
			if err != nil {
				t.Error(err)
				continue
			}

			_, err = promptToContinue(ctx, tc, output, input)
			if err != nil {
				t.Error(err)
				continue
			}

			want := fmt.Sprintf("%s (y/N): ", strings.TrimSpace(tc))
			if got := output.String(); got != want {
				t.Logf("got:  %q\nwant: %q", got, want)
				t.Fail()
			}
		}
	}
}

func getTestDepositAmount(ctx context.Context) func(*testing.T) {
	return func(t *testing.T) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		testCases := []struct {
			Name     string
			Original envelopes.State
			Updated  envelopes.State
			Want     envelopes.Balance
		}{
			{
				Name: "simple_deposit",
				Original: envelopes.State{
					Accounts: envelopes.Accounts{
						"checking": envelopes.Balance{"USD": big.NewRat(10000, 100)},
					},
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"groceries": {
								Balance: envelopes.Balance{"USD": big.NewRat(5000, 100)},
							},
							"entertainment": {
								Balance: envelopes.Balance{"USD": big.NewRat(5000, 100)},
							},
						},
					},
				},
				Updated: envelopes.State{
					Accounts: envelopes.Accounts{
						"checking": envelopes.Balance{"USD": big.NewRat(15000, 100)},
					},
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"groceries": {
								Balance: envelopes.Balance{"USD": big.NewRat(7500, 100)},
							},
							"entertainment": {
								Balance: envelopes.Balance{"USD": big.NewRat(7500, 100)},
							},
						},
					},
				},
				Want: envelopes.Balance{"USD": big.NewRat(5000, 100)},
			},
		}

		for _, tc := range testCases {
			got := findAmount(tc.Original, tc.Updated)

			if !got.Equal(tc.Want) {
				t.Logf("%s: got: %s want: %s", tc.Name, got, tc.Want)
				t.Fail()
			}
		}
	}
}

func getTestCreditAmount(ctx context.Context) func(*testing.T) {
	return func(t *testing.T) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		testCases := []struct {
			Name     string
			Original envelopes.State
			Updated  envelopes.State
			Want     envelopes.Balance
		}{
			{
				Name: "simple_credit",
				Original: envelopes.State{
					Accounts: envelopes.Accounts{
						"checking": envelopes.Balance{"USD": big.NewRat(10000, 100)},
					},
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"groceries": {
								Balance: envelopes.Balance{"USD": big.NewRat(5000, 100)},
							},
							"entertainment": {
								Balance: envelopes.Balance{"USD": big.NewRat(5000, 100)},
							},
						},
					},
				},
				Updated: envelopes.State{
					Accounts: envelopes.Accounts{
						"checking": envelopes.Balance{"USD": big.NewRat(5000, 100)},
					},
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"groceries": {
								Balance: envelopes.Balance{"USD": big.NewRat(5000, 100)},
							},
							"entertainment": {
								Balance: envelopes.Balance{"USD": big.NewRat(0, 100)},
							},
						},
					},
				},
				Want: envelopes.Balance{"USD": big.NewRat(-5000, 100)},
			},
		}

		for _, tc := range testCases {
			got := findAmount(tc.Original, tc.Updated)

			if !got.Equal(tc.Want) {
				t.Logf("%s: got: %s want: %s", tc.Name, got, tc.Want)
				t.Fail()
			}
		}
	}
}

func getTestAccountTransferAmount(ctx context.Context) func(*testing.T) {
	return func(t *testing.T) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		testCases := []struct {
			Name     string
			Original envelopes.State
			Updated  envelopes.State
			Want     envelopes.Balance
		}{
			{
				Name: "two-party",
				Original: envelopes.State{
					Accounts: map[string]envelopes.Balance{
						"checking": {"USD": big.NewRat(10000, 100)},
						"savings":  {"USD": big.NewRat(0, 1)},
					},
				},
				Updated: envelopes.State{
					Accounts: map[string]envelopes.Balance{
						"checking": {"USD": big.NewRat(5000, 100)},
						"savings":  {"USD": big.NewRat(5000, 100)},
					},
				},
				Want: envelopes.Balance{"USD": big.NewRat(5000, 100)},
			},
			{
				Name: "three-party",
				Original: envelopes.State{
					Accounts: map[string]envelopes.Balance{
						"checking": {"USD": big.NewRat(2200000, 100)},
						"savings":  {"USD": big.NewRat(4000000, 100)},
					},
				},
				Updated: envelopes.State{
					Accounts: map[string]envelopes.Balance{
						"checking": {"USD": big.NewRat(500000, 100)},
						"savings":  {"USD": big.NewRat(0, 1)},
						"escrow":   {"USD": big.NewRat(5700000, 100)},
					},
				},
				Want: envelopes.Balance{"USD": big.NewRat(5700000, 100)},
			},
		}

		for _, tc := range testCases {
			got := findAmount(tc.Original, tc.Updated)

			if !got.Equal(tc.Want) {
				t.Logf("%s: got: %s want: %s", tc.Name, got, tc.Want)
				t.Fail()
			}
		}
	}
}

func getTestBudgetTransferAmount(ctx context.Context) func(*testing.T) {
	return func(t *testing.T) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		testCases := []struct {
			Name     string
			Original envelopes.State
			Updated  envelopes.State
			Want     envelopes.Balance
		}{
			{
				Name: "two-parties",
				Original: envelopes.State{
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"child1": {
								Balance: envelopes.Balance{"USD": big.NewRat(4590, 100)},
							},
							"child2": {
								Balance: envelopes.Balance{"USD": big.NewRat(1000, 100)},
							},
						},
					},
				},
				Updated: envelopes.State{
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"child1": {
								Balance: envelopes.Balance{"USD": big.NewRat(2250, 100)},
							},
							"child2": {
								Balance: envelopes.Balance{"USD": big.NewRat(3340, 100)},
							},
						},
					},
				},
				Want: envelopes.Balance{"USD": big.NewRat(2340, 100)},
			},
			{
				Name: "three-parties",
				Original: envelopes.State{
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"child1": {
								Balance: envelopes.Balance{"USD": big.NewRat(20000, 100)},
							},
							"child2": {
								Balance: envelopes.Balance{"USD": big.NewRat(0, 1)},
							},
							"child3": {
								Balance: envelopes.Balance{"USD": big.NewRat(0, 1)},
							},
						},
					},
				},
				Updated: envelopes.State{
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"child1": {
								Balance: envelopes.Balance{"USD": big.NewRat(10000, 100)},
							},
							"child2": {
								Balance: envelopes.Balance{"USD": big.NewRat(7500, 100)},
							},
							"child3": {
								Balance: envelopes.Balance{"USD": big.NewRat(2500, 100)},
							},
						},
					},
				},
				Want: envelopes.Balance{"USD": big.NewRat(10000, 100)},
			},
		}

		for _, tc := range testCases {
			got := findAmount(tc.Original, tc.Updated)

			if !got.Equal(tc.Want) {
				t.Logf("%s: got: %s want: %s", tc.Name, got, tc.Want)
				t.Fail()
			}
		}
	}
}
