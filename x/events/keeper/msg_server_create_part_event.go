package keeper

import (
	"context"

	"bettery/x/events/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) CreatePartEvent(ctx context.Context, msg *types.MsgCreatePartEvent) (*types.MsgCreatePartEventResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgCreatePartEventResponse{}, nil
}
