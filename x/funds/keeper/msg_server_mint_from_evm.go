package keeper

import (
	"context"

	"bettery/x/funds/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) MintFromEvm(ctx context.Context, msg *types.MsgMintFromEvm) (*types.MsgMintFromEvmResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgMintFromEvmResponse{}, nil
}
