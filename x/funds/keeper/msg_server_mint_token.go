package keeper

import (
	"context"
	"fmt"

	"bettery/x/funds/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) MintToken(ctx context.Context, msg *types.MsgMintToken) (*types.MsgMintTokenResponse, error) {
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

	has, err := k.HasMint(ctx, msg.Receiver)
	if err != nil {
		return nil, err
	}

	if has {
		return &types.MsgMintTokenResponse{
			Status: "exist",
		}, nil
	} else {
		receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
		if err != nil {
			return nil, err
		}
		amount, ok := sdkmath.NewIntFromString(types.Amount)
		if !ok {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("parse string to init error, amount: %s, user: %s", types.Amount, msg.Receiver))
		}

		err = k.MintTokens(ctx, receiver, sdk.NewCoin(types.BetToken, amount))
		if err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("error from burn mint, amount: %s, user: %s", types.Amount, msg.Receiver))
		}

		var mintData = types.MintEvent{
			Creator:  msg.Creator,
			Receiver: msg.Receiver,
			Amount:   types.Amount,
			Token:    types.BetToken,
			Time:     uint64(sdk.UnwrapSDKContext(ctx).BlockTime().Unix()),
		}

		k.AppendMintData(
			ctx,
			mintData,
		)

		return &types.MsgMintTokenResponse{
			Status: "done",
		}, nil
	}

}
