package keeper

import (
	"context"

	"bettery/x/funds/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) SetOwner(ctx context.Context, msg *types.MsgSetOwner) (*types.MsgSetOwnerResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgSetOwnerResponse{}, nil
}
