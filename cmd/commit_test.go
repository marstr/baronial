package cmd

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func Test_promptToContinue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	t.Run("affirmative", getTestAffirmativePromptReponses(ctx))
	t.Run("negative", getTestNegativePromptResponses(ctx))
	t.Run("prompt", getTestPromptText(ctx))
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
