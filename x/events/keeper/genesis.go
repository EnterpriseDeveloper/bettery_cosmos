package keeper

import (
	"context"

	"bettery/x/events/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	for _, elem := range genState.EventsList {
		if err := k.Events.Set(ctx, elem.Id, elem); err != nil {
			return err
		}
	}

	if err := k.EventsSeq.Set(ctx, genState.EventsCount); err != nil {
		return err
	}
	for _, elem := range genState.ParticipantList {
		if err := k.Participant.Set(ctx, elem.Id, elem); err != nil {
			return err
		}
	}

	if err := k.ParticipantSeq.Set(ctx, genState.ParticipantCount); err != nil {
		return err
	}

	return k.Params.Set(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params, err = k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	err = k.Events.Walk(ctx, nil, func(key uint64, elem types.Events) (bool, error) {
		genesis.EventsList = append(genesis.EventsList, elem)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	genesis.EventsCount, err = k.EventsSeq.Peek(ctx)
	if err != nil {
		return nil, err
	}
	err = k.Participant.Walk(ctx, nil, func(key uint64, elem types.Participant) (bool, error) {
		genesis.ParticipantList = append(genesis.ParticipantList, elem)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	genesis.ParticipantCount, err = k.ParticipantSeq.Peek(ctx)
	if err != nil {
		return nil, err
	}

	return genesis, nil
}
