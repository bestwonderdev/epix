package types

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
)

type InflationTestSuite struct {
	suite.Suite
}

func TestInflationSuite(t *testing.T) {
	suite.Run(t, new(InflationTestSuite))
}

func (suite *InflationTestSuite) TestCalculateEpochMintProvision() {
	bondingParams := DefaultParams()
	bondingParams.ExponentialCalculation.MaxVariance = sdkmath.LegacyNewDecWithPrec(0, 2)
	epochsPerPeriod := int64(30)

	testCases := []struct {
		name              string
		params            Params
		period            uint64
		bondedRatio       sdkmath.LegacyDec
		expEpochProvision sdkmath.LegacyDec
		expPass           bool
	}{
		{
			"pass - initial period",
			DefaultParams(),
			uint64(0),
			sdkmath.LegacyOneDec(),
			sdkmath.LegacyMustNewDecFromStr("96666666666666666666667"), // 2.9M/30 epochs = ~96,666.67 per epoch
			true,
		},
		{
			"pass - period 1 (after 1 year)",
			DefaultParams(),
			uint64(1),
			sdkmath.LegacyOneDec(),
			sdkmath.LegacyMustNewDecFromStr("81296666666666666666667"), // ~2.44M/30 epochs after 15.9% reduction
			true,
		},
		{
			"pass - period 2 (after 2 years)",
			DefaultParams(),
			uint64(2),
			sdkmath.LegacyOneDec(),
			sdkmath.LegacyMustNewDecFromStr("68370496666666666666667"), // ~2.05M/30 epochs after (1-0.159)^2 reduction
			true,
		},
		{
			"pass - period 4 (after 4 years)",
			DefaultParams(),
			uint64(4),
			sdkmath.LegacyOneDec(),
			sdkmath.LegacyMustNewDecFromStr("48357153252896666666667"), // ~1.45M/30 epochs after (1-0.159)^4 reduction
			true,
		},
		{
			"pass - period 8 (after 8 years)",
			DefaultParams(),
			uint64(8),
			sdkmath.LegacyOneDec(),
			sdkmath.LegacyMustNewDecFromStr("24190492455766910403333"), // ~0.725M/30 epochs after (1-0.159)^8 reduction
			true,
		},
		{
			"pass - period 12 (after 12 years)",
			DefaultParams(),
			uint64(12),
			sdkmath.LegacyOneDec(),
			sdkmath.LegacyMustNewDecFromStr("12101207078757528843333"), // ~0.363M/30 epochs after (1-0.159)^12 reduction
			true,
		},
		{
			"pass - 0 percent bonding - initial period",
			bondingParams,
			uint64(0),
			sdkmath.LegacyZeroDec(),
			sdkmath.LegacyMustNewDecFromStr("96666666666666666666667"),
			true,
		},
		{
			"pass - 0 percent bonding - period 1",
			bondingParams,
			uint64(1),
			sdkmath.LegacyZeroDec(),
			sdkmath.LegacyMustNewDecFromStr("81296666666666666666667"),
			true,
		},
		{
			"pass - 0 percent bonding - period 2",
			bondingParams,
			uint64(2),
			sdkmath.LegacyZeroDec(),
			sdkmath.LegacyMustNewDecFromStr("68370496666666666666667"),
			true,
		},
		{
			"pass - 0 percent bonding - period 4",
			bondingParams,
			uint64(4),
			sdkmath.LegacyZeroDec(),
			sdkmath.LegacyMustNewDecFromStr("48357153252896666666667"),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			epochMintProvisions := CalculateEpochMintProvision(
				tc.params,
				tc.period,
				epochsPerPeriod,
				tc.bondedRatio,
			)

			suite.Require().Equal(tc.expEpochProvision, epochMintProvisions)
		})
	}
}
