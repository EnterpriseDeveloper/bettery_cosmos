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

func (q queryServer) ListValidator(ctx context.Context, req *types.QueryAllValidatorRequest) (*types.QueryAllValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	validators, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Validator,
		req.Pagination,
		func(_ uint64, value types.Validator) (types.Validator, error) {
			return value, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllValidatorResponse{Validator: validators, Pagination: pageRes}, nil
}

func (q queryServer) ListEventsForValidator(ctx context.Context, req *types.QueryAllEventsForValidatorRequest) (*types.QueryEventsForValidatorResponse, error) {
	events, err := q.k.GetEventsForValidation(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryEventsForValidatorResponse{Events: events}, nil
}

func (q queryServer) GetValidator(ctx context.Context, req *types.QueryGetValidatorRequest) (*types.QueryGetValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	validator, err := q.k.Validator.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrKeyNotFound
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetValidatorResponse{Validator: validator}, nil
}
