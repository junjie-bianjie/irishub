package bank

import (
	"encoding/json"
	"fmt"
	"regexp"

	sdk "github.com/irisnet/irishub/types"
)

const memoRegexpLengthLimit = 50

// MsgSend - high level transaction of the coin module
type MsgSend struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

var _ sdk.Msg = MsgSend{}

// NewMsgSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgSend(in []Input, out []Output) MsgSend {
	return MsgSend{Inputs: in, Outputs: out}
}

// Implements Msg.
// nolint
func (msg MsgSend) Route() string { return "bank" } // TODO: "bank/send"
func (msg MsgSend) Type() string  { return "send" }

// Implements Msg.
func (msg MsgSend) ValidateBasic() sdk.Error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs(DefaultCodespace).TraceSDK("")
	}
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs(DefaultCodespace).TraceSDK("")
	}
	// make sure all inputs and outputs are individually valid
	var totalIn, totalOut sdk.Coins
	for _, in := range msg.Inputs {
		if err := in.ValidateBasic(); err != nil {
			return err.TraceSDK("")
		}
		totalIn = totalIn.Add(in.Coins)
	}
	for _, out := range msg.Outputs {
		if err := out.ValidateBasic(); err != nil {
			return err.TraceSDK("")
		}
		totalOut = totalOut.Add(out.Coins)
	}
	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return sdk.ErrInvalidCoins(totalIn.String()).TraceSDK("inputs and outputs don't match")
	}
	return nil
}

// Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	var inputs, outputs []json.RawMessage
	for _, input := range msg.Inputs {
		inputs = append(inputs, input.GetSignBytes())
	}
	for _, output := range msg.Outputs {
		outputs = append(outputs, output.GetSignBytes())
	}
	b, err := msgCdc.MarshalJSON(struct {
		Inputs  []json.RawMessage `json:"inputs"`
		Outputs []json.RawMessage `json:"outputs"`
	}{
		Inputs:  inputs,
		Outputs: outputs,
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg.
func (msg MsgSend) GetSigners() []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

//----------------------------------------
// Input

// Transaction Input
type Input struct {
	Address sdk.AccAddress `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
}

// Return bytes to sign for Input
func (in Input) GetSignBytes() []byte {
	bin, err := msgCdc.MarshalJSON(in)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bin)
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() sdk.Error {
	if len(in.Address) == 0 {
		return sdk.ErrInvalidAddress(in.Address.String())
	}
	if in.Coins.Empty() {
		return sdk.ErrInvalidCoins("empty input coins")
	}
	if !in.Coins.IsValid() {
		return sdk.ErrInvalidCoins(fmt.Sprintf("invalid input coins [%s]", in.Coins))
	}
	return nil
}

// NewInput - create a transaction input, used with MsgSend
func NewInput(addr sdk.AccAddress, coins sdk.Coins) Input {
	input := Input{
		Address: addr,
		Coins:   coins,
	}
	return input
}

//----------------------------------------
// Output

// Transaction Output
type Output struct {
	Address sdk.AccAddress `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
}

// Return bytes to sign for Output
func (out Output) GetSignBytes() []byte {
	bin, err := msgCdc.MarshalJSON(out)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bin)
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() sdk.Error {
	if len(out.Address) == 0 {
		return sdk.ErrInvalidAddress(out.Address.String())
	}
	if out.Coins.Empty() {
		return sdk.ErrInvalidCoins("empty output coins")
	}
	if !out.Coins.IsValid() {
		return sdk.ErrInvalidCoins(fmt.Sprintf("invalid output coins [%s]", out.Coins))
	}
	return nil
}

// NewOutput - create a transaction output, used with MsgSend
func NewOutput(addr sdk.AccAddress, coins sdk.Coins) Output {
	output := Output{
		Address: addr,
		Coins:   coins,
	}
	return output
}

//----------------------------------------
// MsgBurn

// MsgBurn - high level transaction of the coin module
type MsgBurn struct {
	Owner sdk.AccAddress `json:"owner"`
	Coins sdk.Coins      `json:"coins"`
}

var _ sdk.Msg = MsgBurn{}

// NewMsgBurn - construct MsgBurn
func NewMsgBurn(owner sdk.AccAddress, coins sdk.Coins) MsgBurn {
	return MsgBurn{Owner: owner, Coins: coins}
}

// Implements Msg.
// nolint
func (msg MsgBurn) Route() string { return "bank" }
func (msg MsgBurn) Type() string  { return "burn" }

// Implements Msg.
func (msg MsgBurn) ValidateBasic() sdk.Error {
	if len(msg.Owner) == 0 {
		return sdk.ErrInvalidAddress(msg.Owner.String())
	}
	if msg.Coins.Empty() {
		return sdk.ErrInvalidCoins("empty coins to burn")
	}
	if !msg.Coins.IsValid() {
		return sdk.ErrInvalidCoins(fmt.Sprintf("invalid coins to burn [%s]", msg.Coins))
	}
	return nil
}

// Implements Msg.
func (msg MsgBurn) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg.
func (msg MsgBurn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

//----------------------------------------
// MsgSetMemoRegexp

// MsgSetMemoRegexp - set memo regexp
type MsgSetMemoRegexp struct {
	Owner      sdk.AccAddress `json:"owner"`
	MemoRegexp string         `json:"memo_regexp"`
}

var _ sdk.Msg = MsgSetMemoRegexp{}

// NewMsgSetMemoRegexp - construct MsgSetMemoRegexp
func NewMsgSetMemoRegexp(owner sdk.AccAddress, memoRegexp string) MsgSetMemoRegexp {
	return MsgSetMemoRegexp{Owner: owner, MemoRegexp: memoRegexp}
}

// Implements Msg.
// nolint
func (msg MsgSetMemoRegexp) Route() string { return "bank" }
func (msg MsgSetMemoRegexp) Type() string  { return "set-memo-regexp" }

// Implements Msg.
func (msg MsgSetMemoRegexp) ValidateBasic() sdk.Error {
	if len(msg.Owner) == 0 {
		return sdk.ErrInvalidAddress(msg.Owner.String())
	}
	if len(msg.MemoRegexp) > memoRegexpLengthLimit {
		return ErrInvalidMemoRegexp(DefaultCodespace, "memo regexp length exceeds limit")
	}
	if _, err := regexp.Compile(msg.MemoRegexp); err != nil {
		return ErrInvalidMemoRegexp(DefaultCodespace, "invalid memo regexp")
	}
	return nil
}

// Implements Msg.
func (msg MsgSetMemoRegexp) GetSignBytes() []byte {
	b, err := msgCdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg.
func (msg MsgSetMemoRegexp) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}
