package keeper

import (
	"errors"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	cctpkeeper "github.com/circlefin/noble-cctp/x/cctp/keeper"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	"github.com/noble-assets/noble/v5/x/autocctp/types"
)

const (
	// If the forward transfer fails, the tokens are sent to the fallback address
	// which is a less than ideal UX
	// As a result, we decided to use a long timeout here such, even in the case
	// of high activity, a timeout should be very unlikely to occur
	// Empirically we found that times of high market stress took roughly
	// 2 hours for transfers to complete
	CctpForwardTransferTimeout = (time.Hour * 3)
)

// Attempts to do an x/autocctp cctp transfer (and optional forward)
// The cctp transfer is only allowed if the inbound packet came along a trusted channel
func (k Keeper) TryCctp(
	ctx sdk.Context,
	packet channeltypes.Packet,
	transferMetadata transfertypes.FungibleTokenPacketData,
	autopilotMetadata types.CctpPacketMetadata,
) error {
	params := k.GetParams(ctx)
	if !params.CctpActive {
		return errorsmod.Wrapf(types.ErrAutoCctpInactive, "x/autocctp cctp routing is inactive")
	}

	// Verify the amount is valid
	amount, ok := sdkmath.NewIntFromString(transferMetadata.Amount)
	if !ok {
		return errors.New("not a parsable amount field")
	}

	// TODO: verify the denom to cctp
	// and if it matches that of the amount sent

	// TODO: // In this case, we can't process a cctp transaction, because we're dealing with native tokens (e.g. STRD, stATOM)
	// if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), transferMetadata.Denom) {
	// 	return fmt.Errorf("native token is not supported for cctp (%s)", transferMetadata.Denom)
	// }

	// TODO: // Note: the denom in the packet is the base denom e.g. uatom - not ibc/xxx
	// // We need to use the port and channel to build the IBC denom
	// prefixedDenom := transfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), transferMetadata.Denom)
	// ibcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	// hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, transferMetadata.Denom)
	// if err != nil {
	// 	return err
	// }

	// // Verify the IBC denom of the packet matches the host zone, to confirm the packet
	// // was sent over a trusted channel
	// if hostZone.IbcDenom != ibcDenom {
	// 	return fmt.Errorf("ibc denom %s is not equal to host zone ibc denom %s", ibcDenom, hostZone.IbcDenom)
	// }

	return k.RunCctp(ctx, amount, transferMetadata, autopilotMetadata)
}

// Submits a LiquidStake message from the transfer receiver
// If a forwarding recipient is specified, the stTokens are ibc transferred
func (k Keeper) RunCctp(
	ctx sdk.Context,
	amount sdkmath.Int,
	transferMetadata transfertypes.FungibleTokenPacketData,
	autopilotMetadata types.CctpPacketMetadata,
) error {
	// cctp transaction message
	msg := &cctptypes.MsgDepositForBurn{
		From:              transferMetadata.Receiver, // TODO: check if noble address
		Amount:            amount,
		DestinationDomain: autopilotMetadata.DestinationDomain, // TODO: fix this hardcode, fetch the map of domains and validate the "memo" packet domain value
		MintRecipient:     []byte(autopilotMetadata.MintRecipient),
		BurnToken:         "",
	}

	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	msgServer := cctpkeeper.NewMsgServerImpl(&k.cctpKeeper)
	_, err := msgServer.DepositForBurn(
		sdk.WrapSDKContext(ctx),
		msg,
	)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to cctp")
	}

	// Note: no IBC forwarding, unlike stride's x/autopilot
	return nil
}
