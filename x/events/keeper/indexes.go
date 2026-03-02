package keeper

import (
	"bettery/x/events/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
)

type ParticipantIndexes struct {
	EventId *indexes.Multi[
		uint64, // index key: event_id
		uint64, // primary key
		types.Participant,
	]
}

func (i ParticipantIndexes) IndexesList() []collections.Index[uint64, types.Participant] {
	return []collections.Index[uint64, types.Participant]{i.EventId}
}
