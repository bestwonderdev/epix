package keeper_test

import (
	sdkmath "cosmossdk.io/math"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	// Check initial epoch mint provision
	// 2.9M per year / 365 epochs = ~7945.21 per epoch
	expMintProvision := sdkmath.LegacyMustNewDecFromStr("7945205479452054794521")

	mintProvision, found := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(expMintProvision, mintProvision)
}
