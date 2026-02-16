package keeper

import (
	"context"

	"bettery/x/events/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: Move to DB for future optimization if needed
func (q queryServer) ParticipantById(ctx context.Context, req *types.QueryParticipantByIdRequest) (*types.QueryParticipantByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var found *types.Participant

	err := q.k.Participant.Walk(
		ctx,
		nil,
		func(_ uint64, p types.Participant) (bool, error) {

			if p.EventId == req.EventId && p.Creator == req.Creator {
				found = &p
				return true, nil // stop iteration
			}

			return false, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if found == nil {
		return nil, status.Error(
			codes.NotFound,
			"participant not found",
		)
	}

	return &types.QueryParticipantByIdResponse{Participant: found}, nil
}
