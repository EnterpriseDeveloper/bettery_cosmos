package events

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"bettery/x/events/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "ListEvents",
					Use:       "list-events",
					Short:     "List all events",
				},
				{
					RpcMethod:      "GetEvents",
					Use:            "get-events [id]",
					Short:          "Gets a events by id",
					Alias:          []string{"show-events"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod: "ListParticipant",
					Use:       "list-participant",
					Short:     "List all participant",
				},
				{
					RpcMethod:      "GetParticipant",
					Use:            "get-participant [id]",
					Short:          "Gets a participant by id",
					Alias:          []string{"show-participant"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod: "ListValidator",
					Use:       "list-validator",
					Short:     "List all validator",
				},
				{
					RpcMethod:      "GetValidator",
					Use:            "get-validator [id]",
					Short:          "Gets a validator by id",
					Alias:          []string{"show-validator"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod:      "CreateEvent",
					Use:            "create-event [question] [answers] [start-time] [end-time] [category]",
					Short:          "Send a create-event tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "question"}, {ProtoField: "answers"}, {ProtoField: "end_time"}, {ProtoField: "category"}},
				},
				{
					RpcMethod:      "CreatePartEvent",
					Use:            "create-part-event [event-id] [answers] [amount]",
					Short:          "Send a create-part-event tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "event_id"}, {ProtoField: "answers"}, {ProtoField: "amount"}},
				},
				{
					RpcMethod:      "ValidateEvent",
					Use:            "validate-event [event-id] [answers] [source]",
					Short:          "Send a validate-event tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "event_id"}, {ProtoField: "answers"}, {ProtoField: "source"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
