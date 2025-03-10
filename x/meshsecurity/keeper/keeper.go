package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec
	bank     types.XBankKeeper
	staking  types.XStakingKeeper
	wasm     types.WasmKeeper
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper constructor with vanilla sdk keepers
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bank types.SDKBankKeeper,
	staking types.SDKStakingKeeper,
	wasm types.WasmKeeper,
	authority string,
) *Keeper {
	return NewKeeperX(cdc, storeKey, NewBankKeeperAdapter(bank), NewStakingKeeperAdapter(staking, bank), wasm, authority)
}

// NewKeeperX constructor with extended Osmosis SDK keepers
func NewKeeperX(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bank types.XBankKeeper,
	staking types.XStakingKeeper,
	wasm types.WasmKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		bank:      bank,
		staking:   staking,
		wasm:      wasm,
		authority: authority,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// HasMaxCapLimit returns true when any max cap limit was set. The amount is not taken into account for the result.
// A 0 value would be true as well.
func (k Keeper) HasMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BuildMaxCapLimitKey(actor))
}

// GetMaxCapLimit the cap limit is set per consumer contract. Different providers can have different limits
// Returns zero amount when no limit is stored.
func (k Keeper) GetMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
	return sdk.NewCoin(k.staking.BondDenom(ctx), k.mustLoadInt(ctx, k.storeKey, types.BuildMaxCapLimitKey(actor)))
}

// SetMaxCapLimit stores the max cap limit for the given contract address.
// Any existing limit for this contract will be overwritten
func (k Keeper) SetMaxCapLimit(ctx sdk.Context, contract sdk.AccAddress, newAmount sdk.Coin) error {
	if k.staking.BondDenom(ctx) != newAmount.Denom {
		return sdkerrors.ErrInvalidCoins
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := newAmount.Amount.Marshal()
	if err != nil { // always nil
		return errorsmod.Wrap(err, "marshal amount")
	}
	store.Set(types.BuildMaxCapLimitKey(contract), bz)
	return nil
}

// GetTotalDelegated returns the total amount delegated by the given consumer contract.
// This amount can be 0 is never negative.
func (k Keeper) GetTotalDelegated(ctx sdk.Context, actor sdk.AccAddress) sdk.Coin {
	v := k.mustLoadInt(ctx, k.storeKey, types.BuildTotalDelegatedAmountKey(actor))
	if v.IsNegative() {
		v = math.ZeroInt()
	}
	return sdk.NewCoin(k.staking.BondDenom(ctx), v)
}

// internal setter. must only be used with bonding token denom or panics
func (k Keeper) setTotalDelegated(ctx sdk.Context, actor sdk.AccAddress, newAmount sdk.Coin) {
	if k.staking.BondDenom(ctx) != newAmount.Denom {
		panic(sdkerrors.ErrInvalidCoins.Wrapf("not a staking denom: %s", newAmount.Denom))
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := newAmount.Amount.Marshal()
	if err != nil { // always nil
		panic(err)
	}
	store.Set(types.BuildTotalDelegatedAmountKey(actor), bz)
}

// helper to deserialize a math.Int from store. Returns zero when key does not exist.
// Panics when Unmarshal fails
func (k Keeper) mustLoadInt(ctx sdk.Context, storeKey storetypes.StoreKey, key []byte) math.Int {
	store := ctx.KVStore(storeKey)
	bz := store.Get(key)
	if bz == nil {
		return sdk.ZeroInt()
	}
	var r math.Int
	if err := r.Unmarshal(bz); err != nil {
		panic(err)
	}
	return r
}

// IterateMaxCapLimit iterate over contract addresses with max cap limt set
// Callback can return true to stop early
func (k Keeper) IterateMaxCapLimit(ctx sdk.Context, cb func(sdk.AccAddress, math.Int) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.MaxCapLimitKeyPrefix)
	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var r math.Int
		if err := r.Unmarshal(iter.Value()); err != nil {
			panic(err)
		}
		// cb returns true to stop early
		if cb(iter.Key(), r) {
			return
		}
	}
}

// ModuleLogger returns logger with module attribute
func ModuleLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
