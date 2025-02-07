package keeper_test

import (
	"math/big"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"
	evm "github.com/evmos/ethermint/x/evm/types"

	"github.com/EpixZone/epix/v8/app"
	epochstypes "github.com/EpixZone/epix/v8/x/epochs/types"
	"github.com/EpixZone/epix/v8/x/inflation/types"

	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	abci "github.com/cometbft/cometbft/abci/types"
)

var denomMint = "aepix"

type KeeperTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	app            *app.Epix
	queryClientEvm evm.QueryClient
	queryClient    types.QueryClient
	consAddress    sdk.ConsAddress
	address        common.Address
	ethSigner      ethtypes.Signer
	validator      stakingtypes.Validator
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

// Test helpers
func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	checkTx := false

	// init app
	suite.app = app.Setup(checkTx, nil)

	pubKey := ed25519.GenPrivKey().PubKey()
	suite.consAddress = sdk.ConsAddress(pubKey.Address())
	// setup context
	suite.ctx = suite.app.BaseApp.NewContextLegacy(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "epix_1916-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	// setup query helpers
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.InflationKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	// Initialize fee market keeper
	feeMarketParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
	feeMarketParams.EnableHeight = 1
	feeMarketParams.NoBaseFee = true // Disable base fee calculation
	suite.app.FeeMarketKeeper.SetParams(suite.ctx, feeMarketParams)

	// Set initial base fee
	bigInt := &big.Int{}
	bigInt.SetUint64(100)
	suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, bigInt)

	// Set EVM params
	evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
	evmParams.EvmDenom = denomMint
	suite.app.EvmKeeper.SetParams(suite.ctx, evmParams)

	// Set staking params
	stakingParams, err := suite.app.StakingKeeper.GetParams(suite.ctx)
	require.NoError(t, err)
	stakingParams.BondDenom = denomMint
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr.String(), pubKey, stakingtypes.Description{})
	require.NoError(t, err)

	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	valbz, err := s.app.StakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	s.NoError(err)
	suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, valbz)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.validator = validator

	suite.ethSigner = ethtypes.LatestSignerForChainID(s.app.EvmKeeper.ChainID())

	// Set epoch start time and height for all epoch identifiers
	identifiers := []string{epochstypes.WeekEpochID, epochstypes.DayEpochID}
	for _, identifier := range identifiers {
		epoch, found := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, identifier)
		suite.Require().True(found)
		epoch.StartTime = suite.ctx.BlockTime()
		epoch.CurrentEpochStartHeight = suite.ctx.BlockHeight()
		suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epoch)
	}
}

func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Nanosecond)
}

// CommitAfter commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	// Get current header
	header := suite.ctx.BlockHeader()

	// Update time
	header.Time = header.Time.Add(t)

	// Begin block with updated time
	suite.ctx = suite.ctx.WithBlockTime(header.Time)
	suite.app.BeginBlocker(suite.ctx)

	// End block
	suite.app.EndBlocker(suite.ctx)

	// Finalize block
	res, err := suite.app.BaseApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.ctx.BlockHeight(),
		Time:   header.Time,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	// Commit block
	_, err = suite.app.BaseApp.Commit()
	suite.Require().NoError(err)

	// Update context for next height
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)

	// Begin new block
	suite.app.BeginBlocker(suite.ctx)

	// Update query helpers
	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.InflationKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
	evm.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelper)
}
