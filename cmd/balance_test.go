package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marstr/baronial/internal/diagnostics"
	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/spf13/cobra"
)

func TestBalanceClassificationProbesAreDebugOnly(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, index.RepoName), 0700); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"budget/grocery", "accounts/checking"} {
		path := filepath.Join(root, target)
		if err := os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "cash.txt"), []byte("USD 1\n"), 0600); err != nil {
			t.Fatal(err)
		}
		for _, level := range []slog.Level{slog.LevelInfo, slog.LevelDebug} {
			t.Run(target+"/"+level.String(), func(t *testing.T) {
				var output, stderr, logs bytes.Buffer
				logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: level}))
				cmd := &cobra.Command{Use: "balance"}
				cmd.Flags().Duration(timeoutFlag, 0, "")
				cmd.Flags().Uint8(balanceDepthFlag, balanceDepthDefault, "")
				cmd.SetContext(diagnostics.WithLogger(context.Background(), logger))
				cmd.SetOut(&output)
				cmd.SetErr(&stderr)
				if err := balanceCmd.RunE(cmd, []string{path}); err != nil {
					t.Fatal(err)
				}
				if output.Len() == 0 || stderr.Len() != 0 {
					t.Fatalf("classification affected user output: %q, %q", output.String(), stderr.String())
				}
				probes := 0
				decoder := json.NewDecoder(&logs)
				for decoder.More() {
					var record struct {
						Level   string `json:"level"`
						Message string `json:"msg"`
					}
					if err := decoder.Decode(&record); err != nil {
						t.Fatal(err)
					}
					if strings.HasPrefix(record.Message, "target is not ") {
						probes++
						if record.Level != "DEBUG" {
							t.Fatalf("classification probe logged at %s", record.Level)
						}
					}
				}
				expected := 0
				if level == slog.LevelDebug {
					expected = 1
				}
				if probes != expected {
					t.Fatalf("got %d classification probes at %s, want %d", probes, level, expected)
				}
			})
		}
	}
}

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
