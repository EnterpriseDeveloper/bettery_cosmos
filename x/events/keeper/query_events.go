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

func (q queryServer) ListEvents(ctx context.Context, req *types.QueryAllEventsRequest) (*types.QueryAllEventsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	eventss, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Events,
		req.Pagination,
		func(_ uint64, value types.Events) (types.Events, error) {
			return value, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllEventsResponse{Events: eventss, Pagination: pageRes}, nil
}

func (q queryServer) GetEvents(ctx context.Context, req *types.QueryGetEventsRequest) (*types.QueryGetEventsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	events, err := q.k.Events.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrKeyNotFound
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetEventsResponse{Events: events}, nil
}
