package keeper

import (
	"bettery/x/events/types"
	"context"
	"encoding/binary"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) validateEvent(ctx context.Context, data types.Validator) (uint64, error) {
	allUsers, winUsers, totalPool, winnerPool, err := k.GetParticipantsByEvent(ctx, data.EventId, data.Answer)
	if err != nil {
		return 0, err
	}
	if len(allUsers) != 0 {
		if len(winUsers) == 0 && totalPool > 0 {
			_, err := k.sendMoney(ctx, types.CompanyAddress, totalPool)
			if err != nil {
				return 0, err
			}
			companyAmount := strconv.FormatUint(totalPool, 10)
			id, err := k.AppendValidator(ctx, data, companyAmount, false)
			if err != nil {
				return 0, err
			}
			return id, nil
		} else if len(allUsers)-len(winUsers) == 0 {
			id, err := k.refundEvent(ctx, data)
			if err != nil {
				return 0, err
			}
			return id, nil
		} else {
			companyFee := totalPool * uint64(types.CompanyPercent) / 100
			_, err := k.sendMoney(ctx, types.CompanyAddress, companyFee)
			if err != nil {
				return 0, err
			}

			rewardPool := totalPool - companyFee
			for _, p := range winUsers {
				reward := rewardPool * p.Amount / winnerPool
				_, err := k.sendMoney(ctx, p.Creator, reward)
				if err != nil {
					return 0, err
				}
			}
			id, err := k.AppendValidator(ctx, data, strconv.FormatUint(companyFee, 10), false)
			if err != nil {
				return 0, err
			}
			return id, nil
		}

	} else {
		id, err := k.AppendValidator(ctx, data, "0", false)
		if err != nil {
			return 0, err
		}
		return id, nil
	}
}

func (k Keeper) refundEvent(ctx context.Context, msg types.Validator) (uint64, error) {
	allUsers, _, _, _, err := k.GetParticipantsByEvent(ctx, msg.EventId, msg.Answer)
	if err != nil {
		return 0, err
	}
	if len(allUsers) != 0 {
		for i := range allUsers {
			_, err := k.sendMoney(ctx, allUsers[i].Creator, allUsers[i].Amount)
			if err != nil {
				return 0, err
			}
		}
	}
	id, err := k.AppendValidator(ctx, msg, "0", true)
	if err != nil {
		return 0, err
	}
	return id, nil

}

func (k Keeper) sendMoney(ctx context.Context, address string, amount uint64) (bool, error) {
	sender, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return false, err
	}
	coin, err := sdk.ParseCoinNormalized(strconv.FormatUint(amount, 10))
	if err != nil {
		return false, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, sdk.NewCoins(coin))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (k Keeper) AppendValidator(
	ctx context.Context,
	validator types.Validator,
	amount string,
	refunded bool,
) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	id, err := k.GetValidatorCount(ctx)
	if err != nil {
		return 0, err
	}
	validator.Id = id
	validator.CompanyAmount = amount
	validator.Refunded = refunded
	appendedValue := k.cdc.MustMarshal(&validator)
	store.Set(types.ValidatorKey(validator.Id), appendedValue)
	k.SetValidatorCount(ctx, id+1)

	var status string
	if refunded {
		status = types.RefundEvent
	} else {
		status = types.FinishedEvent
	}
	_, err = k.updateEventFromValidator(ctx, validator, status)
	if err != nil {
		return 0, err
	}

	return validator.Id, nil

}

func (k Keeper) GetValidatorCount(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.ValidatorCountKey)
	if err != nil || bz == nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(bz), nil
}

func (k Keeper) SetValidatorCount(ctx context.Context, count uint64) {
	store := k.storeService.OpenKVStore(ctx)

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)

	store.Set(types.ValidatorCountKey, bz)
}
