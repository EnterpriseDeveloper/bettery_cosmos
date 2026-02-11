package keeper

import (
	"context"
	"errors"

	"bettery/x/events/types"

	"cosmossdk.io/collections"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ListParticipant(ctx context.Context, req *types.QueryAllParticipantRequest) (*types.QueryAllParticipantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	participants, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Participant,
		req.Pagination,
		func(_ uint64, value types.Participant) (types.Participant, error) {
			return value, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllParticipantResponse{Participant: participants, Pagination: pageRes}, nil
}

func (q queryServer) GetParticipant(ctx context.Context, req *types.QueryGetParticipantRequest) (*types.QueryGetParticipantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	participant, err := q.k.Participant.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrKeyNotFound
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetParticipantResponse{Participant: participant}, nil
}
