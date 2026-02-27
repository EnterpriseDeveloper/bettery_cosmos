package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SetNewOwner(ctx context.Context, owner sdk.AccAddress) {
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

func (k Keeper) IsOwner(ctx context.Context, addr sdk.AccAddress) (bool, error) {
	owner, err := k.GetOwner(ctx)
	if err != nil {
		return false, err
	}
	if owner == nil {
		return false, nil
	}
	return addr.Equals(owner), nil
}
