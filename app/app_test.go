package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-db"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/state"
	cbfttypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	"cosmossdk.io/store/iavl"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/EpixZone/epix/types"
	coinswaptypes "github.com/EpixZone/epix/x/coinswap/types"
)

func TestEpixExport(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewEpix(
		log.NewLogger(os.Stdout),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		0,
		false,
		simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
		baseapp.SetChainID(types.TestnetChainID+"-1"),
	)

	genesisState := NewDefaultGenesisState()
	stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
	require.NoError(t, err)

	// Initialize the chain
	app.InitChain(
		&abci.RequestInitChain{
			ChainId:         types.TestnetChainID + "-1",
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
	})
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewEpix(
		log.NewLogger(os.Stdout),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		0,
		false,
		simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
		baseapp.SetChainID(types.TestnetChainID+"-1"),
	)
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func TestPrintModuleAddresses(t *testing.T) {
	PrintModuleAddresses()
}

// TestWorkingHash tests that the working hash of the IAVL store is calculated correctly during the initialization phase of the genesis, given the initial height specified in the GenesisDoc.
func TestWorkingHash(t *testing.T) {
	gdoc, err := state.MakeGenesisDocFromFile("test-genesis.json")
	require.NoError(t, err)

	gs, err := state.MakeGenesisState(gdoc)
	require.NoError(t, err)

	fmt.Printf("Genesis Time: %v\n", gdoc.GenesisTime)
	fmt.Printf("Chain ID: %s\n", gdoc.ChainID)
	fmt.Printf("Initial Height: %d\n", gdoc.InitialHeight)

	tmpDir := "test-working-hash"
	db, err := dbm.NewGoLevelDB("test", tmpDir, nil)
	require.NoError(t, err)
	app := NewEpix(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, false, simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome), baseapp.SetChainID(gdoc.ChainID))

	// delete tmpDir
	defer require.NoError(t, os.RemoveAll(tmpDir))

	pbparams := gdoc.ConsensusParams.ToProto()

	// Use genesis time for all time-related fields to ensure determinism
	blockTime := gdoc.GenesisTime
	fmt.Printf("Block Time: %v\n", blockTime)

	// Initialize the chain
	_, err = app.InitChain(&abci.RequestInitChain{
		Time:            blockTime,
		ChainId:         gdoc.ChainID,
		ConsensusParams: &pbparams,
		Validators:      cbfttypes.TM2PB.ValidatorUpdates(gs.Validators),
		AppStateBytes:   gdoc.AppState,
		InitialHeight:   gdoc.InitialHeight,
	})
	require.NoError(t, err)

	// Call FinalizeBlock to calculate each module's working hash.
	// Without calling this, all module's root node will have empty hash.
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: gdoc.InitialHeight,
		Time:   blockTime,
	})
	require.NoError(t, err)

	storeKeys := app.GetStoreKeys()
	// Print all store keys first
	fmt.Println("\nAll Store Keys:")
	for _, key := range storeKeys {
		if key != nil {
			fmt.Printf("- %s\n", key.Name())
		}
	}

	fmt.Println("\nWorking Hashes:")
	// deterministicKeys are module keys which has always same working hash whenever run this test. (non deterministic module: staking, epoch, inflation)
	deterministicKeys := []string{
		authtypes.StoreKey, banktypes.StoreKey, capabilitytypes.StoreKey, coinswaptypes.StoreKey,
		consensustypes.StoreKey, crisistypes.StoreKey, distrtypes.StoreKey,
		evmtypes.StoreKey, feemarkettypes.StoreKey, govtypes.StoreKey, ibctransfertypes.StoreKey,
		paramstypes.StoreKey, slashingtypes.StoreKey, upgradetypes.StoreKey}

	// workingHashWithZeroInitialHeight is the working hash of the IAVL store with initial height 0 with given genesis.
	workingHashWithZeroInitialHeight := map[string]string{
		authtypes.StoreKey:        "3d70457791f9da7db97ef1ced7334ca6303e62af3655e3d436c25404f2c47f41",
		banktypes.StoreKey:        "3050131cafbf54eb00192a628f6bfa07fbd39ceb960aaac416b5c49dbab429d1",
		capabilitytypes.StoreKey:  "e9261548b1c687638721f75920e1c8e3f4f52cbd7ab3aeddc6c626cd8abc8718",
		coinswaptypes.StoreKey:    "3d307d35fcb33a4f4892f763c630f1b209b5a7cd360255943ab3217150a0aa3e",
		consensustypes.StoreKey:   "2573460b2ef3a08c825bdfb485e51680038530f70c45f0b723167ae09599761c",
		crisistypes.StoreKey:      "2723146982236ef56bc903407e54819a0c6508837ed201dd887a57d4a803f1aa",
		distrtypes.StoreKey:       "65cfb5c307b023ed80255b3ffc14e08b39cc00ecd7b710ff8b5bc96d1efdbda6",
		evmtypes.StoreKey:         "842c2cd4989bf517916822d54d848aaf8c4c1550b0226ec98d09d7f5dc8db9c2",
		feemarkettypes.StoreKey:   "b8a66cce8e7809f521db9fdd71bfeb980966f1b74ef252bda65804d8b89da7de",
		govtypes.StoreKey:         "52de365ff0f0b784cf2c1583fb7f9dc816cb1e630644d481825bf714b9d7ba53",
		ibctransfertypes.StoreKey: "3ffd548eb86288efc51964649e36dc710f591c3d60d6f9c1b42f2a4d17870904",
		paramstypes.StoreKey:      "b6ab1016e05113fe581360a4f01af87ec1375a00280d23a7e7e63e225af72ce4",
		slashingtypes.StoreKey:    "be189bc6a79688b8c051044c855fb659b7adc3904c354de89f2688328c22b1f8",
		upgradetypes.StoreKey:     "9677219870ca98ba9868589ccdcd97411d9b82abc6a7aa1910016457b4730a93",
	}

	matchAny := func(key string) bool {
		for _, dk := range deterministicKeys {
			if dk == key {
				return true
			}
		}
		return false
	}

	for _, key := range storeKeys {
		if key != nil && matchAny(key.Name()) {
			kvstore := app.CommitMultiStore().GetCommitKVStore(key)
			require.Equal(t, storetypes.StoreTypeIAVL, kvstore.GetStoreType())
			iavlStore, ok := kvstore.(*iavl.Store)
			require.True(t, ok)
			workingHash := hex.EncodeToString(iavlStore.WorkingHash())
			fmt.Printf("Module: %s\n", key.Name())
			fmt.Printf("  Expected: %s\n", workingHashWithZeroInitialHeight[key.Name()])
			fmt.Printf("  Actual:   %s\n", workingHash)
			require.Equal(t, workingHashWithZeroInitialHeight[key.Name()], workingHash, key.Name())
		}
	}
}

