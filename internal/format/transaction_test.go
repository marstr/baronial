package format

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/marstr/envelopes"
)

// A transaction with no parents (the first transaction in a repository) must print without panicking.
func TestPrettyPrintTransactionWithoutParents(t *testing.T) {
	subject := envelopes.Transaction{
		State: &envelopes.State{},
	}

	var output bytes.Buffer
	if err := PrettyPrintTransaction(context.Background(), &output, nil, subject); err != nil {
		t.Fatalf("PrettyPrintTransaction returned an error: %v", err)
	}

	if !strings.Contains(output.String(), "Impacts:") {
		t.Errorf("expected output to contain %q, got:\n%s", "Impacts:", output.String())
	}
}
