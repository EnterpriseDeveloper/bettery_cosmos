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

	owner, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse creator address failed")
	}

	isOwner, err := k.guardKeeper.IsOwner(ctx, owner)
	if err != nil {
		return nil, errorsmod.Wrap(err, "IsOwner err")
	}

	if !isOwner {
		return nil, errorsmod.Wrap(nil, "invalid owner")
	}

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

	// Emit event for indexers and external consumers.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"MINT_FROM_EVM",
			sdk.NewAttribute("chain_id", fmt.Sprint(msg.EvmChainId)),
			sdk.NewAttribute("bridge", msg.EvmBridge),
			sdk.NewAttribute("token", msg.EvmToken),
			sdk.NewAttribute("sender", msg.EvmSender),
			sdk.NewAttribute("recipient", msg.CosmosReceiver),
			sdk.NewAttribute("transfer_amount", msg.Amount),
			sdk.NewAttribute("cosmos_amount", coin.Amount.String()),
			sdk.NewAttribute("nonce", fmt.Sprint(msg.Nonce)),
			sdk.NewAttribute("tx_hash", msg.TxHash),
		),
	)

	return &types.MsgMintFromEvmResponse{}, nil
}
