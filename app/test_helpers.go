package app

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/EpixZone/epix/cmd/config"
	"github.com/EpixZone/epix/types"
	coinswaptypes "github.com/EpixZone/epix/x/coinswap/types"
	inflationtypes "github.com/EpixZone/epix/x/inflation/types"
)

func init() {
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)
	config.SetBip44CoinType(cfg)
}

// DefaultConsensusParams defines the default Tendermint consensus params used in
// epix testing.
var DefaultConsensusParams = &tmproto.ConsensusParams{
	Block: &tmproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   -1, // no limit
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

func init() {
	feemarkettypes.DefaultMinGasPrice = sdkmath.LegacyZeroDec()
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)
	config.SetBip44CoinType(cfg)
}

// Setup initializes a new epix. A Nop logger is set in epix.
func Setup(
	isCheckTx bool,
	feemarketGenesis *feemarkettypes.GenesisState,
) *Epix {
	db := dbm.NewMemDB()
	app := NewEpix(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, false, simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome), baseapp.SetChainID(types.MainnetChainID+"-1"))
	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewDefaultGenesisState()

		// Verify feeMarket genesis
		if feemarketGenesis != nil {
			if err := feemarketGenesis.Validate(); err != nil {
				panic(err)
			}
			genesisState[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
		}

		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			&abci.RequestInitChain{
				ChainId:         types.MainnetChainID + "-1",
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

func SetupWithGenesisAccounts(genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *Epix {
	app := Setup(false, feemarkettypes.DefaultGenesisState())
	genesisState := NewDefaultGenesisState()
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		totalSupply = totalSupply.Add(b.Coins...)
	}

	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(err)
	}

	app.InitChain(
		&abci.RequestInitChain{
			ChainId:         types.MainnetChainID + "-1",
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	return app
}

// PrintModuleAddresses prints out the bech32 addresses for all module accounts
func PrintModuleAddresses() {
	fmt.Printf("Module Addresses with epix1 prefix:\n")
	fmt.Printf("bonded_tokens_pool: %s\n", authtypes.NewModuleAddress(stakingtypes.BondedPoolName))
	fmt.Printf("not_bonded_tokens_pool: %s\n", authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName))
	fmt.Printf("fee_collector: %s\n", authtypes.NewModuleAddress(authtypes.FeeCollectorName))
	fmt.Printf("distribution: %s\n", authtypes.NewModuleAddress(distrtypes.ModuleName))
	fmt.Printf("gov: %s\n", authtypes.NewModuleAddress(govtypes.ModuleName))
	fmt.Printf("inflation: %s\n", authtypes.NewModuleAddress(inflationtypes.ModuleName))
	fmt.Printf("erc20: %s\n", authtypes.NewModuleAddress(evmtypes.ModuleName))
	fmt.Printf("coinswap: %s\n", authtypes.NewModuleAddress(coinswaptypes.ModuleName))
}
