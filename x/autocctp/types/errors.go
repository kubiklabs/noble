package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/autocctp module sentinel errors
var (
	ErrInvalidPacketMetadata  = errorsmod.Register(ModuleName, 1501, "invalid packet metadata")
	ErrUnsupportedCctpAction  = errorsmod.Register(ModuleName, 1502, "unsupported cctp action")
	ErrInvalidModuleRoutes    = errorsmod.Register(ModuleName, 1504, "invalid number of module routes, only 1 module is allowed at a time")
	ErrUnsupportedRoute       = errorsmod.Register(ModuleName, 1505, "unsupported autocctp route")
	ErrInvalidReceiverAddress = errorsmod.Register(ModuleName, 1506, "receiver address must be specified when using autocctp")
	ErrInvalidMemoSize        = errorsmod.Register(ModuleName, 1508, "the memo or receiver field exceeded the max allowable size")
	ErrAutoCctpInactive       = errorsmod.Register(ModuleName, 1507, "autocctp packet forwarding is disabled")
)
