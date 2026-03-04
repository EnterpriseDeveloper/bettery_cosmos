package keeper

import (
	"context"

	"bettery/x/funds/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) GetCreatorPercent(ctx context.Context, req *types.QueryGetCreatorPercentRequest) (*types.QueryGetCreatorPercentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	percent, err := q.k.GetCreatorPercent(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetCreatorPercentResponse{Percent: percent}, nil
}
