package keeper_test

import (
	"testing"

	"bettery/x/events/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:           types.DefaultParams(),
		EventsList:       []types.Events{{Id: 0}, {Id: 1}},
		EventsCount:      2,
		ParticipantList:  []types.Participant{{Id: 0}, {Id: 1}},
		ParticipantCount: 2,
		ValidatorList:    []types.Validator{{Id: 0}, {Id: 1}},
		ValidatorCount:   2,
	}
	f := initFixture(t)
	err := f.keeper.InitGenesis(f.ctx, genesisState)
	require.NoError(t, err)
	got, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.EqualExportedValues(t, genesisState.Params, got.Params)
	require.EqualExportedValues(t, genesisState.EventsList, got.EventsList)
	require.Equal(t, genesisState.EventsCount, got.EventsCount)
	require.EqualExportedValues(t, genesisState.ParticipantList, got.ParticipantList)
	require.Equal(t, genesisState.ParticipantCount, got.ParticipantCount)
	require.EqualExportedValues(t, genesisState.ValidatorList, got.ValidatorList)
	require.Equal(t, genesisState.ValidatorCount, got.ValidatorCount)

}
