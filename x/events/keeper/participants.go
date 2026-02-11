package keeper

import (
	"bettery/x/events/types"
	"context"
)

func (k Keeper) findPartEvent(
	ctx context.Context,
	eventId uint64,
	creator string,
) bool {
	store := k.storeService.OpenKVStore(ctx)
	id := k.GetParticipantCount(ctx)
	event.Id = id
	appendedValue := k.cdc.MustMarshal(&event)
	store.Set(types.ParticipantKey(event.Id), appendedValue)
	k.SetParticipantCount(ctx, id+1)

	return event.Id
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
