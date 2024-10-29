package accnum

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AccountNumberDecorator is a custom ante handler that increments the account number 
// based on the execution mode (Simulate, CheckTx, Finalize) to avoid conflicts 
// during concurrent transactions.
type AccountNumberDecorator struct {
	ak cosmosante.AccountKeeper
}

// NewAccountNumberDecorator creates a new instance of AccountNumberDecorator.
func NewAccountNumberDecorator(ak cosmosante.AccountKeeper) AccountNumberDecorator {
	return AccountNumberDecorator{ak: ak}
}

// AnteHandle increments the account number as needed and passes control to the next AnteHandler in the chain.
func (and AccountNumberDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !ctx.IsCheckTx() && !ctx.IsReCheckTx() && !simulate {
		return next(ctx, tx, simulate)
	}

	ak, ok := and.ak.(*authkeeper.AccountKeeper)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "invalid account keeper type")
	}

	// Create a gas-free context for account operations
	gasFreeCtx := ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
	num, err := ak.AccountNumber.Peek(gasFreeCtx)
	if err != nil {
		return ctx, sdkerrors.Wrap(err, "failed to peek account number")
	}

	// Adjust the account number for simulation to avoid conflicts
	accountNumAddition := uint64(1_000_000 * (1 + boolToUint64(simulate)))
	if err := ak.AccountNumber.Set(gasFreeCtx, num+accountNumAddition); err != nil {
		return ctx, sdkerrors.Wrap(err, "failed to set account number in gas-free context")
	}

	return next(ctx, tx, simulate)
}

// Helper function for converting bool to uint64
func boolToUint64(b bool) uint64 {
	if b {
		return 1
	}
	return 0
}
