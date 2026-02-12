package keeper

import (
	"context"
	"fmt"

	"bettery/x/events/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ValidateEvent(ctx context.Context, msg *types.MsgValidateEvent) (*types.MsgValidateEventResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: check user wallet. Only owner can execute this action.

	// check if event exist
	exist, err := k.HasCreateEvents(ctx, msg.EventId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check event exists: %s, event id: %d", err.Error(), msg.EventId))
	}
	if !exist {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("event doesn't exist by id: %d", msg.EventId))
	}

	// check if event not finished
	finished, err := k.GetEventFinished(ctx, msg.EventId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check event finished: %s, event id: %d", err.Error(), msg.EventId))
	}
	if finished {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("event already finished by id: %d", msg.EventId))
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	timeNow := sdkCtx.BlockTime().Unix()

	// check if validator can validate event, validator can validate only finished event
	endTime, err := k.getEventEndTime(ctx, msg.EventId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get event end time: %s, event id: %d", err.Error(), msg.EventId))
	}

	if endTime >= uint64(timeNow) {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("event not finished by id: %d", msg.EventId))
	}

	validate := types.Validator{
		EventId: msg.EventId,
		Answer:  msg.Answers,
		Source:  msg.Source,
	}

	if msg.Answers == types.RefundEvent {
		_, err := k.refundEvent(ctx, validate)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to refund event: %s, event id: %d", err.Error(), msg.EventId))
		}
	} else {
		_, err := k.validateEvent(ctx, validate)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to validate event: %s, event id: %d", err.Error(), msg.EventId))
		}
	}

	return &types.MsgValidateEventResponse{}, nil
}
