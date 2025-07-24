package cmd

import (
	"github.com/marstr/baronial/internal/index"
	"github.com/marstr/envelopes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	bringToImmediateFlag      = "immediate"
	bringToImmediateShorthand = "i"
	bringToImmediateDefault   = false
	bringToImmediateUsage     = "Use the balance of the specified destination budget, but don't include its children."
)

var bringToCmd = &cobra.Command{
	Use:     "bring-to {balance} {src} {dest}",
	Aliases: []string{"bring", "br"},
	Short:   "Bring a budget to a given balance by transferring funds from another.",
	Args:    cobra.ExactArgs(3),
	RunE:    RunBringTo,
}

func init() {
	rootCmd.AddCommand(bringToCmd)
	bringToCmd.Flags().BoolP(bringToImmediateFlag, bringToImmediateShorthand, bringToImmediateDefault, bringToImmediateUsage)
}

func RunBringTo(cmd *cobra.Command, args []string) error {
	ctx, cancel := RootContext(cmd)
	defer cancel()

	desiredBal, err := envelopes.ParseBalance([]byte(args[0]))
	if err != nil {
		return err
	}

	srcPath := args[1]
	src, err := index.LoadBudget(ctx, srcPath)
	if err != nil {
		return err
	}

	destPath := args[2]
	dest, err := index.LoadBudget(ctx, destPath)
	if err != nil {
		return err
	}

	var currentBalance envelopes.Balance
	var useImmediate bool
	useImmediate, err = cmd.Flags().GetBool(bringToImmediateFlag)
	if err != nil {
		logrus.Fatal(err)
	}

	if useImmediate {
		currentBalance = dest.Balance
	} else {
		currentBalance = dest.RecursiveBalance()
	}

	amountToTransfer := desiredBal.Sub(currentBalance)

	src.Balance = src.Balance.Sub(amountToTransfer)
	dest.Balance = dest.Balance.Add(amountToTransfer)

	err = index.WriteBudget(ctx, srcPath, *src)
	if err != nil {
		return err
	}

	err = index.WriteBudget(ctx, destPath, *dest)
	if err != nil {
		return err
	}

	return nil
}
