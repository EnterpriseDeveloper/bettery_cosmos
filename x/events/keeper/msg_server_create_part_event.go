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
	exist, err := k.HasCreateEvents(ctx, msg.EventId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check event exists: %s", err.Error()))
	}
	if !exist {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("event doesn't exist by id: %d", msg.EventId))
	}

	fmt.Print("WORK 2")

	// check if event not finished
	finished, err := k.GetEventFinished(ctx, msg.EventId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check event finished: %s", err.Error()))
	}
	if finished {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("event already finished by id: %d", msg.EventId))
	}

	fmt.Print("WORK 3")

	// check if user alredy part in event
	find, err := k.findPartEvent(ctx, msg.EventId, msg.Creator)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to check participant: %s", err.Error()))
	}
	if find {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("user: %s already participate in event by id: %d", msg.Creator, msg.EventId))
	}

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

	sendAmount := msg.Amount.Amount.Uint64()

	// check if user have enough balance for participate in event
	resAmount := k.bankKeeper.GetBalance(ctx, sender, types.BetToken) // TODO check coins type for participate in event
	if sendAmount >= resAmount.Amount.Uint64() {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("user does not have enough bet token, his amount: %s", resAmount.Amount.String()))
	}

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
		Amount:  sendAmount,
		Token:   msg.Amount.Denom,
	}

	_, err = k.AppendParticipant(
		ctx,
		partPubEvents,
	)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to append participant: %v", err))
	}

	return &types.MsgCreatePartEventResponse{}, nil
}
