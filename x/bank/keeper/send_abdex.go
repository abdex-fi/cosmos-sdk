package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k BaseSendKeeper) SubUnlockedCoins(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	return k.subUnlockedCoins(ctx, addr, amt)
}

func (k BaseSendKeeper) AddCoins(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	return k.addCoins(ctx, addr, amt)
}