func TestPrintEthAccountAddress(t *testing.T) {
	// The old address with invalid checksum
	oldAddress := "epix1jfdykyt4hhmwhegh669c6exnjtfr3yparej8y8"

	// Extract the human readable part and data part
	parts := strings.Split(oldAddress, "1")
	require.Equal(t, 2, len(parts))

	// The data part without checksum
	data := parts[1][:len(parts[1])-6]
	fmt.Printf("Data without checksum: %s\n", data)

	// Construct the correct address
	correctAddress := fmt.Sprintf("epix1%syyazh5", data)
	fmt.Printf("Correct address: %s\n", correctAddress)

	// Verify the address is valid
	_, err := sdk.AccAddressFromBech32(correctAddress)
	require.NoError(t, err)
}

func TestPrintEVMAddresses(t *testing.T) {
	// Convert hex addresses to bech32
	hexAddresses := []string{
		"0x925a4b1175Bdf6EBE517d68B8D64D392D238903D",
		"0x9dDc0FF8D6D4a81f73B8c7067E218f3A7e8d3675",
	}

	for _, hexAddr := range hexAddresses {
		// Remove 0x prefix
		hexAddr = strings.TrimPrefix(hexAddr, "0x")
		// Convert to bytes
		addrBytes, err := hex.DecodeString(hexAddr)
		require.NoError(t, err)
		// Convert to bech32
		bech32Addr := sdk.AccAddress(addrBytes)
		fmt.Printf("Hex address: %s\nBech32 address: %s\n\n", hexAddr, bech32Addr.String())
	}
}

func TestPrintConsensusAddress(t *testing.T) {
	// Convert hex consensus address to bech32
	hexAddr := "CFF57A76F8200F0358989D9ABED90F9B8E2F5D80"
	addrBytes, err := hex.DecodeString(hexAddr)
	require.NoError(t, err)
	// Convert to bech32
	bech32Addr := sdk.ConsAddress(addrBytes)
	fmt.Printf("Hex address: %s\nBech32 address: %s\n\n", hexAddr, bech32Addr.String())
}
