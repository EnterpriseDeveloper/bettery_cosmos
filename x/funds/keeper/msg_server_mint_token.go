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

	has, err := k.HasMint(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}

	if has {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("User already minted tokens"))
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}
	amount, ok := sdkmath.NewIntFromString(types.Amount)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("parse string to init error, amount: %s, user: %s", types.Amount, msg.Creator))
	}

	err = k.MintTokens(ctx, receiver, sdk.NewCoin(types.BetToken, amount))
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("error from burn mint, amount: %s, user: %s", types.Amount, msg.Creator))
	}

	var mintData = types.MintEvent{
		Creator: msg.Creator,
		Amount:  types.Amount,
		Token:   types.BetToken,
		Time:    uint64(sdk.UnwrapSDKContext(ctx).BlockTime().Unix()),
	}

	k.AppendMintData(
		ctx,
		mintData,
	)

	return &types.MsgMintTokenResponse{
		Status: "done",
	}, nil
}
