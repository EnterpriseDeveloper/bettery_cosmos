package keeper

import (
	"context"

	"bettery/x/events/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) ValidateEvent(ctx context.Context, msg *types.MsgValidateEvent) (*types.MsgValidateEventResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgValidateEventResponse{}, nil
}
