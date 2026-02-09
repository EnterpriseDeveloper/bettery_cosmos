package keeper

import (
	"bettery/x/events/types"
	"context"
	"encoding/binary"
)

func (k Keeper) AppendCreatePubEvents(
	ctx context.Context,
	createPubEvents types.MsgCreateEvent,
) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	appendedValue := k.cdc.MustMarshal(&createPubEvents)
	store.Set(GetCreatePubEventsIDBytes(createPubEvents.Id), appendedValue)

	return createPubEvents.Id
}

func (k Keeper) HasCreatePubEvents(ctx context.Context, id uint64) bool {
	store := k.storeService.OpenKVStore(ctx)
	data, err := store.Has(GetCreatePubEventsIDBytes(id))
	if err != nil {
		panic(err)
	}
	return data
}

// GetCreatePubEventsIDBytes returns the byte representation of the ID
func GetCreatePubEventsIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}
