package keeper_test

import (
	sdkmath "cosmossdk.io/math"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	// Check initial epoch mint provision
	// 2.9M per year / 30 epochs = ~96,666.67 per epoch
	expMintProvision := sdkmath.LegacyMustNewDecFromStr("96666666666666666666667")

	mintProvision, found := suite.app.InflationKeeper.GetEpochMintProvision(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(expMintProvision, mintProvision)
}
