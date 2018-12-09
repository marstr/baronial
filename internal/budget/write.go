package budget

import (
	"context"
	"fmt"
	"github.com/marstr/envelopes"
	"os"
	"path/filepath"
)

func Write(ctx context.Context, targetDir string, budget envelopes.Budget) error {
	targetFile := filepath.Join(targetDir, cashName)
	handle, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer handle.Close()

	payload := envelopes.FormatAmount(budget.Balance())

	_, err = fmt.Fprintln(handle, payload)
	return nil
}
