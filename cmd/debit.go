package cmd

import (
	"context"
	"time"

	"github.com/marstr/envelopes"
	"github.com/marstr/ledger/internal/budget"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
			logrus.Fatal(err)
		}

		rawMagnitude := args[1]
		magnitude, err := envelopes.ParseAmount(rawMagnitude)
		if err != nil {
			logrus.Fatal(err)
		}

		bdg = bdg.DecreaseBalance(magnitude)
		budget.Write(ctx, targetDir, bdg)
	},
}

func init() {
	rootCmd.AddCommand(debitCmd)
}