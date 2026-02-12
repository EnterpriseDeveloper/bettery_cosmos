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

func (k msgServer) CreatePartEvent(ctx context.Context, msg *types.MsgCreatePartEvent) (*types.MsgCreatePartEventResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	fmt.Print("WORK 1")

	// check if event exist
	if !k.HasCreateEvents(ctx, msg.EventId) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("event doesn't exist by id: %d", msg.EventId))
	}

	fmt.Print("WORK 2")

	// check if event not finished
	if k.GetEventFinished(ctx, msg.EventId) {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("event already finished by id: %d", msg.EventId))
	}

	fmt.Print("WORK 3")

	// TODO
	// check if user alredy part in event
	// find := k.findPartEvent(ctx, msg.EventId, msg.Creator)
	// if find {
	// 	return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("user: %s alredy participate in event by id: %d", msg.Creator, msg.EventId))
	// }

	coin, err := sdk.ParseCoinNormalized(msg.Amount.String())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid amount: %s", err.Error()))
	}

	if !coin.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("amount must be positive, got: %s", coin.String()))
	}

	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid creator address: %s", err.Error()))
	}

	// TODO check if user have enough balance for participate in event
	// TODO check coins type for participate in event
	// TODO update event data for collecting amount for each answer
	err = k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sender,
		types.ModuleName,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to send coins: %s", err.Error()))
	}

	var partPubEvents = types.Participant{
		Creator: msg.Creator,
		EventId: msg.EventId,
		Answer:  msg.Answers,
		Amount:  msg.Amount.Amount.Uint64(),
		Token:   msg.Amount.Denom,
	}

	_ = k.AppendParticipant(
		ctx,
		partPubEvents,
	)

	return &types.MsgCreatePartEventResponse{}, nil
}
