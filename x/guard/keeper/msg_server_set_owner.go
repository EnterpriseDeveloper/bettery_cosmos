package keeper

import (
	"context"

	"bettery/x/guard/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) SetOwner(ctx context.Context, msg *types.MsgSetOwner) (*types.MsgSetOwnerResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse creator address failed")
	}

	owner, err := k.GetOwner(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "GetOwner failed")
	}
	if owner == nil {
		k.SetNewOwner(ctx, creator)
	}

	return &types.MsgSetOwnerResponse{}, nil
}
