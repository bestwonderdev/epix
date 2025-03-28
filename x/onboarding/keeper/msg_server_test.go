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
	onboardingtypes "github.com/EpixZone/epix/x/onboarding/types"
)

func (suite *KeeperTestSuite) TestMsgExecutionByProposal() {
	suite.SetupTest()

	// set denom
	stakingParams, err := suite.app.StakingKeeper.GetParams(suite.ctx)
	suite.Require().NoError(err)
	denom := stakingParams.BondDenom

	// change mindeposit for denom
	govParams, err := suite.app.GovKeeper.Params.Get(suite.ctx)
	suite.Require().NoError(err)
	govParams.MinDeposit = []sdk.Coin{sdk.NewCoin(denom, sdkmath.NewInt(1))}
	err = suite.app.GovKeeper.Params.Set(suite.ctx, govParams)
	suite.Require().NoError(err)

	// create account
	privKey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	proposer := sdk.AccAddress(privKey.PubKey().Address().Bytes())

	// deligate to validator
	initAmount := sdkmath.NewInt(int64(math.Pow10(18)) * 2)
	initBalance := sdk.NewCoins(sdk.NewCoin(denom, initAmount))
	testutil.FundAccount(suite.app.BankKeeper, suite.ctx, proposer, initBalance)
	shares, err := suite.app.StakingKeeper.Delegate(suite.ctx, proposer, sdk.DefaultPowerReduction, stakingtypes.Unbonded, suite.validator, true)
	suite.Require().NoError(err)
	suite.Require().True(shares.GT(sdkmath.LegacyNewDec(0)))

	testCases := []struct {
		name      string
		msg       sdk.Msg
		checkFunc func(uint64)
	}{
		{
			"ok - proposal MsgUpdateParams",
			&onboardingtypes.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    onboardingtypes.NewParams(true, sdkmath.NewInt(10000), []string{"channel-0"}),
			},
			func(proposalId uint64) {
				changeParams := onboardingtypes.NewParams(true, sdkmath.NewInt(10000), []string{"channel-0"})

				proposal, err := suite.app.GovKeeper.Proposals.Get(suite.ctx, proposalId)
				suite.Require().NoError(err)
				suite.Require().Equal(govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
				suite.Require().Equal(suite.app.OnboardingKeeper.GetParams(suite.ctx), changeParams)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// submit proposal
			proposal, err := suite.app.GovKeeper.SubmitProposal(suite.ctx, []sdk.Msg{tc.msg}, "", "test", "description", proposer, false)
			suite.Require().NoError(err)
			suite.Commit()

			ok, err := suite.app.GovKeeper.AddDeposit(suite.ctx, proposal.Id, proposer, govParams.MinDeposit)
			suite.Require().NoError(err)
			suite.Require().True(ok)
			suite.Commit()

			err = suite.app.GovKeeper.AddVote(suite.ctx, proposal.Id, proposer, govtypesv1.NewNonSplitVoteOption(govtypesv1.OptionYes), "")
			suite.Require().NoError(err)
			suite.CommitAfter(*govParams.VotingPeriod)

			// check proposal result
			tc.checkFunc(proposal.Id)
		})
	}
}
