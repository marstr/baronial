package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/marstr/envelopes"
	"github.com/marstr/envelopes/persist"
	"github.com/marstr/envelopes/persist/filesystem"

	"github.com/marstr/baronial/internal/index"
)

func TestGetDiffStatesWithEmptyRepository(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	for _, dir := range []string{index.RepoName, index.AccountsDir, index.BudgetDir} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0750); err != nil {
			t.Fatalf("create %s: %v", dir, err)
		}
	}

	repo, err := filesystem.OpenRepositoryWithCache(ctx, filepath.Join(root, index.RepoName), 10000)
	if err != nil {
		t.Fatalf("open repository: %v", err)
	}

	if err := repo.WriteBranch(ctx, persist.DefaultBranch, envelopes.ID{}); err != nil {
		t.Fatalf("create initial branch: %v", err)
	}
	if err := repo.SetCurrent(ctx, persist.DefaultBranch); err != nil {
		t.Fatalf("set current branch: %v", err)
	}

	left, right, err := getDiffStates(ctx, nil, root)
	if err != nil {
		t.Fatalf("get diff states: %v", err)
	}
	if left == nil || right == nil {
		t.Fatal("expected both diff states to be initialized")
	}
	if !left.Equal(*right) {
		t.Fatalf("expected empty index and empty HEAD to match, got index=%v and HEAD=%v", left, right)
	}
}
