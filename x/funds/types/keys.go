package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "funds"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/gov/types/keys.go#L9
	GovModuleName = "gov"
	BetToken      = "bet"
)

// ParamsKey is the prefix to retrieve all Params
var ParamsKey = collections.NewPrefix("p_funds")

func FindToken(symbol string) string {
	switch sb := symbol; sb {
	case "BET":
		return BetToken
	// TODO: add more tokens here
	default:
		return "Unknown"
	}
}
