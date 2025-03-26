package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	epochstypes "github.com/EpixZone/epix/x/epochs/types"
	"github.com/EpixZone/epix/x/inflation/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	epochNumber int64
	skipped     uint64
	provision   sdkmath.LegacyDec
	found       bool
)

var _ = Describe("Inflation", Ordered, func() {
	BeforeEach(func() {
		s.SetupTest()
	})

	Describe("Commiting a block", func() {
		Context("with inflation param enabled", func() {
			BeforeEach(func() {
				// Set inflation parameters
				params := s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = true
				params.MintDenom = denomMint
				// Set staking rewards to 10% - the community pool will automatically
				// receive the remaining 90% of minted tokens. This tests the default
				// case where most tokens go to the community pool.
				// Note: The CommunityPool value is not directly used; instead, the
				// module sends all remaining balance to the community pool. This
				// design ensures no tokens are left behind and makes the system
				// more robust against configuration errors.
				params.InflationDistribution.StakingRewards = sdkmath.LegacyNewDecWithPrec(10, 2) // 10%
				params.InflationDistribution.CommunityPool = sdkmath.LegacyNewDecWithPrec(90, 2)  // 90%
				s.app.InflationKeeper.SetParams(s.ctx, params)

				// Verify params are set correctly
				setParams := s.app.InflationKeeper.GetParams(s.ctx)
				s.Require().Equal(denomMint, setParams.MintDenom)
				s.Require().True(setParams.EnableInflation)

				fmt.Printf("1. Set inflation Params \n")
				fmt.Printf("MintDenom %s\n", setParams.MintDenom)
				fmt.Printf("ExponentialCalculation param A %d\n", params.ExponentialCalculation.A)
				fmt.Printf("ExponentialCalculation param R %d\n", params.ExponentialCalculation.R)
				fmt.Printf("ExponentialCalculation param C %d\n", params.ExponentialCalculation.C)

				// Set epoch identifier
				s.app.InflationKeeper.SetEpochIdentifier(s.ctx, epochstypes.DayEpochID)

				// Set epochs per period
				s.app.InflationKeeper.SetEpochsPerPeriod(s.ctx, 30)

				// Set initial epoch mint provision

				fmt.Printf("EpochIdentifier %s\n", epochstypes.DayEpochID)
				fmt.Printf("EpochsPerPeriod 30\n")

				genesisProvision := sdkmath.LegacyMustNewDecFromStr("96666666666666666666667")
				s.app.InflationKeeper.SetEpochMintProvision(s.ctx, genesisProvision)

				// Set initial epoch info with matching identifier
				epochInfo := epochstypes.EpochInfo{
					Identifier:            epochstypes.DayEpochID,
					StartTime:             s.ctx.BlockTime(),
					Duration:              time.Hour * 24,
					CurrentEpoch:          0,
					CurrentEpochStartTime: s.ctx.BlockTime(),
					EpochCountingStarted:  true,
				}
				s.app.EpochsKeeper.SetEpochInfo(s.ctx, epochInfo)

				// Verify epoch identifier matches
				expIdentifier := s.app.InflationKeeper.GetEpochIdentifier(s.ctx)
				s.Require().Equal(epochstypes.DayEpochID, expIdentifier)

				// Setup staking module with bonded tokens
				s.clearValidatorsAndInitPool(1000)
				valAddrs := MakeValAccts(1)
				pk := GenKeys(1)
				v, err := stakingtypes.NewValidator(valAddrs[0].String(), pk[0].PubKey(), stakingtypes.Description{})
				s.Require().NoError(err)
				tokens := s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, 1000)
				v, _ = v.AddTokensFromDel(tokens)
				s.app.StakingKeeper.SetValidator(s.ctx, v)
				s.app.StakingKeeper.SetValidatorByPowerIndex(s.ctx, v)

				// Bond tokens to the validator
				bondDenom, err := s.app.StakingKeeper.BondDenom(s.ctx)
				s.Require().NoError(err)
				s.app.BankKeeper.MintCoins(s.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(bondDenom, tokens)))
				s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, types.ModuleName, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(bondDenom, tokens)))

				// Update validator status to bonded
				v = v.UpdateStatus(stakingtypes.Bonded)
				s.app.StakingKeeper.SetValidator(s.ctx, v)

				// Update validator set
				_, err = s.app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
				s.Require().NoError(err)

				// Initialize fee collector account
				feeCollector := s.app.AccountKeeper.GetModuleAddress("fee_collector")
				s.Require().NotNil(feeCollector)

				// Initialize the module account
				moduleAcc := s.app.AccountKeeper.GetModuleAccount(s.ctx, "fee_collector")
				s.app.AccountKeeper.SetModuleAccount(s.ctx, moduleAcc)

				// Commit blocks to trigger epoch
				futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Hour * 24))
				newHeight := s.app.LastBlockHeight() + 1

				// Manually trigger epoch hooks for first epoch
				s.app.EpochsKeeper.BeforeEpochStart(futureCtx, epochstypes.DayEpochID, newHeight)
				s.app.EpochsKeeper.AfterEpochEnd(futureCtx, epochstypes.DayEpochID, newHeight)

				// Update epoch info to reflect the completed epoch
				epochInfo.CurrentEpoch = 1
				epochInfo.CurrentEpochStartTime = futureCtx.BlockTime()
				s.app.EpochsKeeper.SetEpochInfo(futureCtx, epochInfo)
			})

			It("should allocate staking rewards to the fee collector", func() {
				// Get fee collector balance
				feeCollector := s.app.AccountKeeper.GetModuleAddress("fee_collector")
				feeCollectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollector, denomMint)
				feeCollectorDec := feeCollectorBalance.Amount.ToLegacyDec()

				// Get community pool balance
				feePool, err := s.app.DistrKeeper.FeePool.Get(s.ctx)
				s.Require().NoError(err)
				communityPoolBalance := feePool.CommunityPool.AmountOf(denomMint)

				// Verify staking rewards in fee collector (10% of total minted)
				// Note: Truncation occurs when calculating staking rewards
				expectedStakingDec := sdkmath.LegacyMustNewDecFromStr("9666666666666666666666")
				s.Require().Equal(expectedStakingDec, feeCollectorDec)
				fmt.Printf("After one epoch expectedStakingDec: feeCollectorDec %s:%s\n", expectedStakingDec, feeCollectorDec)
				// Verify community pool allocation (remaining balance after staking rewards)
				// Note: Community pool gets the remaining balance, which includes the truncated unit
				expectedCommunityDec := sdkmath.LegacyMustNewDecFromStr("87000000000000000000001")
				s.Require().Equal(expectedCommunityDec, communityPoolBalance)
				fmt.Printf("After one epoch expectedCommunityDec: communityPoolBalance %s:%s\n", expectedCommunityDec, communityPoolBalance)
			})
		})

		Context("with inflation param disabled", func() {
			BeforeEach(func() {
				params := s.app.InflationKeeper.GetParams(s.ctx)
				params.EnableInflation = false
				s.app.InflationKeeper.SetParams(s.ctx, params)
			})

			Context("after the network was offline for several days/epochs", func() {
				BeforeEach(func() {
					s.CommitAfter(time.Minute)        // start initial epoch
					s.CommitAfter(time.Hour * 24 * 5) // end epoch after several days
				})

				When("the epoch start time has not caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 3 blocks to trigger afterEpochEnd let EpochStartTime
						// catch up with BlockTime
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)

						epochInfo, found := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.app.InflationKeeper.GetSkippedEpochs(s.ctx)

						// Add one more commit to ensure epoch transition
						s.CommitAfter(time.Hour * 24)
					})

					It("should increase the epoch number ", func() {
						epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber + 1))
						fmt.Printf("epochInfo.CurrentEpoch %d\n", epochInfo.CurrentEpoch)
					})

					It("should increase the skipped epochs number", func() {
						skippedAfter := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
						Expect(skippedAfter).To(Equal(skipped + 1))
					})
				})

				When("the epoch start time has caught up with the block time", func() {
					BeforeEach(func() {
						// commit next 4 blocks to trigger afterEpochEnd hook several times
						// and let EpochStartTime catch up with BlockTime
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)
						s.CommitAfter(time.Second * 6)

						epochInfo, found := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						s.Require().True(found)
						epochNumber = epochInfo.CurrentEpoch

						skipped = s.app.InflationKeeper.GetSkippedEpochs(s.ctx)

						s.CommitAfter(time.Second * 6) // commit next block
					})

					It("should not increase the epoch number ", func() {
						epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
						Expect(epochInfo.CurrentEpoch).To(Equal(epochNumber))
					})

					It("should not increase the skipped epochs number", func() {
						skippedAfter := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
						Expect(skippedAfter).To(Equal(skipped))
					})

					When("epoch number passes epochsPerPeriod + skippedEpochs and inflation re-enabled", func() {
						BeforeEach(func() {
							params := s.app.InflationKeeper.GetParams(s.ctx)
							params.EnableInflation = true
							s.app.InflationKeeper.SetParams(s.ctx, params)

							epochInfo, _ := s.app.EpochsKeeper.GetEpochInfo(s.ctx, epochstypes.DayEpochID)
							epochNumber := epochInfo.CurrentEpoch // 6

							epochsPerPeriod := int64(1) //next year (year = 1)
							s.app.InflationKeeper.SetEpochsPerPeriod(s.ctx, epochsPerPeriod)
							skipped := s.app.InflationKeeper.GetSkippedEpochs(s.ctx)
							s.Require().Equal(epochNumber, epochsPerPeriod+int64(skipped))

							provision, found = s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							s.Require().True(found)

							s.CommitAfter(time.Hour * 23) // commit before next full epoch
							provisionAfter, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							s.Require().Equal(provisionAfter, provision)

							s.CommitAfter(time.Hour * 2) // commit after next full epoch
						})

						It("should recalculate the EpochMintProvision", func() {
							provisionAfter, _ := s.app.InflationKeeper.GetEpochMintProvision(s.ctx)
							Expect(provisionAfter).ToNot(Equal(provision))

							// Calculate expected provision
							params := s.app.InflationKeeper.GetParams(s.ctx)
							bondedRatio := s.app.InflationKeeper.BondedRatio(s.ctx)
							period := s.app.InflationKeeper.GetPeriod(s.ctx)
							epochsPerPeriod := s.app.InflationKeeper.GetEpochsPerPeriod(s.ctx)

							expectedProvision := types.CalculateEpochMintProvision(
								params,
								period,
								epochsPerPeriod,
								bondedRatio,
							)

							Expect(provisionAfter).To(Equal(expectedProvision))
							fmt.Printf("MintProvision after 1 years  %s", provisionAfter)
						})
					})
				})
			})
		})
	})
})

