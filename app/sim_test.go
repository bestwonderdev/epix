package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/EpixZone/epix/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epixconfig "github.com/EpixZone/epix/cmd/config"
	epochstypes "github.com/EpixZone/epix/x/epochs/types"
)

var FlagEnableStreamingValue bool

// Get flags every time the simulator is run
func init() {
	simcli.GetSimulatorFlags()
	flag.BoolVar(&FlagEnableStreamingValue, "EnableStreaming", false, "Enable streaming service")
}

type StoreKeysPrefixes struct {
	A        storetypes.StoreKey
	B        storetypes.StoreKey
	Prefixes [][]byte
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(app *baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

func TestFullAppSimulation(t *testing.T) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = types.TestnetChainID + "-1"

	if simcli.FlagGenesisFileValue == "" {
		sdk.DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000)
	}

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if skip {
		t.Skip("skipping application simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue
	appOptions[flags.FlagChainID] = "epix_1917-1"

	// TODO: shadowed
	epixApp := NewEpix(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, simcli.FlagPeriodValue, true, appOptions, fauxMerkleModeOpt, baseapp.SetChainID(types.TestnetChainID+"-1"))
	require.Equal(t, epixconfig.AppName, epixApp.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		epixApp.BaseApp,
		AppStateFn(epixApp.AppCodec(), epixApp.SimulationManager()),
		RandomAccounts, // replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(epixApp, epixApp.AppCodec(), config),
		epixApp.ModuleAccountAddrs(),
		config,
		epixApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(epixApp, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

func TestAppImportExport(t *testing.T) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = types.TestnetChainID + "-1"

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if skip {
		t.Skip("skipping application import/export simulation")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	sdk.DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000)

	app := NewEpix(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, simcli.FlagPeriodValue, true, appOptions, fauxMerkleModeOpt, baseapp.SetChainID(types.TestnetChainID+"-1"))
	require.Equal(t, epixconfig.AppName, app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		RandomAccounts, // replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	fmt.Println("exporting genesis...")

	exported, err := app.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err)

	fmt.Println("importing genesis...")

	newDB, newDir, _, _, err := simtestutil.SetupSimulation(config, "leveldb-app-sim-2", "Simulation-2", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		newDB.Close()
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewEpix(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, DefaultNodeHome, simcli.FlagPeriodValue, true, appOptions, fauxMerkleModeOpt, baseapp.SetChainID(types.TestnetChainID+"-1"))
	require.Equal(t, epixconfig.AppName, newApp.Name())

	var genesisState GenesisState
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := app.NewContextLegacy(true, tmproto.Header{Height: app.LastBlockHeight()})
	ctxB := newApp.NewContextLegacy(true, tmproto.Header{ChainID: config.ChainID, Height: app.LastBlockHeight()})
	_, err = newApp.ModuleManager.InitGenesis(ctxB, app.AppCodec(), genesisState)

	if err != nil {
		if strings.Contains(err.Error(), "validator set is empty after InitGenesis") {
			logger.Info("Skipping simulation as all validators have been unbonded")
			logger.Info("err", err, "stacktrace", string(debug.Stack()))
			return
		}
	}

	require.NoError(t, err)
	err = newApp.StoreConsensusParams(ctxB, exported.ConsensusParams)
	require.NoError(t, err)

	fmt.Println("comparing stores...")

	// skip certain prefixes
	skipPrefixes := map[string][][]byte{
		stakingtypes.StoreKey: {
			stakingtypes.UnbondingQueueKey, stakingtypes.RedelegationQueueKey, stakingtypes.ValidatorQueueKey,
			stakingtypes.HistoricalInfoKey, stakingtypes.UnbondingIDKey, stakingtypes.UnbondingIndexKey,
			stakingtypes.UnbondingTypeKey, stakingtypes.ValidatorUpdatesKey,
		},
		authzkeeper.StoreKey:   {authzkeeper.GrantQueuePrefix},
		feegrant.StoreKey:      {feegrant.FeeAllowanceQueueKeyPrefix},
		slashingtypes.StoreKey: {slashingtypes.ValidatorMissedBlockBitmapKeyPrefix},
		// In the case of epoch module, the value is updated when importing genesis, so the store consistency is broken
		epochstypes.StoreKey:  {epochstypes.KeyPrefixEpoch},
		upgradetypes.StoreKey: {[]byte{upgradetypes.VersionMapByte}},
	}

	storeKeys := app.GetStoreKeys()
	require.NotEmpty(t, storeKeys)

	for _, appKeyA := range storeKeys {
		// only compare kvstores
		if _, ok := appKeyA.(*storetypes.KVStoreKey); !ok {
			continue
		}

		keyName := appKeyA.Name()
		appKeyB := newApp.GetKey(keyName)
		fmt.Println(appKeyA, appKeyB)

		storeA := ctxA.KVStore(appKeyA)
		storeB := ctxB.KVStore(appKeyB)
		fmt.Println(storeA, storeB)

		failedKVAs, failedKVBs := simtestutil.DiffKVStores(storeA, storeB, skipPrefixes[keyName])
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare %s", keyName)

		fmt.Printf("compared %d different key/value pairs between %s and %s\n", len(failedKVAs), appKeyA, appKeyB)

		require.Equal(t, 0, len(failedKVAs), simtestutil.GetSimulationLog(keyName, app.SimulationManager().StoreDecoders, failedKVAs, failedKVBs))
	}
}

func TestAppStateDeterminism(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = types.TestnetChainID + "-1"

	numSeeds := config.NumBlocks / 10
	numTimesToRunPerSeed := 2
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	// We will be overriding the random seed and just run a single simulation on the provided seed value
	if config.Seed != simcli.DefaultSeedValue {
		numSeeds = 1
	}

	appOptions := viper.New()
	if FlagEnableStreamingValue {
		m := make(map[string]interface{})
		m["streaming.abci.keys"] = []string{"*"}
		m["streaming.abci.plugin"] = "abci_v1"
		m["streaming.abci.stop-node-on-err"] = true
		for key, value := range m {
			appOptions.SetDefault(key, value)
		}
	}
	appOptions.SetDefault(flags.FlagChainID, "epix_1917-1")
	appOptions.SetDefault(flags.FlagHome, DefaultNodeHome)
	appOptions.SetDefault(server.FlagInvCheckPeriod, simcli.FlagPeriodValue)
	if simcli.FlagVerboseValue {
		appOptions.SetDefault(flags.FlagLogLevel, "debug")
	}

	sdk.DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000)

	for i := 0; i < numSeeds; i++ {
		if config.Seed == simcli.DefaultSeedValue {
			config.Seed = rand.Int63()
		}

		fmt.Println("config.Seed: ", config.Seed)

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simcli.FlagVerboseValue {
				logger = log.NewTestLogger(t)
			} else {
				logger = log.NewNopLogger()
			}

			db := dbm.NewMemDB()
			app := NewEpix(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, simcli.FlagPeriodValue, true, appOptions, fauxMerkleModeOpt, baseapp.SetChainID(types.TestnetChainID+"-1"))

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				AppStateFn(app.AppCodec(), app.SimulationManager()),
				RandomAccounts,
				simtestutil.SimulationOperations(app, app.AppCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simtestutil.PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}

func TestAppSimulationAfterImport(t *testing.T) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = types.TestnetChainID + "-1"

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if skip {
		t.Skip("skipping application simulation after import")
	}
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	sdk.DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000)

	app := NewEpix(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, simcli.FlagPeriodValue, true, appOptions, fauxMerkleModeOpt, baseapp.SetChainID(types.TestnetChainID+"-1"))
	require.Equal(t, epixconfig.AppName, app.Name())

	// Run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	if stopEarly {
		fmt.Println("can't export or import a zero-validator genesis, exiting test...")
		return
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(true, []string{}, []string{})
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	newDB, newDir, _, _, err := simtestutil.SetupSimulation(config, "leveldb-app-sim-2", "Simulation-2", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, newDB.Close())
		require.NoError(t, os.RemoveAll(newDir))
	}()

	newApp := NewEpix(log.NewNopLogger(), newDB, nil, true, map[int64]bool{}, DefaultNodeHome, simcli.FlagPeriodValue, true, appOptions, fauxMerkleModeOpt, baseapp.SetChainID(types.TestnetChainID+"-1"))
	require.Equal(t, epixconfig.AppName, newApp.Name())

	newApp.InitChain(&abci.RequestInitChain{
		AppStateBytes: exported.AppState,
		ChainId:       config.ChainID,
	})

	_, _, err = simulation.SimulateFromSeed(
		t,
		os.Stdout,
		newApp.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(newApp, newApp.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)
	require.NoError(t, err)
}
