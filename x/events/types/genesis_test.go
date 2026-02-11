package types_test

import (
	"testing"

	"bettery/x/events/types"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc:     "valid genesis state",
			genState: &types.GenesisState{EventsList: []types.Events{{Id: 0}, {Id: 1}}, EventsCount: 2}, valid: true,
		}, {
			desc: "duplicated events",
			genState: &types.GenesisState{
				EventsList: []types.Events{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
			},
			valid: false,
		}, {
			desc: "invalid events count",
			genState: &types.GenesisState{
				EventsList: []types.Events{
					{
						Id: 1,
					},
				},
				EventsCount: 0,
			},
			valid: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
