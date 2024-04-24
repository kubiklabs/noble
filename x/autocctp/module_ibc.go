package autocctp

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/noble-assets/noble/v5/x/autocctp/keeper"
	"github.com/noble-assets/noble/v5/x/autocctp/types"
)

// Used by the module to define internally how long memo field in ibc msg could be
const MaxMemoCharLength = 2000

// IBC MODULE IMPLEMENTATION
// IBCModule implements the ICS26 interface for transfer given the transfer keeper.
// TODO: Use IBCMiddleware struct
type IBCModule struct {
	keeper keeper.Keeper
	app    porttypes.IBCModule
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper, app porttypes.IBCModule) IBCModule {
	return IBCModule{
		keeper: k,
		app:    app,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		chanCap,
		counterparty,
		version,
	)
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return im.app.OnChanOpenTry(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		chanCap,
		counterparty,
		counterpartyVersion,
	)
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelId string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(
		ctx,
		portID,
		channelID,
		counterpartyChannelId,
		counterpartyVersion,
	)
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanOpenConfirm(
		ctx,
		portID,
		channelID,
	)
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseInit(
		ctx,
		portID,
		channelID,
	)
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseConfirm(
		ctx,
		portID,
		channelID,
	)
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	// TODO: below is rough flow in comments
	// 1. unmarshall the packet
	// 2. check for valid memo and receiver fields
	// 3. check if autocctp metadata is there in memo
	// 4. if not autocctp packet, directly send the packet down the stack
	// 5. if autocctp packet, update the sender, extract metdata and send the packet down the stack
	// 6. do the autocctp related action from the packet memo and this can be moved to keeper

	im.keeper.Logger(ctx).Info(
		fmt.Sprintf("OnRecvPacket (autocctp): Sequence: %d, Source: %s, %s; Destination: %s, %s",
			modulePacket.Sequence,
			modulePacket.SourcePort,
			modulePacket.SourceChannel,
			modulePacket.DestinationPort,
			modulePacket.DestinationChannel,
		),
	)

	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	var tokenPacketData transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(modulePacket.GetData(), &tokenPacketData); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// TODO: UDIT: Uncomment this and add ther errors beforehand
	// Error any transactions with a Memo or Receiver field are greater than the max characters
	// if len(tokenPacketData.Memo) > MaxMemoCharLength {
	// 	return channeltypes.NewErrorAcknowledgement(errorsmod.Wrapf(types.ErrInvalidMemoSize, "memo length: %d", len(tokenPacketData.Memo)))
	// }
	// if len(tokenPacketData.Receiver) > MaxMemoCharLength {
	// 	return channeltypes.NewErrorAcknowledgement(errorsmod.Wrapf(types.ErrInvalidMemoSize, "receiver length: %d", len(tokenPacketData.Receiver)))
	// }

	// ibc-go v5 has a Memo field that can store forwarding info
	// For older version of ibc-go, the data must be stored in the receiver field
	var metadata string
	if tokenPacketData.Memo != "" { // ibc-go v5+
		metadata = tokenPacketData.Memo
	} else { // before ibc-go v5
		metadata = tokenPacketData.Receiver
	}

	// If a valid receiver address has been provided and no memo,
	// this is clearly just an normal IBC transfer
	// Pass down the stack immediately instead of parsing
	_, err := sdk.AccAddressFromBech32(tokenPacketData.Receiver)
	if err == nil && tokenPacketData.Memo == "" {
		return im.app.OnRecvPacket(ctx, modulePacket, relayer)
	}

	// parse out any autocctp forwarding info
	autoCctpMetadata, err := types.ParseAutoCctpMetadata(metadata)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// If the parsed metadata is nil, that means there is no autocctp forwarding logic
	// Pass the packet down to the next middleware
	// PFM packets will also go down this path
	if autoCctpMetadata == nil {
		return im.app.OnRecvPacket(ctx, modulePacket, relayer)
	}

	// -- At this point, we are officially dealing with an autocctp packet --

	// Update the reciever in the packet data so that we can pass the packet down the stack
	// (since the "receiver" may have technically been a full JSON memo, in case of ibc-go before v5)
	tokenPacketData.Receiver = autoCctpMetadata.Receiver

	// For autopilot liquid stake and forward, we'll override the receiver with a hashed address
	// The hashed address will also be the sender of the outbound transfer
	// This is to prevent impersonation at downstream zones
	// We can identify the forwarding step by whether there's a non-empty IBC receiver field
	if _, ok := autoCctpMetadata.RoutingInfo.(types.CctpPacketMetadata); ok {
		hashedReceiver, err := types.GenerateHashedAddress(modulePacket.DestinationChannel, tokenPacketData.Sender)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		tokenPacketData.Receiver = hashedReceiver
	}

	// Now that the receiver's been updated on the transfer metadata,
	// modify the original packet so that we can send it down the stack
	bz, err := transfertypes.ModuleCdc.MarshalJSON(&tokenPacketData)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}
	newPacket := modulePacket
	newPacket.Data = bz

	// Pass the new packet down the middleware stack first to complete the transfer
	ack := im.app.OnRecvPacket(ctx, newPacket, relayer)
	if !ack.Success() {
		return ack
	}

	autocctpParams := im.keeper.GetParams(ctx)
	sender := tokenPacketData.Sender

	// If cctp routing is inactive (but the packet had routing info in the memo) return an ack error
	if !autocctpParams.CctpActive {
		im.keeper.Logger(ctx).Error(
			fmt.Sprintf("Packet from %s had cctp routing info but autopilot cctp routing is disabled", sender),
		)
		return channeltypes.NewErrorAcknowledgement(types.ErrAutoCctpInactive)
	}

	im.keeper.Logger(ctx).Info(fmt.Sprintf("Forwarding packet from %s to cctp route", sender))

	// Try to perform cctp transfer - return an ack error if it fails,
	// otherwise return the ack generated from the earlier packet propogation
	if routingInfo, ok := autoCctpMetadata.RoutingInfo.(types.CctpPacketMetadata); ok {
		if err := im.keeper.TryCctp(ctx, modulePacket, tokenPacketData, routingInfo); err != nil {
			im.keeper.Logger(ctx).Error(
				fmt.Sprintf("Error doing cctp transfer for %s: %s", sender, err.Error()),
			)
			return channeltypes.NewErrorAcknowledgement(err)
		}
		return ack
	} else {
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrapf(types.ErrUnsupportedRoute, "%T", routingInfo))
	}
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	// im.keeper.Logger(ctx).Info(
	// 	fmt.Sprintf("OnAcknowledgementPacket (autoCctp): Packet %v, Acknowledgement %v", modulePacket, acknowledgement),
	// )
	// // First pass the packet down the stack so that, in the event of an ack failure,
	// // the tokens are refunded to the original sender
	// if err := im.app.OnAcknowledgementPacket(ctx, modulePacket, acknowledgement, relayer); err != nil {
	// 	return err
	// }
	// // Then process the autoCctp-specific callback
	// // This will handle bank sending to a fallback address if the original transfer failed
	// return im.keeper.OnAcknowledgementPacket(ctx, modulePacket, acknowledgement)
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "UNIMPLEMENTED")
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// im.keeper.Logger(ctx).Error(
	// 	fmt.Sprintf("OnTimeoutPacket (autoCctp): Packet %v", modulePacket),
	// )
	// // First pass the packet down the stack so that the tokens are refunded to the original sender
	// if err := im.app.OnTimeoutPacket(ctx, modulePacket, relayer); err != nil {
	// 	return err
	// }
	// // Then process the autoCctp-specific callback
	// // This will handle a retry in the event that there was a timeout during an autoCctp action
	// return im.keeper.OnTimeoutPacket(ctx, modulePacket)
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "UNIMPLEMENTED")
}

// TODO: check if below 3 methods are actually needed
// context: copied from Stride-Labs/x/autopilot

// This is implemented by ICS4 and all middleware that are wrapping base application.
// The base application will call `sendPacket` or `writeAcknowledgement` of the middleware directly above them
// which will call the next middleware until it reaches the core IBC handler.
// SendPacket implements the ICS4 Wrapper interface
func (im IBCModule) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
) error {
	return nil
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im IBCModule) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	return nil
}

// GetAppVersion returns the interchain accounts metadata.
func (im IBCModule) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return transfertypes.Version, true // im.keeper.GetAppVersion(ctx, portID, channelID)
}
