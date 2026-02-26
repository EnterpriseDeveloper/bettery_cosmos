package keeper

import (
	"context"
	"fmt"

	"bettery/x/funds/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
)

func (k msgServer) MintFromEvm(ctx context.Context, msg *types.MsgMintFromEvm) (*types.MsgMintFromEvmResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}
	// TODO: IMPORTANT Add whitelisted msg.Creator

	// TODO ADD LOGIC FOR SUPPORTED TOKEN
	// if !k.IsSupportedToken(ctx, msg.EvmToken) {
	// 	return nil, errorsmod.Wrap(nil, "unsupported token")
	// }

	exist, err := k.IsClaimProcessed(ctx, msg.EvmChainId, msg.EvmBridge, msg.Nonce) // check if claim processed
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to check claim processed")
	}

	if exist {
		return nil, errorsmod.Wrap(nil, "claim already processed")
	}

	receiver, err := sdk.AccAddressFromBech32(msg.CosmosReceiver)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid receiver address")
	}

	amount, ok := sdkmath.NewIntFromString(msg.Amount)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("parse string to init error, amount: %s,", msg.Amount))
	}

	diff := uint8(12)
	divisor := pow10(diff)
	coin := sdk.NewCoin(
		types.BetToken,
		amount.Quo(divisor),
	)

	err = k.mintKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin))
	if err != nil {
		return nil, errorsmod.Wrap(err, "unable to mint coins")
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		receiver,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unable to send coins from module to account")
	}

	k.SetClaimProcessed(ctx, msg.EvmChainId, msg.EvmBridge, msg.Nonce)

	return &types.MsgMintFromEvmResponse{}, nil
}
