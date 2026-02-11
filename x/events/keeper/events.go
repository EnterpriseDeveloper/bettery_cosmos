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
	store.Set(types.EventKey(event.Id), appendedValue)
	k.SetEventCount(ctx, id+1)

	return event.Id
}

func (k Keeper) HasCreateEvents(ctx context.Context, id uint64) bool {
	store := k.storeService.OpenKVStore(ctx)
	data, err := store.Has(types.EventKey(id))
	if err != nil {
		panic(err)
	}
	return data
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

// check if event finished
func (k Keeper) GetEventFinished(ctx context.Context, id uint64) bool {
	event := k.GetEventById(ctx, id).Status
	return event == types.FinishedEvent || event == types.RefundEvent
}

func (k Keeper) GetEventById(ctx context.Context, id uint64) types.Events {
	store := k.storeService.OpenKVStore(ctx)
	var event types.Events
	data, err := store.Get(types.EventKey(id))
	if err != nil {
		panic(err)
	}
	k.cdc.MustUnmarshal(data, &event)
	return event
}
