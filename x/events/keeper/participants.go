package keeper

import (
	"bettery/x/events/types"
	"context"
	"encoding/binary"
)

func (k Keeper) findPartEvent(
	ctx context.Context,
	eventId uint64,
	creator string,
) bool {
	store := k.storeService.OpenKVStore(ctx)
	id := k.GetParticipantCount(ctx)
	store.Get(types.ParticipantKey(id))
	// TODO
	// appendedValue := k.cdc.MustMarshal(&event)
	// store.Set(types.ParticipantKey(event.Id), appendedValue)
	// k.SetParticipantCount(ctx, id+1)

	// return event.Id
	return false
}

func (k Keeper) AppendParticipant(
	ctx context.Context,
	event types.Participant,
) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	id := k.GetParticipantCount(ctx)
	event.Id = id
	appendedValue := k.cdc.MustMarshal(&event)
	store.Set(types.ParticipantKey(event.Id), appendedValue)
	k.SetParticipantCount(ctx, id+1)

	return event.Id
}

func (k Keeper) GetParticipantCount(ctx context.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.ParticipantCountKey)
	if err != nil || bz == nil {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) SetParticipantCount(ctx context.Context, count uint64) {
	store := k.storeService.OpenKVStore(ctx)

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)

	store.Set(types.ParticipantCountKey, bz)
}
