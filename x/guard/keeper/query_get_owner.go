package keeper

import (
	"context"

	"bettery/x/guard/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) GetOwner(ctx context.Context, req *types.QueryGetOwnerRequest) (*types.QueryGetOwnerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	owner, err := q.k.GetOwner(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	if owner == nil {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetOwnerResponse{Owner: owner.String()}, nil
}
