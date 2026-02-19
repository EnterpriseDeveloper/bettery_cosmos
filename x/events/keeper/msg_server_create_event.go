package keeper

import (
	"context"
	"fmt"

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
		Creator:     msg.Creator,
		Question:    msg.Question,
		Answers:     msg.Answers,
		StartTime:   uint64(timeNow),
		EndTime:     msg.EndTime,
		Category:    msg.Category,
		Status:      types.ActiveEvent,
		AnswersPool: make([]uint64, len(msg.Answers)),
	}

	if createEvent.EndTime < uint64(timeNow) {
		return nil, status.Error(codes.InvalidArgument, "end time must be in the future")
	}

	id, err := k.AppendEvent(
		ctx,
		createEvent,
	)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to append event: %v", err))
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_event",
			sdk.NewAttribute("creator", createEvent.Creator),
			sdk.NewAttribute("question", createEvent.Question),
			sdk.NewAttribute("answers", fmt.Sprintf("%d", createEvent.Answers)),
			sdk.NewAttribute("startTime", fmt.Sprintf("%d", createEvent.StartTime)),
			sdk.NewAttribute("endTime", fmt.Sprintf("%d", createEvent.EndTime)),
			sdk.NewAttribute("category", createEvent.Category),
			sdk.NewAttribute("status", createEvent.Status),
		),
	)

	return &types.MsgCreateEventResponse{Id: id}, nil
}
