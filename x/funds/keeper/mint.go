package keeper

import (
	"context"
	"fmt"

	"bettery/x/funds/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) MintTokens(
	ctx context.Context,
	creator sdk.AccAddress,
	tokens sdk.Coin,
) error {
	// mint new tokens if the source of the transfer is the same chain
	if err := k.mintKeeper.MintCoins(
		ctx, types.ModuleName, sdk.NewCoins(tokens),
	); err != nil {
		return err
	}
	// send to creator
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx, types.ModuleName, creator, sdk.NewCoins(tokens),
	); err != nil {
		panic(fmt.Sprintf("unable to send coins from module to account despite previously minting coins to module account: %v", err))
	}
	return nil
}
