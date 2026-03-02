package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"bettery/x/events/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema collections.Schema
	Params collections.Item[types.Params]

	mintKeeper     types.MintKeeper
	bankKeeper     types.BankKeeper
	guardKeeper    types.GuardKeeper
	fundsKeeper    types.FundsKeeper
	EventsSeq      collections.Sequence
	Events         collections.Map[uint64, types.Events]
	ParticipantSeq collections.Sequence
	Participant    *collections.IndexedMap[uint64, types.Participant, ParticipantIndexes]
	ValidatorSeq   collections.Sequence
	Validator      collections.Map[uint64, types.Validator]
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

	mintKeeper types.MintKeeper,
	bankKeeper types.BankKeeper,
	guardKeeper types.GuardKeeper,
	fundsKeeper types.FundsKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,

		mintKeeper:  mintKeeper,
		bankKeeper:  bankKeeper,
		guardKeeper: guardKeeper,
		fundsKeeper: fundsKeeper,
		Params:      collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Events:      collections.NewMap(sb, types.EventsKeyPrefix, "events", collections.Uint64Key, codec.CollValue[types.Events](cdc)),
		EventsSeq:   collections.NewSequence(sb, types.EventsCountKey, "eventsSequence"),
		Participant: collections.NewIndexedMap(
			sb,
			types.ParticipantKeyPrefix,
			"participant",
			collections.Uint64Key,
			codec.CollValue[types.Participant](cdc),
			ParticipantIndexes{
				EventId: indexes.NewMulti(
					sb,
					types.ParticipantEventIdIndexKey,
					"participant_by_event_id",
					collections.Uint64Key,
					collections.Uint64Key,
					func(_ uint64, p types.Participant) (uint64, error) {
						return p.EventId, nil
					},
				),
			},
		),
		ParticipantSeq: collections.NewSequence(sb, types.ParticipantCountKey, "participantSequence"),
		Validator:      collections.NewMap(sb, types.ValidatorKeyPrefix, "validator", collections.Uint64Key, codec.CollValue[types.Validator](cdc)),
		ValidatorSeq:   collections.NewSequence(sb, types.ValidatorCountKey, "validatorSequence"),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}
