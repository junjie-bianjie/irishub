package ibc

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(IBCTransferMsg{}, "cosmos-sdk/IBCTransferMsg", nil)
	cdc.RegisterConcrete(IBCReceiveMsg{}, "cosmos-sdk/IBCReceiveMsg", nil)
	cdc.RegisterConcrete(IBCSetMsg{},"cosmos-sdk/IBCSetMsg",nil)
	cdc.RegisterConcrete(IBCGetMsg{},"cosmos-sdk/IBCGetMsg",nil)
}
