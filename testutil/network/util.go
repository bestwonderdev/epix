package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"

	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/errgroup"

	tmcfg "github.com/cometbft/cometbft/config"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/rpc/client/local"
	"github.com/cometbft/cometbft/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	inflationtypes "github.com/EpixZone/epix/x/inflation/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/evmos/ethermint/server"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

func startInProcess(cfg Config, val *Validator) error {
	logger := val.Ctx.Logger
	tmCfg := val.Ctx.Config
	tmCfg.Instrumentation.Prometheus = false

	if err := val.AppConfig.ValidateBasic(); err != nil {
		return err
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(tmCfg.NodeKeyFile())
	if err != nil {
		return err
	}

	app := cfg.AppConstructor(*val)

	appGenesisProvider := func() (*types.GenesisDoc, error) {
		appGenesis, err := genutiltypes.AppGenesisFromFile(tmCfg.GenesisFile())
		if err != nil {
			return nil, err
		}
		return appGenesis.ToGenesisDoc()
	}

	cmtApp := server.NewCometABCIWrapper(app)
	tmNode, err := node.NewNode(
		tmCfg,
		pvm.LoadOrGenFilePV(tmCfg.PrivValidatorKeyFile(), tmCfg.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(cmtApp),
		appGenesisProvider,
		tmcfg.DefaultDBProvider,
		node.DefaultMetricsProvider(tmCfg.Instrumentation),
		servercmtlog.CometLoggerWrapper{Logger: logger.With("module", val.Moniker)},
	)
	if err != nil {
		return err
	}

	if err := tmNode.Start(); err != nil {
		return err
	}

	val.tmNode = tmNode

	if val.RPCAddress != "" {
		val.RPCClient = local.New(tmNode)
	}

	// We'll need a RPC client if the validator exposes a gRPC or REST endpoint.
	if val.APIAddress != "" || val.AppConfig.GRPC.Enable {
		val.ClientCtx = val.ClientCtx.
			WithClient(val.RPCClient)

		// Add the tx service in the gRPC router.
		app.RegisterTxService(val.ClientCtx)

		// Add the tendermint queries service in the gRPC router.
		app.RegisterTendermintService(val.ClientCtx)

		app.RegisterNodeService(val.ClientCtx, val.AppConfig.Config)
	}

	ctx := context.Background()
	ctx, val.cancelFn = context.WithCancel(ctx)
	val.errGroup, ctx = errgroup.WithContext(ctx)

	grpcCfg := val.AppConfig.GRPC

	if grpcCfg.Enable {
		grpcSrv, err := servergrpc.NewGRPCServer(val.ClientCtx, app, grpcCfg)
		if err != nil {
			return err
		}

		// Start the gRPC server in a goroutine. Note, the provided ctx will ensure
		// that the server is gracefully shut down.
		val.errGroup.Go(func() error {
			return servergrpc.StartGRPCServer(ctx, logger.With("module", "grpc-server"), grpcCfg, grpcSrv)
		})

		val.grpc = grpcSrv
	}

	if val.AppConfig.API.Enable && val.APIAddress != "" {
		apiSrv := api.New(val.ClientCtx, logger.With("module", "api-server"), val.grpc)
		app.RegisterAPIRoutes(apiSrv, val.AppConfig.API)

		val.errGroup.Go(func() error {
			return apiSrv.Start(ctx, val.AppConfig.Config)
		})

		val.api = apiSrv
	}

	if val.AppConfig.JSONRPC.Enable && val.AppConfig.JSONRPC.Address != "" {
		if val.Ctx == nil || val.Ctx.Viper == nil {
			return fmt.Errorf("validator %s context is nil", val.Moniker)
		}

		tmEndpoint := "/websocket"
		tmRPCAddr := val.RPCAddress

		val.jsonrpc, val.jsonrpcDone, err = server.StartJSONRPC(val.Ctx, val.ClientCtx, tmRPCAddr, tmEndpoint, val.AppConfig, nil)
		if err != nil {
			return err
		}

		address := fmt.Sprintf("http://%s", val.AppConfig.JSONRPC.Address)

		val.JSONRPCClient, err = ethclient.Dial(address)
		if err != nil {
			return fmt.Errorf("failed to dial JSON-RPC at %s: %w", val.AppConfig.JSONRPC.Address, err)
		}
	}

	return nil
}

func collectGenFiles(cfg Config, vals []*Validator, outputDir string) error {
	genTime := tmtime.Now()

	for i := 0; i < cfg.NumValidators; i++ {
		tmCfg := vals[i].Ctx.Config

		nodeDir := filepath.Join(outputDir, vals[i].Moniker, "epixd")
		gentxsDir := filepath.Join(outputDir, "gentxs")

		tmCfg.Moniker = vals[i].Moniker
		tmCfg.SetRoot(nodeDir)

		initCfg := genutiltypes.NewInitConfig(cfg.ChainID, gentxsDir, vals[i].NodeID, vals[i].PubKey)

		genFile := tmCfg.GenesisFile()
		appGenesis, err := genutiltypes.AppGenesisFromFile(genFile)
		if err != nil {
			return err
		}

		appState, err := genutil.GenAppStateFromConfig(
			cfg.Codec,
			cfg.TxConfig,
			tmCfg,
			initCfg,
			appGenesis,
			banktypes.GenesisBalancesIterator{},
			genutiltypes.DefaultMessageValidator,
			cfg.TxConfig.SigningContext().ValidatorAddressCodec(),
		)
		if err != nil {
			return err
		}

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, cfg.ChainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func initGenFiles(cfg Config, genAccounts []authtypes.GenesisAccount, genBalances []banktypes.Balance, genFiles []string) error {
	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = append(authGenState.Accounts, accounts...)
	cfg.GenesisState[authtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&authGenState)

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	bankGenState.Balances = genBalances
	cfg.GenesisState[banktypes.ModuleName] = cfg.Codec.MustMarshalJSON(&bankGenState)

	var stakingGenState stakingtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[stakingtypes.ModuleName], &stakingGenState)

	stakingGenState.Params.BondDenom = cfg.BondDenom
	cfg.GenesisState[stakingtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&stakingGenState)

	var govGenState govv1.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[govtypes.ModuleName], &govGenState)

	govGenState.DepositParams.MinDeposit[0].Denom = cfg.BondDenom
	cfg.GenesisState[govtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&govGenState)

	var inflationGenState inflationtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[inflationtypes.ModuleName], &inflationGenState)

	inflationGenState.Params.MintDenom = cfg.BondDenom
	cfg.GenesisState[inflationtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&inflationGenState)

	var crisisGenState crisistypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[crisistypes.ModuleName], &crisisGenState)

	crisisGenState.ConstantFee.Denom = cfg.BondDenom
	cfg.GenesisState[crisistypes.ModuleName] = cfg.Codec.MustMarshalJSON(&crisisGenState)

	var evmGenState evmtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[evmtypes.ModuleName], &evmGenState)

	evmGenState.Params.EvmDenom = cfg.BondDenom
	cfg.GenesisState[evmtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&evmGenState)

	appGenStateJSON, err := json.MarshalIndent(cfg.GenesisState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    cfg.ChainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < cfg.NumValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}

	return nil
}

func WriteFile(name string, dir string, contents []byte) error {
	file := filepath.Join(dir, name)

	err := tmos.EnsureDir(dir, 0o755)
	if err != nil {
		return err
	}

	return tmos.WriteFile(file, contents, 0o644)
}

// Get a free address for a test CometBFT server
// protocol is either tcp, http, etc
func FreeTCPAddr() (addr, port string, closeFn func() error, err error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", "", nil, err
	}

	closeFn = func() error {
		return l.Close()
	}

	portI := l.Addr().(*net.TCPAddr).Port
	port = fmt.Sprintf("%d", portI)
	addr = fmt.Sprintf("tcp://0.0.0.0:%s", port)
	return
}
