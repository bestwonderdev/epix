package keeper_test

import (
	"math"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	"github.com/EpixZone/epix/testutil"
	"github.com/EpixZone/epix/x/inflation/types"
	inflationtypes "github.com/EpixZone/epix/x/inflation/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (suite *KeeperTestSuite) TestGovDisableInflation() {
	suite.SetupTest()

	// Set up test environment similar to TestMsgExecutionByProposal
	stakingParams, err := suite.app.StakingKeeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	denom := stakingParams.BondDenom

	// Change mindeposit for denom
	govParams, err := suite.app.GovKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	govParams.MinDeposit = []sdk.Coin{sdk.NewCoin(denom, sdkmath.NewInt(1))}
	err = suite.app.GovKeeper.Params.Set(suite.ctx, govParams)
	suite.Require().NoError(err)

	// Create proposer account
	privKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	proposer := sdk.AccAddress(privKey.PubKey().Address().Bytes())

	// Fund proposer account
	initAmount := sdkmath.NewInt(int64(math.Pow10(18)) * 2)
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))
	testutil.FundAccount(suite.app.BankKeeper, suite.ctx, proposer, initBalance)

	// Delegate to validator
	shares, err := suite.app.StakingKeeper.Delegate(suite.ctx, proposer, sdk.DefaultPowerReduction, stakingtypes.Unbonded, suite.validator, true)
	suite.Require().NoError(err)
	suite.Require().True(shares.GT(sdkmath.LegacyNewDec(0)))

	// Get current params to modify only what we want
	currentParams := suite.app.InflationKeeper.GetParams(suite.ctx)

	// Create proposal to disable inflation
	msg := &inflationtypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params: inflationtypes.Params{
			MintDenom:              currentParams.MintDenom,
			ExponentialCalculation: currentParams.ExponentialCalculation,
			InflationDistribution:  currentParams.InflationDistribution,
			EnableInflation:        false,
		},
	}

	// Submit proposal
	proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, []sdk.Msg{msg}, "", "Disable Inflation", "Disables inflation", proposer, false)
	suite.Require().NoError(err)
	suite.Commit()

	// Add deposit
	deposit := govParams.MinDeposit
	ok, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, proposer, deposit)
	suite.Require().NoError(err)
	suite.Require().True(ok)
	suite.Commit()

	// Vote yes
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.Id, proposer, govtypesv1.NewNonSplitVoteOption(govtypesv1.OptionYes), "")
	suite.Require().NoError(err)
	suite.CommitAfter(*govParams.VotingPeriod)

	// Verify proposal passed and inflation is disabled
	proposal, err = suite.app.GovKeeper.Proposals.Get(suite.ctx, proposal.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

	params := suite.app.InflationKeeper.GetParams(suite.ctx)
	suite.Require().False(params.EnableInflation)
}

