package types

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const AutoCctp = "AutoCctp"

// Packet metadata info specific to autocctp (e.g. 1-click cctp transfer)
type CctpPacketMetadata struct {
	NobleAddress    string
	TransferChannel string `json:"transfer_channel,omitempty"`

	// CCTP specific
	DestinationDomain uint32 `json:"destination_domain,omitempty"`
	MintRecipient     string `json:"mint_recipient,omitempty"`
}

// Validate autocctp packet metadata fields
// including the noble address and action type
func (m CctpPacketMetadata) Validate() error {
	_, err := sdk.AccAddressFromBech32(m.NobleAddress)
	if err != nil {
		return err
	}
	// TODO: see if this error msg for action type is needed
	// return errorsmod.Wrapf(ErrUnsupportedCctpAction, "action %s is not supported", m.Action)

	return nil
}

// Parse packet metadata intended for x/autocctp
// In the ICS-20 packet, the metadata can optionally indicate a module to route to (e.g. cctp)
// The AutoCctpMetadata returned from this function contains attributes for each x/autocctp supported module
// It can only be forward to one module per packet
// Returns nil if there was no x/autocctp metadata found
func ParseAutoCctpMetadata(metadata string) (*AutoCctpMetadata, error) {
	// If we can't unmarshal the metadata into a PacketMetadata struct,
	// assume packet forwarding was no used and pass back nil so that x/autocctp is ignored
	var raw RawPacketMetadata
	if err := json.Unmarshal([]byte(metadata), &raw); err != nil {
		return nil, nil
	}

	// Packets cannot be used for both x/autocctp and PFM at the same time
	// If both fields were provided, reject the packet
	if raw.AutoCctp != nil && raw.Forward != nil {
		return nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, "x/autocctp and pfm cannot both be used in the same packet")
	}

	// If no forwarding logic was used for x/autocctp, return nil to indicate that
	// there's no x/autocctp action needed
	if raw.AutoCctp == nil {
		return nil, nil
	}

	// Confirm a receiver address was supplied
	if _, err := sdk.AccAddressFromBech32(raw.AutoCctp.Receiver); err != nil {
		return nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, ErrInvalidReceiverAddress.Error())
	}

	// Parse the packet info into the specific module type
	// We increment the module count to ensure only one module type was provided
	var routingInfo ModuleRoutingInfo
	if raw.AutoCctp.Cctp == nil {
		return nil, errorsmod.Wrapf(ErrInvalidPacketMetadata, ErrInvalidModuleRoutes.Error())
	}

	// override the noble address with the receiver address
	raw.AutoCctp.Cctp.NobleAddress = raw.AutoCctp.Receiver
	routingInfo = *raw.AutoCctp.Cctp

	// Validate the packet info according to the specific module type
	if err := routingInfo.Validate(); err != nil {
		return nil, errorsmod.Wrapf(err, ErrInvalidPacketMetadata.Error())
	}

	return &AutoCctpMetadata{
		Receiver:    raw.AutoCctp.Receiver,
		RoutingInfo: routingInfo,
	}, nil
}
