package cmd

import (
	"bytes"
	"context"
	"fmt"
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

			fmt.Fprintln(input)

			_, err := promptToContinue(ctx, tc, output, input)
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
						"checking": 10000,
					},
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"groceries": {
								Balance: 5000,
							},
							"entertainment": {
								Balance: 5000,
							},
						},
					},
				},
				Updated: envelopes.State{
					Accounts: envelopes.Accounts{
						"checking": 15000,
					},
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"groceries": {
								Balance: 7500,
							},
							"entertainment": {
								Balance: 7500,
							},
						},
					},
				},
				Want: 5000,
			},
		}

		for _, tc := range testCases {
			got := findAmount(tc.Original, tc.Updated)

			if got != tc.Want {
				t.Logf("%s: got: %d want: %d", tc.Name, got, tc.Want)
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
						"checking": 10000,
					},
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"groceries": {
								Balance: 5000,
							},
							"entertainment": {
								Balance: 5000,
							},
						},
					},
				},
				Updated: envelopes.State{
					Accounts: envelopes.Accounts{
						"checking": 5000,
					},
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"groceries": {
								Balance: 5000,
							},
							"entertainment": {
								Balance: 0,
							},
						},
					},
				},
				Want: -5000,
			},
		}

		for _, tc := range testCases {
			got := findAmount(tc.Original, tc.Updated)

			if got != tc.Want {
				t.Logf("%s: got: %d want: %d", tc.Name, got, tc.Want)
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
						"checking": 10000,
						"savings":  0,
					},
				},
				Updated: envelopes.State{
					Accounts: map[string]envelopes.Balance{
						"checking": 5000,
						"savings":  5000,
					},
				},
				Want: 5000,
			},
			{
				Name: "three-party",
				Original: envelopes.State{
					Accounts: map[string]envelopes.Balance{
						"checking": 2200000,
						"savings":  4000000,
					},
				},
				Updated: envelopes.State{
					Accounts: map[string]envelopes.Balance{
						"checking": 500000,
						"savings":  0,
						"escrow":   5700000,
					},
				},
				Want: 5700000,
			},
		}

		for _, tc := range testCases {
			got := findAmount(tc.Original, tc.Updated)

			if got != tc.Want {
				t.Logf("%s: got: %d want: %d", tc.Name, got, tc.Want)
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
								Balance: 4590,
							},
							"child2": {
								Balance: 1000,
							},
						},
					},
				},
				Updated: envelopes.State{
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"child1": {
								Balance: 2250,
							},
							"child2": {
								Balance: 3340,
							},
						},
					},
				},
				Want: 2340,
			},
			{
				Name: "three-parties",
				Original: envelopes.State{
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"child1": {
								Balance: 20000,
							},
							"child2": {
								Balance: 0,
							},
							"child3": {
								Balance: 0,
							},
						},
					},
				},
				Updated: envelopes.State{
					Budget: &envelopes.Budget{
						Children: map[string]*envelopes.Budget{
							"child1": {
								Balance: 10000,
							},
							"child2": {
								Balance: 7500,
							},
							"child3": {
								Balance: 2500,
							},
						},
					},
				},
				Want: 10000,
			},
		}

		for _, tc := range testCases {
			got := findAmount(tc.Original, tc.Updated)

			if got != tc.Want {
				t.Logf("%s: got: %d want: %d", tc.Name, got, tc.Want)
				t.Fail()
			}
		}
	}
}