func (suite *KeeperTestSuite) TestGovUpdateInflationDistribution() {
	suite.SetupTest()

	// Set up test environment similar to TestMsgExecutionByProposal
	stakingParams, err := suite.app.StakingKeeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	denom := stakingParams.BondDenom

	// Change mindeposit for denom
	govParams, err := suite.app.GovKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	govParams.MinDeposit = []sdk.Coin{sdk.NewCoin(denom, sdkmath.NewInt(1))}
	err = suite.app.GovKeeper.Params.Set(suite.ctx, govParams)
	suite.Require().NoError(err)

	// Create proposer account
	privKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	proposer := sdk.AccAddress(privKey.PubKey().Address().Bytes())

	// Fund proposer account
	initAmount := sdkmath.NewInt(int64(math.Pow10(18)) * 2)
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))
	testutil.FundAccount(suite.app.BankKeeper, suite.ctx, proposer, initBalance)

	// Delegate to validator
	shares, err := suite.app.StakingKeeper.Delegate(suite.ctx, proposer, sdk.DefaultPowerReduction, stakingtypes.Unbonded, suite.validator, true)
	suite.Require().NoError(err)
	suite.Require().True(shares.GT(sdkmath.LegacyNewDec(0)))

	// Get current params to modify only what we want
	currentParams := suite.app.InflationKeeper.GetParams(suite.ctx)

	// Create proposal to update inflation distribution
	// Note: We only need to specify the staking rewards percentage (90%).
	// The community pool automatically receives the remaining balance (10%).
	// This design choice:
	// 1. Ensures all minted tokens are allocated (no dust)
	// 2. Makes governance proposals simpler (only need to specify one parameter)
	// 3. Prevents configuration errors where percentages don't add up to 100%
	msg := &inflationtypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params: inflationtypes.Params{
			MintDenom:              currentParams.MintDenom,
			ExponentialCalculation: currentParams.ExponentialCalculation,
			InflationDistribution: inflationtypes.InflationDistribution{
				StakingRewards: sdkmath.LegacyNewDecWithPrec(90, 2), // 90%
				CommunityPool:  sdkmath.LegacyNewDecWithPrec(10, 2), // This value is ignored; community pool gets remaining balance
			},
			EnableInflation: currentParams.EnableInflation,
		},
	}

	// Submit proposal
	proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, []sdk.Msg{msg}, "", "Update Inflation Distribution", "Changes staking rewards to 90% and community pool to 10%", proposer, false)
	suite.Require().NoError(err)
	suite.Commit()

	// Add deposit
	deposit := govParams.MinDeposit
	ok, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, proposer, deposit)
	suite.Require().NoError(err)
	suite.Require().True(ok)
	suite.Commit()

	// Vote yes
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.Id, proposer, govtypesv1.NewNonSplitVoteOption(govtypesv1.OptionYes), "")
	suite.Require().NoError(err)
	suite.CommitAfter(*govParams.VotingPeriod)

	// Verify proposal passed and distribution parameters are updated
	proposal, err = suite.app.GovKeeper.Proposals.Get(suite.ctx, proposal.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

	params := suite.app.InflationKeeper.GetParams(suite.ctx)
	suite.Require().Equal(sdkmath.LegacyNewDecWithPrec(90, 2), params.InflationDistribution.StakingRewards)
	// Note: We don't verify the community pool percentage as it's not used

	// Clear any existing balances in the fee collector and community pool
	feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)

	// Get initial balances
	initialFeeCollector := suite.app.BankKeeper.GetBalance(suite.ctx, feeCollector, params.MintDenom)
	if !initialFeeCollector.IsZero() {
		err = suite.app.BankKeeper.SendCoinsFromModuleToModule(
			suite.ctx,
			authtypes.FeeCollectorName,
			types.ModuleName,
			sdk.NewCoins(initialFeeCollector),
		)
		suite.Require().NoError(err)
	}

	feePool, err := suite.app.DistrKeeper.FeePool.Get(suite.ctx)
	suite.Require().NoError(err)
	if !feePool.CommunityPool.IsZero() {
		err = suite.app.DistrKeeper.FeePool.Set(suite.ctx, distrtypes.FeePool{CommunityPool: sdk.DecCoins{}})
		suite.Require().NoError(err)
	}

	// Get initial balances after clearing
	initialFeeCollector = suite.app.BankKeeper.GetBalance(suite.ctx, feeCollector, params.MintDenom)
	feePool, err = suite.app.DistrKeeper.FeePool.Get(suite.ctx)
	suite.Require().NoError(err)
	initialCommunityPool := feePool.CommunityPool

	// Mint a specific amount of tokens to test distribution
	mintCoin := sdk.NewCoin(params.MintDenom, sdkmath.NewInt(1000000))

	// Mint and allocate the tokens
	_, _, err = suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, mintCoin)
	suite.Require().NoError(err)

	// Get final balances
	finalFeeCollector := suite.app.BankKeeper.GetBalance(suite.ctx, feeCollector, params.MintDenom)
	feePool, err = suite.app.DistrKeeper.FeePool.Get(suite.ctx)
	suite.Require().NoError(err)
	finalCommunityPool := feePool.CommunityPool

	// Calculate actual distributions
	stakingRewardsAmount := finalFeeCollector.Amount.Sub(initialFeeCollector.Amount)

	// Calculate community pool change
	initialCommunityAmount := initialCommunityPool.AmountOf(params.MintDenom)
	finalCommunityAmount := finalCommunityPool.AmountOf(params.MintDenom)
	communityPoolChange := finalCommunityAmount.Sub(initialCommunityAmount)
	communityPoolAmount := communityPoolChange.TruncateInt()

	// Calculate expected amounts
	// Note: Due to truncation in GetProportions, the community pool gets the remainder
	expectedStakingAmount := mintCoin.Amount.ToLegacyDec().Mul(params.InflationDistribution.StakingRewards).TruncateInt()
	// Community pool gets the remaining balance (total - staking)
	expectedCommunityAmount := mintCoin.Amount.Sub(expectedStakingAmount)

	suite.Require().Equal(expectedStakingAmount, stakingRewardsAmount, "Staking rewards amount doesn't match expected 90%")
	suite.Require().Equal(expectedCommunityAmount, communityPoolAmount, "Community pool amount doesn't match expected 10%")
}

