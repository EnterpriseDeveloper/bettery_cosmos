package keeper

import (
	"bettery/x/events/types"
	"context"
)

func (k Keeper) AppendCreatePubEvents(
	ctx context.Context,
	createPubEvents types.MsgCreateEvent,
) string {
	store := k.storeService.OpenKVStore(ctx)
	appendedValue := k.cdc.MustMarshal(&createPubEvents)
	store.Set(GetCreatePubEventsIDBytes(createPubEvents.Id), appendedValue)

	return createPubEvents.Id
}

func (k Keeper) HasCreatePubEvents(ctx context.Context, id string) bool {
	store := k.storeService.OpenKVStore(ctx)
	data, err := store.Has(GetCreatePubEventsIDBytes(id))
	if err != nil {
		panic(err)
	}
	return data
}

// GetCreatePubEventsIDBytes returns the byte representation of the ID
func GetCreatePubEventsIDBytes(id string) []byte {
	return []byte(id)
}
