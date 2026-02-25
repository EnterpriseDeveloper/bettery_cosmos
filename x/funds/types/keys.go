package types

import (
	"encoding/binary"
	"encoding/hex"
	fmt "fmt"
	"math/big"

	"cosmossdk.io/collections"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	Amount        = "10000000"
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

func HashClaim(msg *MsgMintFromEvm) []byte {

	var packed []byte

	// --- chainId ---
	chainId := new(big.Int).SetUint64(msg.EvmChainId)
	chainIdBytes := common.LeftPadBytes(chainId.Bytes(), 32)
	packed = append(packed, chainIdBytes...)

	// --- bridge ---
	bridgeAddr := common.HexToAddress(msg.EvmBridge)
	packed = append(packed, bridgeAddr.Bytes()...)

	// --- token ---
	tokenAddr := common.HexToAddress(msg.EvmToken)
	packed = append(packed, tokenAddr.Bytes()...)

	// --- sender ---
	senderAddr := common.HexToAddress(msg.EvmSender)
	packed = append(packed, senderAddr.Bytes()...)

	// --- cosmos receiver (raw bytes) ---
	packed = append(packed, []byte(msg.CosmosReceiver)...)

	// --- amount ---
	amountInt, _ := new(big.Int).SetString(msg.Amount, 10)
	amountBytes := common.LeftPadBytes(amountInt.Bytes(), 32)
	packed = append(packed, amountBytes...)

	// --- nonce ---
	nonce := new(big.Int).SetUint64(msg.Nonce)
	nonceBytes := common.LeftPadBytes(nonce.Bytes(), 32)
	packed = append(packed, nonceBytes...)

	// --- txHash ---
	txHashBytes, _ := hex.DecodeString(msg.TxHash[2:]) // remove 0x
	packed = append(packed, txHashBytes...)

	hash := crypto.Keccak256(packed)

	return hash
}
