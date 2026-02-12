package keeper

import (
	"bettery/x/events/types"
	"context"
	"encoding/binary"
	"slices"
)

func (k Keeper) findPartEvent(
	ctx context.Context,
	eventId uint64,
	creator string,
) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	var event types.Events
	data, err := store.Get(types.EventKey(eventId))
	if err != nil {
		return false, err
	}
	k.cdc.MustUnmarshal(data, &event)
	if slices.Contains(event.Participants, creator) {
		return true, nil
	}
	return false, nil
}

func (k Keeper) AppendParticipant(
	ctx context.Context,
	event types.Participant,
) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	id, err := k.GetParticipantCount(ctx)
	if err != nil {
		return 0, err
	}
	event.Id = id
	appendedValue := k.cdc.MustMarshal(&event)
	store.Set(types.ParticipantKey(event.Id), appendedValue)
	k.SetParticipantCount(ctx, id+1)

	_, err = k.updateEventFromParticipant(ctx, event)
	if err != nil {
		return 0, err
	}
	return event.Id, nil

}

func (k Keeper) GetParticipantCount(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.ParticipantCountKey)
	if err != nil || bz == nil {
		return 0, err
	}
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(bz), nil
}

func (k Keeper) SetParticipantCount(ctx context.Context, count uint64) {
	store := k.storeService.OpenKVStore(ctx)

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)

	store.Set(types.ParticipantCountKey, bz)
}

func (k Keeper) GetParticipantsByEvent(
	ctx context.Context,
	eventId uint64,
	answer string,
) ([]types.Participant, []types.Participant, uint64, uint64, error) {

	var allUsers []types.Participant
	var winUsers []types.Participant
	totalPool := uint64(0)
	winnersPool := uint64(0)

	// TODO think about index for participants by event id for optimization
	err := k.Participant.Walk(
		ctx,
		nil,
		func(_ uint64, p types.Participant) (bool, error) {
			if p.EventId == eventId {
				totalPool += p.Amount
				allUsers = append(allUsers, p)
				if p.Answer == answer {
					winUsers = append(winUsers, p)
					winnersPool += p.Amount
				}
			}
			return false, nil
		},
	)

	if err != nil {
		return nil, nil, 0, 0, err
	}

	return allUsers, winUsers, totalPool, winnersPool, nil
}
