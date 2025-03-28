package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

func (suite *KeeperTestSuite) TestSetGetEpochMintProvision() {
	expEpochMintProvision := sdkmath.LegacyNewDec(1_000_000)

	testCases := []struct {
		name     string
		malleate func()
		genesis  bool
	}{
		{
			"default EpochMintProvision",
			func() {},
			true,
		},
		{
			"period EpochMintProvision",
			func() {
				suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, expEpochMintProvision)
			},
			false,
		},
	}

	// 2.9M per year / 365 epochs = 7945.21 per epoch
	genesisProvision := sdkmath.LegacyMustNewDecFromStr("7945205479452054794521")

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			provision, found := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
			suite.Require().True(found)

			if tc.genesis {
				suite.Require().Equal(genesisProvision, provision, tc.name)
			} else {
				suite.Require().Equal(expEpochMintProvision, provision, tc.name)
			}
		})
	}
}
