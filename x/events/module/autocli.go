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
					Use:            "create-event [id] [question] [answers] [start-time] [end-time] [category]",
					Short:          "Send a create-event tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "question"}, {ProtoField: "answers"}, {ProtoField: "start_time"}, {ProtoField: "end_time"}, {ProtoField: "category"}},
				},
				{
					RpcMethod:      "CreatePartEvent",
					Use:            "create-part-event [id] [event-id] [answers] [amount] [answer-index] [token]",
					Short:          "Send a create-part-event tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "event_id"}, {ProtoField: "answers"}, {ProtoField: "amount"}, {ProtoField: "answer_index"}, {ProtoField: "token"}},
				},
				{
					RpcMethod:      "ValidateEvent",
					Use:            "validate-event [id] [event-id] [answer-index] [answers] [source]",
					Short:          "Send a validate-event tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "event_id"}, {ProtoField: "answer_index"}, {ProtoField: "answers"}, {ProtoField: "source"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
