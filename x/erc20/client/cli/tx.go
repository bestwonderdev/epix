package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	addresscodec "cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"

	ethermint "github.com/evmos/ethermint/types"

	"github.com/EpixZone/epix/x/erc20/types"
)

var (
	FlagAuthority = "authority"
)

// NewTxCmd returns a root CLI command handler for erc20 transaction commands
func NewTxCmd(ac addresscodec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "erc20 subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewConvertCoinCmd(),
		NewConvertERC20Cmd(),
		NewRegisterCoinProposalCmd(ac),
		NewRegisterERC20ProposalCmd(ac),
		NewToggleTokenConversionProposalCmd(ac),
	)
	return txCmd
}

// NewConvertCoinCmd returns a CLI command handler for converting a Cosmos coin
func NewConvertCoinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert-coin [coin] [receiver_hex]",
		Short: "Convert a Cosmos coin to ERC20. When the receiver [optional] is omitted, the ERC20 tokens are transferred to the sender.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			var receiver string
			sender := cliCtx.GetFromAddress()

			if len(args) == 2 {
				receiver = args[1]
				if err := ethermint.ValidateAddress(receiver); err != nil {
					return fmt.Errorf("invalid receiver hex address %w", err)
				}
			} else {
				receiver = common.BytesToAddress(sender).Hex()
			}

			msg := &types.MsgConvertCoin{
				Coin:     coin,
				Receiver: receiver,
				Sender:   sender.String(),
			}

			if err := types.ValidateErc20Denom(msg.Coin.Denom); err != nil {
				if err := ibctransfertypes.ValidateIBCDenom(msg.Coin.Denom); err != nil {
					return err
				}
			}

			if !msg.Coin.Amount.IsPositive() {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "cannot mint a non-positive amount")
			}

			_, err = sdk.AccAddressFromBech32(msg.Sender)
			if err != nil {
				return errorsmod.Wrap(err, "invalid sender address")
			}

			if !common.IsHexAddress(msg.Receiver) {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver hex address %s", msg.Receiver)
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewConvertERC20Cmd returns a CLI command handler for converting an ERC20
func NewConvertERC20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert-erc20 [contract-address] [amount] [receiver]",
		Short: "Convert an ERC20 token to Cosmos coin.  When the receiver [optional] is omitted, the Cosmos coins are transferred to the sender.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contract := args[0]
			if err := ethermint.ValidateAddress(contract); err != nil {
				return fmt.Errorf("invalid ERC20 contract address %w", err)
			}

			amount, ok := sdkmath.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid amount %s", args[1])
			}

			from := common.BytesToAddress(cliCtx.GetFromAddress().Bytes())

			receiver := cliCtx.GetFromAddress()
			if len(args) == 3 {
				receiver, err = sdk.AccAddressFromBech32(args[2])
				if err != nil {
					return err
				}
			}

			msg := &types.MsgConvertERC20{
				ContractAddress: contract,
				Amount:          amount,
				Receiver:        receiver.String(),
				Sender:          from.Hex(),
			}

			if !common.IsHexAddress(msg.ContractAddress) {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract hex address '%s'", msg.ContractAddress)
			}

			if !msg.Amount.IsPositive() {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "cannot mint a non-positive amount")
			}

			_, err = sdk.AccAddressFromBech32(msg.Receiver)
			if err != nil {
				return errorsmod.Wrap(err, "invalid receiver address")
			}

			if !common.IsHexAddress(msg.Sender) {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender hex address %s", msg.Sender)
			}

			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewRegisterCoinProposalCmd implements the command to submit a register-coin proposal
func NewRegisterCoinProposalCmd(ac addresscodec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-coin [metadata]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a register coin proposal",
		Long: `Submit a proposal to register a Cosmos coin to the erc20 along with an initial deposit.
Upon passing, the
The proposal details must be supplied via a JSON file.`,
		Example: fmt.Sprintf(`$ %s tx gov submit-proposal register-coin <path/to/metadata.json>

Where metadata.json contains (example):

{
	"description": "The native staking and governance token of the Osmosis chain",
	"denom_units": [
		{
			"denom": "ibc/<HASH>",
			"exponent": 0,
			"aliases": ["ibcuosmo"]
		},
		{
			"denom": "OSMO",
			"exponent": 6
		}
	],
	"base": "ibc/<HASH>",
	"display": "OSMO",
	"name": "Osmo",
	"symbol": "OSMO"
}`, version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := ReadGovPropFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			propMetaData, err := ParseMetadata(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			authority, _ := cmd.Flags().GetString(FlagAuthority)
			if authority != "" {
				if _, err = ac.StringToBytes(authority); err != nil {
					return fmt.Errorf("invalid authority address: %w", err)
				}
			} else {
				authority = sdk.AccAddress(address.Module("gov")).String()
			}

			if err := proposal.SetMsgs([]sdk.Msg{
				&types.MsgRegisterCoin{
					Authority:   authority,
					Title:       proposal.Title,
					Description: proposal.Summary,
					Metadata:    propMetaData,
				},
			}); err != nil {
				return fmt.Errorf("failed to create submit lending market proposal message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	AddGovPropFlagsToCmd(cmd)

	return cmd
}

// NewRegisterERC20ProposalCmd implements the command to submit a register-erc20 proposal
func NewRegisterERC20ProposalCmd(ac addresscodec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-erc20 [erc20-address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a proposal to register an ERC20 token",
		Long:    "Submit a proposal to register an ERC20 token to the erc20 along with an initial deposit.",
		Example: fmt.Sprintf("$ %s tx gov submit-proposal register-erc20 <contract-address>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := ReadGovPropFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			authority, _ := cmd.Flags().GetString(FlagAuthority)
			if authority != "" {
				if _, err = ac.StringToBytes(authority); err != nil {
					return fmt.Errorf("invalid authority address: %w", err)
				}
			} else {
				authority = sdk.AccAddress(address.Module("gov")).String()
			}

			if err := proposal.SetMsgs([]sdk.Msg{
				&types.MsgRegisterERC20{
					Authority:    authority,
					Title:        proposal.Title,
					Description:  proposal.Summary,
					Erc20Address: args[0],
				},
			}); err != nil {
				return fmt.Errorf("failed to create submit lending market proposal message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	AddGovPropFlagsToCmd(cmd)

	return cmd
}

// NewToggleTokenConversionProposalCmd implements the command to submit a toggle-token-conversion proposal
func NewToggleTokenConversionProposalCmd(ac addresscodec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "toggle-token-conversion [token]",
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a toggle token conversion proposal",
		Long:    "Submit a proposal to toggle the conversion of a token pair along with an initial deposit.",
		Example: fmt.Sprintf("$ %s tx gov submit-proposal toggle-token-conversion <denom_or_contract>", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := ReadGovPropFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			authority, _ := cmd.Flags().GetString(FlagAuthority)
			if authority != "" {
				if _, err = ac.StringToBytes(authority); err != nil {
					return fmt.Errorf("invalid authority address: %w", err)
				}
			} else {
				authority = sdk.AccAddress(address.Module("gov")).String()
			}

			if err := proposal.SetMsgs([]sdk.Msg{
				&types.MsgToggleTokenConversion{
					Authority:   authority,
					Title:       proposal.Title,
					Description: proposal.Summary,
					Token:       args[0],
				},
			}); err != nil {
				return fmt.Errorf("failed to create submit lending market proposal message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	AddGovPropFlagsToCmd(cmd)

	return cmd
}