func (s *KeeperTestSuite) clearValidatorsAndInitPool(power int64) {
	amt := s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, power)
	notBondedPool := s.app.StakingKeeper.GetNotBondedPool(s.ctx)
	bondDenom, err := s.app.StakingKeeper.BondDenom(s.ctx)
	s.Require().NoError(err)
	totalSupply := sdk.NewCoins(sdk.NewCoin(bondDenom, amt))
	s.app.AccountKeeper.SetModuleAccount(s.ctx, notBondedPool)
	err = FundModuleAccount(s.app.BankKeeper, s.ctx, notBondedPool.GetName(), totalSupply)
	s.Require().NoError(err)
}

func FundModuleAccount(bk bankKeeper.Keeper, ctx sdk.Context, recipient string, amount sdk.Coins) error {
	if err := bk.MintCoins(ctx, types.ModuleName, amount); err != nil {
		panic(err)
	}
	return bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, recipient, amount)
}

func MakeValAccts(numAccts int) []sdk.ValAddress {
	addrs := make([]sdk.ValAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		addrs[i] = sdk.ValAddress(sdk.AccAddress(pk.Address()))
	}
	return addrs
}

func GenKeys(numKeys int) []*ed25519.PrivKey {
	pks := make([]*ed25519.PrivKey, numKeys)
	for i := 0; i < numKeys; i++ {
		pks[i] = ed25519.GenPrivKey()
	}
	return pks
}
