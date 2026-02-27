package keeper

import (
	"context"

	"bettery/x/guard/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) ChangeOwner(ctx context.Context, msg *types.MsgChangeOwner) (*types.MsgChangeOwnerResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgChangeOwnerResponse{}, nil
}
