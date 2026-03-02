package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetCompanyPercentStore(ctx context.Context, percent uint64) {
	store := k.storeService.OpenKVStore(ctx)
	store.Set([]byte("companyPercent"), sdk.Uint64ToBigEndian(percent))
}

func (k Keeper) GetCompanyPercent(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte("companyPercent"))
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil
	}
	return sdk.BigEndianToUint64(bz), nil
}
