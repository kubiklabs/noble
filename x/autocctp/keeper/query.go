package keeper

import (
	"github.com/noble-assets/noble/v5/x/autocctp/types"
)

var _ types.QueryServer = Keeper{}
