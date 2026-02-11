package keeper

import (
	"context"

	"bettery/x/events/types"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) CreateEvent(ctx context.Context, msg *types.MsgCreateEvent) (*types.MsgCreateEventResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	var createPubEvents = types.MsgCreateEvent{
		Creator:  msg.Creator,
		Id:       msg.Id,
		Question: msg.Question,
		Answers:  msg.Answers,
		EndTime:  msg.EndTime,
		Category: msg.Category,
	}

	// check if event exist
	if k.HasCreatePubEvents(ctx, msg.Id) {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("event by id %d alredy exist", msg.Id))
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	timeNow := sdkCtx.BlockTime().Unix()
	if createPubEvents.EndTime < uint64(timeNow) {
		return nil, status.Error(codes.InvalidArgument, "end time must be in the future")
	}

	id := k.AppendCreatePubEvents(
		ctx,
		createPubEvents,
	)

	return &types.MsgCreateEventResponse{Id: id}, nil
}
