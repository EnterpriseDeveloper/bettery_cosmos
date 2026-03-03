package keeper

import (
	"context"
	"encoding/json"
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
		RoomId:      msg.RoomId,
	}

	if createEvent.EndTime < createEvent.StartTime {
		return nil, status.Error(codes.InvalidArgument, "end time must be in the future")
	}

	id, err := k.AppendEvent(
		ctx,
		createEvent,
	)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to append event: %v", err))
	}

	answersBytes, _ := json.Marshal(createEvent.Answers)
	answersPoolBytes, _ := json.Marshal(createEvent.AnswersPool)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"CREATE_EVENT",
			sdk.NewAttribute("id", fmt.Sprint(id)),
			sdk.NewAttribute("creator", createEvent.Creator),
			sdk.NewAttribute("question", createEvent.Question),
			sdk.NewAttribute("answers", string(answersBytes)),
			sdk.NewAttribute("startTime", fmt.Sprintf("%d", createEvent.StartTime)),
			sdk.NewAttribute("endTime", fmt.Sprintf("%d", createEvent.EndTime)),
			sdk.NewAttribute("category", createEvent.Category),
			sdk.NewAttribute("status", createEvent.Status),
			sdk.NewAttribute("answersPool", string(answersPoolBytes)),
			sdk.NewAttribute("roomId", createEvent.RoomId),
		),
	)

	return &types.MsgCreateEventResponse{Id: id}, nil
}
