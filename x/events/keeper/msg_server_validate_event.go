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

	owner, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse creator address failed")
	}

	isOwner, err := k.guardKeeper.IsOwner(ctx, owner)
	if err != nil {
		return nil, errorsmod.Wrap(err, "IsOwner err")
	}

	if !isOwner {
		return nil, errorsmod.Wrap(nil, "invalid owner")
	}

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
		EventId:   msg.EventId,
		Answer:    msg.Answers,
		Source:    msg.Source,
		CreatedAt: uint64(timeNow),
	}

	var companyFee uint64 = 0

	if msg.Answers == types.RefundEvent {
		companyFee, err = k.refundEvent(ctx, validate)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to refund event: %s, event id: %d", err.Error(), msg.EventId))
		}
	} else {
		companyFee, err = k.validateEvent(ctx, validate)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to validate event: %s, event id: %d", err.Error(), msg.EventId))
		}
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"participate_event",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("eventId", fmt.Sprintf("%d", validate.EventId)),
			sdk.NewAttribute("answer", validate.Answer),
			sdk.NewAttribute("source", validate.Source),
			sdk.NewAttribute("createdAt", fmt.Sprintf("%d", validate.CreatedAt)),
			sdk.NewAttribute("refunded", msg.Answers),
			sdk.NewAttribute("companyFee", fmt.Sprintf("%d", companyFee)),
		),
	)

	return &types.MsgValidateEventResponse{}, nil
}
