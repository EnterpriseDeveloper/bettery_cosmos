package keeper

import (
	"context"

	"bettery/x/funds/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) GetPercent(ctx context.Context, req *types.QueryGetPercentRequest) (*types.QueryGetPercentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	percent, err := q.k.GetCompanyPercent(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetPercentResponse{Percent: percent}, nil
}
