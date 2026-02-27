package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetOwner(ctx context.Context, owner sdk.AccAddress) {
	store := k.storeService.OpenKVStore(ctx)
	store.Set([]byte("owner"), owner.Bytes())
}

func (k Keeper) GetOwner(ctx context.Context) (sdk.AccAddress, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte("owner"))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, nil
	}
	return sdk.AccAddress(bz), nil
}
