package keeper

import (
	"context"

	"bettery/x/guard/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) ChangeOwner(ctx context.Context, msg *types.MsgChangeOwner) (*types.MsgChangeOwnerResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	newOwner, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse new owner address failed")
	}

	owner, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse creator address failed")
	}

	isOwner, err := k.IsOwner(ctx, owner)
	if err != nil {
		return nil, errorsmod.Wrap(err, "IsOwner err")
	}

	if !isOwner {
		return nil, errorsmod.Wrap(nil, "invalid owner")
	} else {
		k.SetNewOwner(ctx, newOwner)
	}

	return &types.MsgChangeOwnerResponse{}, nil
}
