package keeper

import (
	"bettery/x/events/types"
	"context"
	"encoding/binary"
	"fmt"
	"strconv"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) validateEvent(ctx context.Context, data types.Validator) (uint64, uint64, uint64, error) {
	_, winUsers, totalPool, winnerPool, err := k.GetParticipantsByEventWithIndex(ctx, data.EventId, data.Answer)
	if err != nil {
		return 0, 0, 0, err
	}
	if totalPool != 0 {
		if len(winUsers) == 0 && totalPool > 0 {
			companyAddress, err := k.guardKeeper.GetOwner(ctx)
			if err != nil {
				return 0, 0, 0, err
			}
			_, err = k.sendMoney(ctx, companyAddress.String(), totalPool)
			if err != nil {
				return 0, 0, 0, err
			}
			companyAmount := strconv.FormatUint(totalPool, 10)
			id, err := k.AppendValidator(ctx, data, companyAmount, false)
			if err != nil {
				return 0, 0, 0, err
			}
			return id, totalPool, totalPool, nil
		} else if len(winUsers) != 0 && totalPool > 0 {
			id, companyFeeSafe, err := k.letsPayWinners(ctx, data, totalPool, winnerPool, winUsers)
			if err != nil {
				return 0, 0, 0, err
			}
			return id, totalPool, companyFeeSafe, nil
		} else {
			fmt.Println("------------NO CONDITION---------") // TODO CHECK WITH TEST
			return 0, 0, 0, nil
		}

	} else {
		id, err := k.AppendValidator(ctx, data, "0", false)
		if err != nil {
			return 0, 0, 0, err
		}
		return id, 0, 0, nil
	}
}

func (k Keeper) letsPayWinners(ctx context.Context, data types.Validator, totalPool uint64, winnerPool uint64, winUsers []types.Participant) (uint64, uint64, error) {
	totalPoolSafe := sdkmath.NewIntFromUint64(totalPool)
	winnerPoolSafe := sdkmath.NewIntFromUint64(winnerPool)

	companyPercent, err := k.fundsKeeper.GetCompanyPercent(ctx)
	if err != nil {
		return 0, 0, err
	}
	CompanyPercentSafe := sdkmath.NewIntFromUint64(companyPercent)

	companyFeeSafe := totalPoolSafe.Mul(CompanyPercentSafe).Quo(sdkmath.NewInt(100))

	companyAddress, err := k.guardKeeper.GetOwner(ctx)
	if err != nil {
		return 0, 0, err
	}
	_, err = k.sendMoney(ctx, companyAddress.String(), companyFeeSafe.Uint64())
	if err != nil {
		return 0, 0, err
	}

	rewardPool := totalPoolSafe.Sub(companyFeeSafe)
	for _, p := range winUsers {
		amount := sdkmath.NewIntFromUint64(p.Amount)
		rewardSafe := rewardPool.Mul(amount).Quo(winnerPoolSafe)

		_, err := k.sendMoney(ctx, p.Creator, rewardSafe.Uint64())
		if err != nil {
			return 0, 0, err
		}

		_, err = k.updateParticipantFromValidator(ctx, p, rewardSafe.Uint64())
		if err != nil {
			return 0, 0, err
		}
	}
	id, err := k.AppendValidator(ctx, data, strconv.FormatUint(companyFeeSafe.Uint64(), 10), false)
	if err != nil {
		return 0, 0, err
	}

	return id, companyFeeSafe.Uint64(), nil
}

func (k Keeper) refundEvent(ctx context.Context, msg types.Validator) (uint64, uint64, error) {
	allUsers, _, totalPool, _, err := k.GetParticipantsByEventWithIndex(ctx, msg.EventId, msg.Answer)
	if err != nil {
		return 0, 0, err
	}
	if len(allUsers) != 0 && totalPool != 0 {
		for i := range allUsers {
			if allUsers[i].Amount != 0 {
				_, err := k.sendMoney(ctx, allUsers[i].Creator, allUsers[i].Amount)
				if err != nil {
					return 0, 0, err
				}
			}
		}
	}
	id, err := k.AppendValidator(ctx, msg, "0", true)
	if err != nil {
		return 0, 0, err
	}
	return id, 0, nil

}

func (k Keeper) sendMoney(ctx context.Context, address string, amount uint64) (bool, error) {
	if amount == 0 {
		return true, nil
	} else {
		sender, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			return false, err
		}

		coin := sdk.NewCoin(
			types.BetToken,
			math.NewIntFromUint64(amount),
		)

		err = k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			sender,
			sdk.NewCoins(coin),
		)
		if err != nil {
			return false, err
		}

		return true, nil
	}

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
