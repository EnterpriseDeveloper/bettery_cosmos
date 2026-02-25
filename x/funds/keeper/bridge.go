package keeper

import (
	"bettery/x/funds/types"
	"context"
	"encoding/binary"

	sdkmath "cosmossdk.io/math"
)

// TODO think about security relayer: Proposition was adding at least three relayers and check signatures from them.
// func (k Keeper) VerifySignatures(
// 	ctx context.Context,
// 	hash []byte,
// 	signatures [][]byte,
// ) bool {

// 	params := k.GetParams(ctx)
// 	validCount := 0

// 	for _, sig := range signatures {

// 		pubKey, err := secp256k1.RecoverPubkey(hash, sig)
// 		if err != nil {
// 			continue
// 		}

// 		addr := sdk.AccAddress(crypto.Keccak256(pubKey)[12:])

// 		if k.IsRelayer(ctx, addr.String()) {
// 			validCount++
// 		}
// 	}

// 	return uint64(validCount) >= params.Threshold
// }

func (k Keeper) SetClaimProcessed(
	ctx context.Context,
	chainID uint64,
	bridge string,
	nonce uint64,
) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.ClaimProcessedKey(chainID, bridge, nonce)
	store.Set(key, []byte{1})
}

func (k Keeper) IsClaimProcessed(
	ctx context.Context,
	chainID uint64,
	bridge string,
	nonce uint64,
) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.ClaimProcessedKey(chainID, bridge, nonce)
	exist, err := store.Has(key)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func pow10(exp uint8) sdkmath.Int {
	result := sdkmath.NewInt(1)

	for i := uint8(0); i < exp; i++ {
		result = result.Mul(sdkmath.NewInt(10))
	}

	return result
}

func (k Keeper) GetNextBurnNonce(ctx context.Context, chainID uint64) (uint64, error) {

	store := k.storeService.OpenKVStore(ctx)

	key := types.BurnNonceKey(chainID)

	bz, err := store.Get(key)
	if err != nil {
		return 0, err
	}

	var nonce uint64
	if bz != nil {
		nonce = binary.BigEndian.Uint64(bz)
	}

	nonce++

	newBz := make([]byte, 8)
	binary.BigEndian.PutUint64(newBz, nonce)

	store.Set(key, newBz)

	return nonce, nil
}
