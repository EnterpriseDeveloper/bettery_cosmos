package keeper

import (
	"bettery/x/events/types"
	"context"
	"encoding/binary"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) AppendEvent(
	ctx context.Context,
	event types.Events,
) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	id, err := k.GetEventCount(ctx)
	if err != nil {
		return 0, err
	}
	event.Id = id
	appendedValue := k.cdc.MustMarshal(&event)
	store.Set(types.EventKey(event.Id), appendedValue)
	k.SetEventCount(ctx, id+1)

	return event.Id, nil
}

func (k Keeper) HasCreateEvents(ctx context.Context, id uint64) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	data, err := store.Has(types.EventKey(id))
	if err != nil {
		return false, err
	}
	return data, nil
}

func (k Keeper) SetEventCount(ctx context.Context, count uint64) {
	store := k.storeService.OpenKVStore(ctx)

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)

	store.Set(types.EventsCountKey, bz)
}

func (k Keeper) GetEventCount(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.EventsCountKey)
	if err != nil || bz == nil {
		return 0, err
	}
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(bz), nil
}

// check if event finished
func (k Keeper) GetEventFinished(ctx context.Context, id uint64) (bool, error) {
	event, err := k.GetEventById(ctx, id)
	if err != nil {
		return false, err
	}
	return event.Status == types.FinishedEvent || event.Status == types.RefundEvent, nil
}

func (k Keeper) GetEventById(ctx context.Context, id uint64) (types.Events, error) {
	store := k.storeService.OpenKVStore(ctx)
	var event types.Events
	data, err := store.Get(types.EventKey(id))
	if err != nil {
		return event, err
	}
	k.cdc.MustUnmarshal(data, &event)
	return event, nil
}

func (k Keeper) updateEvent(ctx context.Context, participant types.Participant) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	var event types.Events
	data, err := store.Get(types.EventKey(participant.EventId))
	if err != nil {
		return false, err
	}
	k.cdc.MustUnmarshal(data, &event)
	event.Participants = append(event.Participants, participant.Creator)
	answerIndex := indexOf(event.Answers, participant.Answer)
	if answerIndex == -1 {
		return false, status.Error(codes.InvalidArgument, fmt.Sprintf("answer: %s not found in event by id: %d", participant.Answer, participant.EventId))
	} else {
		event.AnswersPool[answerIndex] += participant.Amount
		appendedValue := k.cdc.MustMarshal(&event)
		store.Set(types.EventKey(event.Id), appendedValue)
		return true, nil
	}

}

func indexOf(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1 // not found
}
