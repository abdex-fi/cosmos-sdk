package types

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ SendHooks = MultiSendHooks{}

// SendHooks defines the interface for bank module hooks
type SendHooks interface {
	BeforeSend(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type SendHooksWrapper struct{ SendHooks }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (SendHooksWrapper) IsOnePerModuleType() {}

// MultiSendHooks combines multiple BankHooks
type MultiSendHooks []SendHooks

// NewMultiSendHooks creates a new MultiBankHooks
func NewMultiSendHooks(hooks ...SendHooks) MultiSendHooks {
	return hooks
}

func (h MultiSendHooks) BeforeSend(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	var errs error
	for _, hook := range h {
		errs = errors.Join(errs, hook.BeforeSend(ctx, fromAddr, toAddr, amt))
	}

	return errs
}
