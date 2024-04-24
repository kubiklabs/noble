package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// RawPacketMetadata defines the raw JSON memo that's used in an autocctp transfer
// The PFM forward key is also used here to validate that the packet is not trying
// to use autocctp and PFM at the same time
// As a result, only the forward key is needed, cause the actual parsing of the PFM
// packet will occur in the PFM module
type RawPacketMetadata struct {
	AutoCctp *struct {
		Receiver string              `json:"receiver"`
		Cctp     *CctpPacketMetadata `json:"cctp,omitempty"`
	} `json:"autocctp"`
	Forward *interface{} `json:"forward"`
}

// AutoCctpActionMetadata stores the metadata that's specific to the autocctp action
// e.g. Fields required for LiquidStake
type AutoCctpMetadata struct {
	Receiver    string
	RoutingInfo ModuleRoutingInfo
}

// ModuleRoutingInfo defines the interface required for each autocctp action
type ModuleRoutingInfo interface {
	Validate() error
}

// GenerateHashedSender generates a new  address for a packet, by hashing
// the channel and original sender.
// This makes the address deterministic and can used to identify the sender
// from the preivous hop
// Additionally, this prevents a forwarded packet from impersonating a different account
// when moving to the next hop (i.e. receiver of one hop, becomes sender of next)
//
// This function was borrowed from PFM
func GenerateHashedAddress(channelId, originalSender string) (string, error) {
	senderStr := fmt.Sprintf("%s/%s", channelId, originalSender)
	senderHash32 := address.Hash(ModuleName, []byte(senderStr))
	sender := sdk.AccAddress(senderHash32[:20])
	bech32Prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	return sdk.Bech32ifyAddressBytes(bech32Prefix, sender)
}
