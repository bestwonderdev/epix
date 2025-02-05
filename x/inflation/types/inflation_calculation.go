package types

import (
	sdkmath "cosmossdk.io/math"

	ethermint "github.com/evmos/ethermint/types"
)

// CalculateEpochProvisions returns mint provision per epoch
func CalculateEpochMintProvision(
	params Params,
	period uint64,
	epochsPerPeriod int64,
	bondedRatio sdkmath.LegacyDec,
) sdkmath.LegacyDec {
	x := period                                              // period
	a := params.ExponentialCalculation.A                     // initial value
	r := params.ExponentialCalculation.R                     // reduction factor
	c := params.ExponentialCalculation.C                     // long term inflation target
	bTarget := params.ExponentialCalculation.BondingTarget   // bonding target
	maxVariance := params.ExponentialCalculation.MaxVariance // max percentage that inflation can be increased by

	// exponentialDecay := a * (1 - r) ^ x
	decay := sdkmath.LegacyOneDec().Sub(r)
	exponentialDecay := a.Mul(decay.Power(x))

	// bondingIncentive doesn't increase beyond bonding target (0 < b < bonding_target)
	if bondedRatio.GTE(bTarget) {
		bondedRatio = bTarget
	}

	// bondingIncentive = 1 + max_variance - bondingRatio * (max_variance / bonding_target)
	sub := bondedRatio.Mul(maxVariance.Quo(bTarget))
	bondingIncentive := sdkmath.LegacyOneDec().Add(maxVariance).Sub(sub)

	// periodProvision = exponentialDecay * bondingIncentive
	periodProvision := exponentialDecay.Mul(bondingIncentive)

	// epochProvision = periodProvision / epochsPerPeriod
	epochProvision := periodProvision.Quo(sdkmath.LegacyNewDec(epochsPerPeriod))

	// Multiply epochMintProvision with power reduction (10^18 for epix) as the
	// calculation is based on `epix` and the issued tokens need to be given in
	// `aepix`
	epochProvision = epochProvision.Mul(ethermint.PowerReduction.ToLegacyDec())

	// Cap the epoch provision at the long term inflation target divided by epochs per period
	maxEpochProvision := c.Mul(ethermint.PowerReduction.ToLegacyDec()).Quo(sdkmath.LegacyNewDec(epochsPerPeriod))
	if epochProvision.GT(maxEpochProvision) {
		return maxEpochProvision
	}

	return epochProvision
}
