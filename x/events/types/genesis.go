package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		EventsList: []Events{}}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	eventsIdMap := make(map[uint64]bool)
	eventsCount := gs.GetEventsCount()
	for _, elem := range gs.EventsList {
		if _, ok := eventsIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for events")
		}
		if elem.Id >= eventsCount {
			return fmt.Errorf("events id should be lower or equal than the last id")
		}
		eventsIdMap[elem.Id] = true
	}

	return gs.Params.Validate()
}