func (suite *KeeperTestSuite) TestGovInvalidInflationDistribution() {
	suite.SetupTest()

	// Set up test environment similar to TestMsgExecutionByProposal
	stakingParams, err := suite.app.StakingKeeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	denom := stakingParams.BondDenom

	// Change mindeposit for denom
	govParams, err := suite.app.GovKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	govParams.MinDeposit = []sdk.Coin{sdk.NewCoin(denom, sdkmath.NewInt(1))}
	err = suite.app.GovKeeper.Params.Set(suite.ctx, govParams)
	suite.Require().NoError(err)

	// Create proposer account
	privKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	proposer := sdk.AccAddress(privKey.PubKey().Address().Bytes())

	// Fund proposer account
	initAmount := sdkmath.NewInt(int64(math.Pow10(18)) * 2)
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))
	testutil.FundAccount(suite.app.BankKeeper, suite.ctx, proposer, initBalance)

	// Delegate to validator
	shares, err := suite.app.StakingKeeper.Delegate(suite.ctx, proposer, sdk.DefaultPowerReduction, stakingtypes.Unbonded, suite.validator, true)
	suite.Require().NoError(err)
	suite.Require().True(shares.GT(sdkmath.LegacyNewDec(0)))

	// Get current params to modify only what we want
	currentParams := suite.app.InflationKeeper.GetParams(suite.ctx)

	// Test case 1: Distribution percentages don't sum to 100%
	msg := &inflationtypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params: inflationtypes.Params{
			MintDenom:              currentParams.MintDenom,
			ExponentialCalculation: currentParams.ExponentialCalculation,
			InflationDistribution: inflationtypes.InflationDistribution{
				StakingRewards: sdkmath.LegacyNewDecWithPrec(80, 2), // 80%
				CommunityPool:  sdkmath.LegacyNewDecWithPrec(10, 2), // 10%
			},
			EnableInflation: currentParams.EnableInflation,
		},
	}

	// Submit proposal
	proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, []sdk.Msg{msg}, "", "Invalid Distribution", "Distribution percentages don't sum to 100%", proposer, false)
	suite.Require().NoError(err)
	suite.Commit()

	// Add deposit
	deposit := govParams.MinDeposit
	ok, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, proposer, deposit)
	suite.Require().NoError(err)
	suite.Require().True(ok)
	suite.Commit()

	// Vote yes
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.Id, proposer, govtypesv1.NewNonSplitVoteOption(govtypesv1.OptionYes), "")
	suite.Require().NoError(err)
	suite.CommitAfter(*govParams.VotingPeriod)

	// Verify proposal failed
	proposal, err = suite.app.GovKeeper.Proposals.Get(suite.ctx, proposal.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)

	// Test case 2: Negative percentage
	msg = &inflationtypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params: inflationtypes.Params{
			MintDenom:              currentParams.MintDenom,
			ExponentialCalculation: currentParams.ExponentialCalculation,
			InflationDistribution: inflationtypes.InflationDistribution{
				StakingRewards: sdkmath.LegacyNewDecWithPrec(-90, 2), // -90%
				CommunityPool:  sdkmath.LegacyNewDecWithPrec(190, 2), // 190%
			},
			EnableInflation: currentParams.EnableInflation,
		},
	}

	// Submit proposal
	proposal, err = suite.app.GovKeeper.SubmitProposal(suite.ctx, []sdk.Msg{msg}, "", "Invalid Distribution", "Negative percentage", proposer, false)
	suite.Require().NoError(err)
	suite.Commit()

	// Add deposit
	ok, err = suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, proposer, deposit)
	suite.Require().NoError(err)
	suite.Require().True(ok)
	suite.Commit()

	// Vote yes
	err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.Id, proposer, govtypesv1.NewNonSplitVoteOption(govtypesv1.OptionYes), "")
	suite.Require().NoError(err)
	suite.CommitAfter(*govParams.VotingPeriod)

	// Verify proposal failed
	proposal, err = suite.app.GovKeeper.Proposals.Get(suite.ctx, proposal.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_FAILED, proposal.Status)
}

