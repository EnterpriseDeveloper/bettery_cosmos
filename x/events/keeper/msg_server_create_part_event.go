package keeper

import (
	"context"
	"fmt"
	"math/big"

	"bettery/x/events/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
)

func (k msgServer) CreatePartEvent(ctx context.Context, msg *types.MsgCreatePartEvent) (*types.MsgCreatePartEventResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	fmt.Print("WORK")

	// check if event not finished
	if k.GetEventFinished(ctx, msg.PubId) {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("event already finished by id: %d", msg.PubId))
	}

	fmt.Print("WORK 2")

	// check if event exist
	if !k.HasCreatePubEvents(ctx, msg.PubId) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("event doesn't exist by id: %d", msg.PubId))
	}

	fmt.Print("WORK 3")

	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	// check if user alredy part in event
	find := k.findPartPubEvent(ctx, msg.PubId, msg.Creator)
	if find {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("user: %s alredy participate in event by id: %d", msg.Creator, msg.PubId))
	}

	// find answer index
	answerIndex := k.GetAnswerIndex(ctx, msg.PubId, msg.Answers)
	if answerIndex == -1 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("answer %s not found in event by id: %d", msg.Answers, msg.PubId))
	}

	// check balance of user
	sendAmount, ok := new(big.Int).SetString(msg.Amount, 0)
	if !ok {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("parse big init error, amount: %s, user: %s", msg.Creator, msg.Amount))
	}

	resAmount := k.bankKeeper.GetBalance(ctx, sender, types.BetToken)
	if sendAmount.Cmp(resAmount.Amount.BigInt()) == 1 {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("user does not have enought bet token, his amount: %s", resAmount.Amount.String()))
	}
	// send money to the event
	betAmount, ok := math.NewIntFromString(msg.Amount)
	if !ok {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("parse string to init error, amount: %s, user: %s", msg.Amount, msg.Creator))
	}
	err = k.TransferToModule(ctx, sender, sdk.NewCoin(types.BetToken, betAmount))
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("send bet token to module error, amount: %s", err.Error()))
	}

	var partPubEvents = types.MsgCreatePartPubEvents{
		Creator:     msg.Creator,
		PubId:       msg.PubId,
		Answers:     msg.Answers,
		Amount:      msg.Amount,
		AnswerIndex: uint32(answerIndex),
	}

	id := k.AppendPartPubEvents(
		ctx,
		partPubEvents,
	)

	return &types.MsgCreatePartEventResponse{}, nil
}
