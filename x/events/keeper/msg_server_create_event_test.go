package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"bettery/x/events/keeper"
	"bettery/x/events/types"
)

// TestMsgServerCreateEvent_EmitEvent verifies that routing a MsgCreateEvent
// through the real MsgServer:
//   - successfully creates and stores the event, and
//   - emits the expected CREATE_EVENT SDK event with all attributes set.
func TestMsgServerCreateEvent_EmitEvent(t *testing.T) {
	f := initFixture(t)

	// Sanity check: with the fixed block time in the fixture this should be 0.
	require.Equal(t, int64(0), f.ctx.BlockTime().Unix())

	// Prepare a valid creator address using the same address codec as the keeper.
	addrBytes := make([]byte, 20)
	copy(addrBytes, []byte("creator_address_1234"))

	creator, err := f.addressCodec.BytesToString(addrBytes)
	require.NoError(t, err)

	msg := &types.MsgCreateEvent{
		Creator:  creator,
		Question: "Will BTC price go up tomorrow?",
		Answers:  []string{"YES", "NO"},
		// EndTime must be >= StartTime; with block time fixed at 0 in the test
		// fixture, any positive value is valid here.
		EndTime:  10,
		Category: "crypto",
		RoomId:   "room-1",
	}

	server := keeper.NewMsgServerImpl(f.keeper)

	resp, err := server.CreateEvent(f.ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify the event was stored with the returned ID.
	storedEvent, err := f.keeper.GetEventById(f.ctx, resp.Id)
	require.NoError(t, err)
	require.Equal(t, msg.Creator, storedEvent.Creator)
	require.Equal(t, msg.Question, storedEvent.Question)
	require.Equal(t, msg.Answers, storedEvent.Answers)
	require.Equal(t, msg.EndTime, storedEvent.EndTime)
	require.Equal(t, msg.Category, storedEvent.Category)
	require.Equal(t, types.ActiveEvent, storedEvent.Status)
	require.Equal(t, msg.RoomId, storedEvent.RoomId)
	require.Len(t, storedEvent.AnswersPool, len(msg.Answers))

	// Now verify that the CREATE_EVENT SDK event was emitted with the expected attributes.
	events := f.ctx.EventManager().Events()
	require.NotEmpty(t, events, "no events were emitted")

	var found types.Events
	var foundEvent bool

	for _, ev := range events {
		if ev.Type != "CREATE_EVENT" {
			continue
		}

		foundEvent = true

		attrMap := make(map[string]string, len(ev.Attributes))
		for _, a := range ev.Attributes {
			attrMap[string(a.Key)] = string(a.Value)
		}

		// Basic attributes.
		require.Equal(t, fmt.Sprint(resp.Id), attrMap["id"])
		require.Equal(t, storedEvent.Creator, attrMap["creator"])
		require.Equal(t, storedEvent.Question, attrMap["question"])
		require.Equal(t, storedEvent.Category, attrMap["category"])
		require.Equal(t, storedEvent.Status, attrMap["status"])
		require.Equal(t, storedEvent.RoomId, attrMap["roomId"])

		// Time-related attributes.
		require.Equal(t, fmt.Sprintf("%d", storedEvent.StartTime), attrMap["startTime"])
		require.Equal(t, fmt.Sprintf("%d", storedEvent.EndTime), attrMap["endTime"])

		// JSON-encoded answers and answersPool.
		expectedAnswersJSON, err := json.Marshal(storedEvent.Answers)
		require.NoError(t, err)
		require.Equal(t, string(expectedAnswersJSON), attrMap["answers"])

		expectedAnswersPoolJSON, err := json.Marshal(storedEvent.AnswersPool)
		require.NoError(t, err)
		require.Equal(t, string(expectedAnswersPoolJSON), attrMap["answersPool"])

		found = storedEvent
		break
	}

	require.True(t, foundEvent, "CREATE_EVENT was not emitted")
	_ = found // reserved for potential future assertions
}
