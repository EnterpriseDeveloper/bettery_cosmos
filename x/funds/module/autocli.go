package funds

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"bettery/x/funds/types"
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
					RpcMethod:      "MintToken",
					Use:            "mint-token",
					Short:          "Send a mint-token tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
				{
					RpcMethod:      "MintFromEvm",
					Use:            "mint-from-evm [evm-chain-id] [evm-bridge] [evm-token] [evm-sender] [cosmos-receiver] [amount] [nonce] [tx-hash]",
					Short:          "Send a mint-from-evm tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "evm_chain_id"}, {ProtoField: "evm_bridge"}, {ProtoField: "evm_token"}, {ProtoField: "evm_sender"}, {ProtoField: "cosmos_receiver"}, {ProtoField: "amount"}, {ProtoField: "nonce"}, {ProtoField: "tx_hash"}},
				},
				{
					RpcMethod:      "BurnToEvm",
					Use:            "burn-to-evm [evm-chain-id] [evm-bridge] [evm-token] [evm-recipient] [amount]",
					Short:          "Send a burn-to-evm tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "evm_chain_id"}, {ProtoField: "evm_bridge"}, {ProtoField: "evm_token"}, {ProtoField: "evm_recipient"}, {ProtoField: "amount"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
