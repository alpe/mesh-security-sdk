package keeper

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func TestSetVirtualStakingMaxCap(t *testing.T) {
	pCtx, keepers := CreateDefaultTestInput(t)
	k := keepers.MeshKeeper
	myContract := sdk.AccAddress(rand.Bytes(32))
	myAmount := sdk.NewInt64Coin(keepers.StakingKeeper.BondDenom(pCtx), 123)

	k.wasm = MockWasmKeeper{HasContractInfoFn: func(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
		return contractAddress.Equals(myContract)
	}}

	ctx, _ := pCtx.CacheContext()
	specs := map[string]struct {
		src      types.MsgSetVirtualStakingMaxCap
		setup    func(ctx sdk.Context)
		expErr   bool
		expLimit sdk.Coin
	}{
		"limit stored with scheduler for existing contract": {
			src: types.MsgSetVirtualStakingMaxCap{
				Authority: k.GetAuthority(),
				Contract:  myContract.String(),
				MaxCap:    myAmount,
			},
			expLimit: myAmount,
		},
		"fails for non existing contract": {
			src: types.MsgSetVirtualStakingMaxCap{
				Authority: k.GetAuthority(),
				Contract:  sdk.AccAddress(rand.Bytes(32)).String(),
				MaxCap:    myAmount,
			},
			expErr: true,
		},
		"unauthorized rejected": {
			src: types.MsgSetVirtualStakingMaxCap{
				Authority: myContract.String(),
				Contract:  myContract.String(),
				MaxCap:    myAmount,
			},
			expErr: true,
		},
		"invalid data rejected": {
			src:    types.MsgSetVirtualStakingMaxCap{},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			m := NewMsgServer(k)

			// when
			gotRsp, gotErr := m.SetVirtualStakingMaxCap(sdk.WrapSDKContext(ctx), &spec.src)

			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.NotNil(t, gotRsp)
			require.True(t, k.HasMaxCapLimit(ctx, myContract))
			assert.Equal(t, spec.expLimit, k.GetMaxCapLimit(ctx, myContract))
			// and scheduled
			assert.True(t, k.HasScheduledTask(ctx, types.SchedulerTaskRebalance, myContract))
		})
	}
}
