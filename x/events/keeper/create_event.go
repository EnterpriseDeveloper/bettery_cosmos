package keeper

import (
	"bettery/x/events/types"
	"context"
	"encoding/binary"
)

func (k Keeper) AppendEvent(
	ctx context.Context,
	event types.Events,
) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	id := k.GetEventCount(ctx)
	event.Id = id
	appendedValue := k.cdc.MustMarshal(&event)
	store.Set(GetEventByIDBytes(event.Id), appendedValue)
	k.SetEventCount(ctx, id+1)

	return event.Id
}

func (k Keeper) HasCreatePubEvents(ctx context.Context, id uint64) bool {
	store := k.storeService.OpenKVStore(ctx)
	data, err := store.Has(GetEventByIDBytes(id))
	if err != nil {
		panic(err)
	}
	return data
}

// GetEventByIDBytes returns the byte representation of the ID
func GetEventByIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

func (k Keeper) SetEventCount(ctx context.Context, count uint64) {
	store := k.storeService.OpenKVStore(ctx)

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)

	store.Set(types.EventsCountKey, bz)
}

func (k Keeper) GetEventCount(ctx context.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.EventsCountKey)
	if err != nil || bz == nil {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}
