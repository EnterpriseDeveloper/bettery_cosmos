package keeper

import (
	"context"
	"fmt"

	"bettery/x/funds/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
)

func (k msgServer) BurnToEvm(ctx context.Context, msg *types.MsgBurnToEvm) (*types.MsgBurnToEvmResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	if !common.IsHexAddress(msg.EvmRecipient) {
		return nil, errorsmod.Wrap(nil, "invalid evm address")
	}

	amount, ok := sdkmath.NewIntFromString(msg.Amount)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("parse string to init error, amount: %s,", msg.Amount))
	}

	coin := sdk.NewCoin(
		types.BetToken,
		amount,
	)

	if !coin.IsPositive() {
		return nil, errorsmod.Wrap(sdkerrors.ErrNotSupported, fmt.Sprintf("amount must be positive, got: %s", coin.String()))
	}

	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid creator address: %s", err.Error()))
	}

	sendAmount := coin.Amount.Uint64()

	// check if user have enough balance for participate in event
	resAmount := k.bankKeeper.GetBalance(ctx, sender, types.BetToken)
	if sendAmount >= resAmount.Amount.Uint64() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, fmt.Sprintf("user does not have enough bet token, his amount: %s", resAmount.Amount.String()))
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sender,
		types.ModuleName,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "SendCoinsFromAccountToModule")
	}

	err = k.bankKeeper.BurnCoins(
		ctx,
		types.ModuleName,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "Burn tokens")
	}

	diff := uint8(12)
	divisor := pow10(diff)
	normalizedAmount := coin.Amount.Mul(divisor)

	nonce, err := k.GetNextBurnNonce(ctx, msg.EvmChainId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "Nonce error")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"burn_to_evm",
			sdk.NewAttribute("chain_id", fmt.Sprint(msg.EvmChainId)),
			sdk.NewAttribute("bridge", msg.EvmBridge),
			sdk.NewAttribute("token", msg.EvmToken),
			sdk.NewAttribute("recipient", msg.EvmRecipient),
			sdk.NewAttribute("amount", normalizedAmount.String()),
			sdk.NewAttribute("nonce", fmt.Sprint(nonce)),
		),
	)

	return &types.MsgBurnToEvmResponse{}, nil
}
