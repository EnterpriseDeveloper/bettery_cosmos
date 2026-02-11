package types

import (
	"encoding/binary"

	"cosmossdk.io/collections"
)

const (
	// ModuleName defines the module name
	ModuleName = "events"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/gov/types/keys.go#L9
	GovModuleName = "gov"
)

// ParamsKey is the prefix to retrieve all Params
var ParamsKey = collections.NewPrefix("p_events")

var (
	EventsKeyPrefix = collections.NewPrefix("events/value/")
	EventsCountKey  = collections.NewPrefix("events/count/")
)

func EventKey(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return append(EventsKeyPrefix, bz...)
}

var (
	ParticipantKey      = collections.NewPrefix("participant/value/")
	ParticipantCountKey = collections.NewPrefix("participant/count/")
)
