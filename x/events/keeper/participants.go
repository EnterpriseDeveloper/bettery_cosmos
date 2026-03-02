package keeper

import (
	"bettery/x/events/types"
	"context"
	"encoding/binary"
	"errors"
	"slices"

	"cosmossdk.io/collections"
)

func (k Keeper) findPartEvent(
	ctx context.Context,
	eventId uint64,
	creator string,
) (bool, error) {
	event, err := k.Events.Get(ctx, eventId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return slices.Contains(event.Participants, creator), nil
}

func (k Keeper) AppendParticipant(
	ctx context.Context,
	event types.Participant,
) (uint64, error) {
	id, err := k.ParticipantSeq.Next(ctx)
	if err != nil {
		return 0, err
	}
	event.Id = id
	if err := k.Participant.Set(ctx, id, event); err != nil {
		return 0, err
	}

	_, err = k.updateEventFromParticipant(ctx, event)
	if err != nil {
		return 0, err
	}
	return event.Id, nil

}

func (k Keeper) updateParticipantFromValidator(ctx context.Context, participant types.Participant, amount uint64) (bool, error) {
	participant.Result = amount
	if err := k.Participant.Set(ctx, participant.Id, participant); err != nil {
		return false, err
	}
	return true, nil
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

func (k Keeper) GetParticipantsByEventWithIndex(
	ctx context.Context,
	eventId uint64,
	answer string,
) ([]types.Participant, []types.Participant, uint64, uint64, error) {

	var allUsers []types.Participant
	var winUsers []types.Participant
	totalPool := uint64(0)
	winnersPool := uint64(0)

	err := k.Participant.Indexes.EventId.Walk(
		ctx,
		collections.NewPrefixedPairRange[uint64, uint64](eventId),
		func(indexingKey uint64, primaryKey uint64) (bool, error) {
			p, err := k.Participant.Get(ctx, primaryKey)
			if err != nil {
				return true, err
			}

			totalPool += p.Amount
			allUsers = append(allUsers, p)

			if p.Answer == answer {
				winUsers = append(winUsers, p)
				winnersPool += p.Amount
			}

			return false, nil
		},
	)

	if err != nil {
		return nil, nil, 0, 0, err
	}

	return allUsers, winUsers, totalPool, winnersPool, nil
}
