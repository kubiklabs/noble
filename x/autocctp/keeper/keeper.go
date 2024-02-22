package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	cctpkeeper "github.com/circlefin/noble-cctp/x/cctp/keeper"
	"github.com/noble-assets/noble/v5/x/autocctp/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		// memKey     storetypes.StoreKey // TODO: check if needed
		paramstore paramtypes.Subspace

		// channelKeeper types.ChannelKeeper
		// portKeeper    types.PortKeeper
		// scopedKeeper  exported.ScopedKeeper
		cctpKeeper cctpkeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	// memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	// channelKeeper types.ChannelKeeper,
	// portKeeper types.PortKeeper,
	// scopedKeeper exported.ScopedKeeper,
	cctpKeeper cctpkeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		// memKey:     memKey,
		paramstore: ps,

		// channelKeeper: channelKeeper,
		// portKeeper:    portKeeper,
		// scopedKeeper:  scopedKeeper,
		cctpKeeper: cctpKeeper,
	}
}

// ----------------------------------------------------------------------------
// IBC Keeper Logic
// ----------------------------------------------------------------------------

// // ChanCloseInit defines a wrapper function for the channel Keeper's function.
// func (k Keeper) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
// 	capName := host.ChannelCapabilityPath(portID, channelID)
// 	chanCap, ok := k.scopedKeeper.GetCapability(ctx, capName)
// 	if !ok {
// 		return sdkerrors.Wrapf(channeltypes.ErrChannelCapabilityNotFound, "could not retrieve channel capability at: %s", capName)
// 	}
// 	return k.channelKeeper.ChanCloseInit(ctx, portID, channelID, chanCap)
// }

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
