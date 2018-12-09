package cmd

import (
	"context"
	"fmt"
	"github.com/marstr/envelopes"
	"github.com/marstr/ledger/internal/budget"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var debitCmd = &cobra.Command{
	Use: `debit {budget} {amount}`,
	Short: `Removes funds from a category of spending.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
		defer cancel()

		targetDir := args[0]
		bdg, err := budget.Load(ctx, targetDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "FATAL: ", err)
			return
		}

		rawMagnitude := args[1]
		magnitude, err := envelopes.ParseAmount(rawMagnitude)
		if err != nil {
			fmt.Fprintln(os.Stderr, "FATAL: ", err)
			return
		}

		bdg = bdg.DecreaseBalance(magnitude)
		budget.Write(ctx, targetDir, bdg)
	},
}

func init() {
	rootCmd.AddCommand(debitCmd)
}