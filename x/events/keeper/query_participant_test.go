package keeper_test

import (
	"context"
	"strconv"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"bettery/x/events/keeper"
	"bettery/x/events/types"
)

func createNParticipant(keeper keeper.Keeper, ctx context.Context, n int) []types.Participant {
	items := make([]types.Participant, n)
	for i := range items {
		iu := uint64(i)
		items[i].Id = iu
		items[i].EventId = uint64(i)
		items[i].Answer = strconv.Itoa(i)
		items[i].Amount = uint64(i)
		items[i].Token = strconv.Itoa(i)
		items[i].Result = uint64(i)
		_ = keeper.Participant.Set(ctx, iu, items[i])
		_ = keeper.ParticipantSeq.Set(ctx, iu)
	}
	return items
}

func TestParticipantQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNParticipant(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetParticipantRequest
		response *types.QueryGetParticipantResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetParticipantRequest{Id: msgs[0].Id},
			response: &types.QueryGetParticipantResponse{Participant: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetParticipantRequest{Id: msgs[1].Id},
			response: &types.QueryGetParticipantResponse{Participant: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetParticipantRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetParticipant(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestParticipantQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNParticipant(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllParticipantRequest {
		return &types.QueryAllParticipantRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListParticipant(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Participant), step)
			require.Subset(t, msgs, resp.Participant)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListParticipant(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Participant), step)
			require.Subset(t, msgs, resp.Participant)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListParticipant(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Participant)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListParticipant(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
