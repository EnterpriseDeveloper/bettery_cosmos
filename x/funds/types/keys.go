package types

import (
	"encoding/binary"
	fmt "fmt"

	"cosmossdk.io/collections"
)

const (
	// ModuleName defines the module name
	ModuleName = "funds"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/gov/types/keys.go#L9
	GovModuleName = "gov"
	BetToken      = "ubet"
	Amount        = "1"
)

// ParamsKey is the prefix to retrieve all Params
var ParamsKey = collections.NewPrefix("p_funds")

var ClaimProcessedPrefix = collections.NewPrefix("funds/claim_processed/")
var BurnNoncePrefix = collections.NewPrefix("funds/burn_nonce/")

func BurnNonceKey(chainID uint64) []byte {

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, chainID)

	return append(BurnNoncePrefix, bz...)
}

func ClaimProcessedKey(
	chainID uint64,
	bridge string,
	nonce uint64,
) []byte {

	key := fmt.Sprintf(
		"%d/%s/%d",
		chainID,
		bridge,
		nonce,
	)

	return append(
		[]byte(ClaimProcessedPrefix),
		[]byte(key)...,
	)
}
