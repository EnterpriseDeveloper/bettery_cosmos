package keeper

import (
	"context"

	"bettery/x/funds/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SetCompanyPercent(ctx context.Context, msg *types.MsgSetCompanyPercent) (*types.MsgSetCompanyPercentResponse, error) {
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

	k.SetCompanyPercentStore(ctx, msg.Percent)

	return &types.MsgSetCompanyPercentResponse{}, nil
}
