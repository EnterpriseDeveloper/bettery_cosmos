package events

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	eventssimulation "bettery/x/events/simulation"
	"bettery/x/events/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	eventsGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&eventsGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgCreateEvent          = "op_weight_msg_events"
		defaultWeightMsgCreateEvent int = 100
	)

	var weightMsgCreateEvent int
	simState.AppParams.GetOrGenerate(opWeightMsgCreateEvent, &weightMsgCreateEvent, nil,
		func(_ *rand.Rand) {
			weightMsgCreateEvent = defaultWeightMsgCreateEvent
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateEvent,
		eventssimulation.SimulateMsgCreateEvent(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgCreatePartEvent          = "op_weight_msg_events"
		defaultWeightMsgCreatePartEvent int = 100
	)

	var weightMsgCreatePartEvent int
	simState.AppParams.GetOrGenerate(opWeightMsgCreatePartEvent, &weightMsgCreatePartEvent, nil,
		func(_ *rand.Rand) {
			weightMsgCreatePartEvent = defaultWeightMsgCreatePartEvent
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreatePartEvent,
		eventssimulation.SimulateMsgCreatePartEvent(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgValidateEvent          = "op_weight_msg_events"
		defaultWeightMsgValidateEvent int = 100
	)

	var weightMsgValidateEvent int
	simState.AppParams.GetOrGenerate(opWeightMsgValidateEvent, &weightMsgValidateEvent, nil,
		func(_ *rand.Rand) {
			weightMsgValidateEvent = defaultWeightMsgValidateEvent
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgValidateEvent,
		eventssimulation.SimulateMsgValidateEvent(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
