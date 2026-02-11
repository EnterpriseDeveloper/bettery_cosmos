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

func createNEvents(keeper keeper.Keeper, ctx context.Context, n int) []types.Events {
	items := make([]types.Events, n)
	for i := range items {
		iu := uint64(i)
		items[i].Id = iu
		items[i].Creator = strconv.Itoa(i)
		items[i].Question = strconv.Itoa(i)
		items[i].EndTime = uint64(i)
		items[i].StartTime = uint64(i)
		items[i].Category = strconv.Itoa(i)
		items[i].Status = strconv.Itoa(i)
		items[i].TotalPool = uint64(i)
		items[i].WinningAnswer = strconv.Itoa(i)
		items[i].AnswerSource = strconv.Itoa(i)
		_ = keeper.Events.Set(ctx, iu, items[i])
		_ = keeper.EventsSeq.Set(ctx, iu)
	}
	return items
}

func TestEventsQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNEvents(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetEventsRequest
		response *types.QueryGetEventsResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetEventsRequest{Id: msgs[0].Id},
			response: &types.QueryGetEventsResponse{Events: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetEventsRequest{Id: msgs[1].Id},
			response: &types.QueryGetEventsResponse{Events: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetEventsRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetEvents(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestEventsQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNEvents(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllEventsRequest {
		return &types.QueryAllEventsRequest{
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
			resp, err := qs.ListEvents(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Events), step)
			require.Subset(t, msgs, resp.Events)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListEvents(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Events), step)
			require.Subset(t, msgs, resp.Events)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListEvents(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Events)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListEvents(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
