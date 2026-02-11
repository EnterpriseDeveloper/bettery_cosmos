package keeper

import (
	"context"

	"bettery/x/events/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) CreateEvent(ctx context.Context, msg *types.MsgCreateEvent) (*types.MsgCreateEventResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	timeNow := sdkCtx.BlockTime().Unix()

	var createEvent = types.Events{
		Creator:   msg.Creator,
		Question:  msg.Question,
		Answers:   msg.Answers,
		StartTime: uint64(timeNow),
		EndTime:   msg.EndTime,
		Category:  msg.Category,
		Status:    types.ActiveEvent,
	}

	if createEvent.EndTime < uint64(timeNow) {
		return nil, status.Error(codes.InvalidArgument, "end time must be in the future")
	}

	id := k.AppendEvent(
		ctx,
		createEvent,
	)

	return &types.MsgCreateEventResponse{Id: id}, nil
}