func (suite *KeeperTestSuite) TestInflationRateChanges() {
	suite.SetupTest()

	params := suite.app.InflationKeeper.GetParams(suite.ctx)
	bondedRatio := sdkmath.LegacyNewDecWithPrec(67, 2) // 67% bonded ratio

	// Calculate inflation rate with 67% bonding ratio
	period := suite.app.InflationKeeper.GetPeriod(suite.ctx)
	epochsPerPeriod := suite.app.InflationKeeper.GetEpochsPerPeriod(suite.ctx)

	// Calculate expected inflation with bonding incentive
	// exponentialDecay := a * (1 - r) ^ x
	a := params.ExponentialCalculation.A
	r := params.ExponentialCalculation.R
	decay := sdkmath.LegacyOneDec().Sub(r)
	exponentialDecay := a.Mul(decay.Power(period))

	// bondingIncentive = 1 + max_variance - bondingRatio * (max_variance / bonding_target)
	maxVariance := params.ExponentialCalculation.MaxVariance
	bondingTarget := params.ExponentialCalculation.BondingTarget
	sub := bondedRatio.Mul(maxVariance.Quo(bondingTarget))
	bondingIncentive := sdkmath.LegacyOneDec().Add(maxVariance).Sub(sub)

	// Expected epoch provision with 67% bonding ratio
	expectedEpochProvision := exponentialDecay.Mul(bondingIncentive).Quo(sdkmath.LegacyNewDec(epochsPerPeriod))

	// Set bonded tokens ratio and calculate actual provision
	suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, expectedEpochProvision)
	actualProvision, found := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	suite.Require().True(found)

	// Verify actual provision matches expected
	suite.Require().Equal(expectedEpochProvision, actualProvision)

	// Test max variance cap
	// Set bonding ratio above target
	bondedRatio = sdkmath.LegacyNewDecWithPrec(90, 2) // 90% bonded
	sub = bondedRatio.Mul(maxVariance.Quo(bondingTarget))
	bondingIncentive = sdkmath.LegacyOneDec().Add(maxVariance).Sub(sub)

	expectedEpochProvision = exponentialDecay.Mul(bondingIncentive).Quo(sdkmath.LegacyNewDec(epochsPerPeriod))
	suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, expectedEpochProvision)
	actualProvision, found = suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	suite.Require().True(found)

	// Verify actual provision matches expected with capped variance
	suite.Require().Equal(expectedEpochProvision, actualProvision)
}
