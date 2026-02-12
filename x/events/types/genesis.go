package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		EventsList: []Events{}, ParticipantList: []Participant{}, ValidatorList: []Validator{}}
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
	participantIdMap := make(map[uint64]bool)
	participantCount := gs.GetParticipantCount()
	for _, elem := range gs.ParticipantList {
		if _, ok := participantIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for participant")
		}
		if elem.Id >= participantCount {
			return fmt.Errorf("participant id should be lower or equal than the last id")
		}
		participantIdMap[elem.Id] = true
	}
	validatorIdMap := make(map[uint64]bool)
	validatorCount := gs.GetValidatorCount()
	for _, elem := range gs.ValidatorList {
		if _, ok := validatorIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for validator")
		}
		if elem.Id >= validatorCount {
			return fmt.Errorf("validator id should be lower or equal than the last id")
		}
		validatorIdMap[elem.Id] = true
	}

	return gs.Params.Validate()
}
