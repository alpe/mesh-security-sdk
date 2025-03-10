package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the mesh security module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}
	queryCmd.AddCommand(
		GetCmdQueryMaxCapLimit(),
		GetCmdQueryMaxCapLimits(),
	)
	return queryCmd
}

// GetCmdQueryMaxCapLimit implements a command to return the current
// max cap limit for the given contract.
func GetCmdQueryMaxCapLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "max-cap-limit [address]",
		Short: "Query the current max cap limit for the given contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			req := &types.QueryVirtualStakingMaxCapLimitRequest{
				Address: args[0],
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.VirtualStakingMaxCapLimit(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryMaxCapLimit implements a command to return the current
// max cap limit for each contract.
func GetCmdQueryMaxCapLimits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "max-cap-limits",
		Short: "Query the current max cap limit for each contract",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			req := &types.QueryVirtualStakingMaxCapLimitsRequest{}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.VirtualStakingMaxCapLimits(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
		SilenceUsage: true,
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
