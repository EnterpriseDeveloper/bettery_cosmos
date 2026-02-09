package keeper

import (
	"bettery/x/funds/types"
	"context"
)

func (k Keeper) HasMint(ctx context.Context, creator string) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	data, err := store.Has([]byte(creator))
	if err != nil {
		return false, err
	}
	return data, nil
}

func (k Keeper) AppendMintData(
	ctx context.Context,
	mintData types.MintEvent,
) string {
	store := k.storeService.OpenKVStore(ctx)
	appendedValue := k.cdc.MustMarshal(&mintData)
	store.Set([]byte(mintData.Creator), appendedValue)
	return mintData.Creator
}
