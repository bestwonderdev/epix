package types

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var DefaultInflationDenom = "aepix"

// Parameter store keys
var (
	ParamStoreKeyMintDenom              = []byte("ParamStoreKeyMintDenom")
	ParamStoreKeyExponentialCalculation = []byte("ParamStoreKeyExponentialCalculation")
	ParamStoreKeyInflationDistribution  = []byte("ParamStoreKeyInflationDistribution")
	ParamStoreKeyEnableInflation        = []byte("ParamStoreKeyEnableInflation")
)

// ParamTable for inflation module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	exponentialCalculation ExponentialCalculation,
	inflationDistribution InflationDistribution,
	enableInflation bool,
) Params {
	return Params{
		MintDenom:              mintDenom,
		ExponentialCalculation: exponentialCalculation,
		InflationDistribution:  inflationDistribution,
		EnableInflation:        enableInflation,
	}
}

// default minting module parameter
func DefaultParams() Params {
	return Params{
		MintDenom: "aepix",
		ExponentialCalculation: ExponentialCalculation{
			A:             sdkmath.LegacyNewDec(2900000),        // 2.9M first year
			R:             sdkmath.LegacyNewDecWithPrec(159, 3), // ~15.9% for 4-year halving
			C:             sdkmath.LegacyNewDec(42000000),       // 42M max supply
			BondingTarget: sdkmath.LegacyNewDecWithPrec(80, 2),  // not relevant; max variance is 0
			MaxVariance:   sdkmath.LegacyNewDecWithPrec(0, 2),   // 0% variance
		},
		InflationDistribution: InflationDistribution{
			StakingRewards: sdkmath.LegacyNewDecWithPrec(10, 2), // 10% minimum to stakers
			CommunityPool:  sdkmath.LegacyNewDecWithPrec(90, 2), // 90% maximum to contributors
		},
		EnableInflation: true,
	}
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyExponentialCalculation, &p.ExponentialCalculation, validateExponentialCalculation),
		paramtypes.NewParamSetPair(ParamStoreKeyInflationDistribution, &p.InflationDistribution, validateInflationDistribution),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableInflation, &p.EnableInflation, validateBool),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateExponentialCalculation(i interface{}) error {
	v, ok := i.(ExponentialCalculation)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// validate initial value
	if v.A.IsNegative() {
		return fmt.Errorf("initial value cannot be negative")
	}

	// validate reduction factor
	if v.R.GT(sdkmath.LegacyNewDec(1)) {
		return fmt.Errorf("reduction factor cannot be greater than 1")
	}

	if v.R.IsNegative() {
		return fmt.Errorf("reduction factor cannot be negative")
	}

	// validate long term inflation
	if v.C.IsNegative() {
		return fmt.Errorf("long term inflation cannot be negative")
	}

	// validate bonded target
	if v.BondingTarget.GT(sdkmath.LegacyNewDec(1)) {
		return fmt.Errorf("bonded target cannot be greater than 1")
	}

	if !v.BondingTarget.IsPositive() {
		return fmt.Errorf("bonded target cannot be zero or negative")
	}

	// validate max variance
	if v.MaxVariance.IsNegative() {
		return fmt.Errorf("max variance cannot be negative")
	}

	return nil
}

func validateInflationDistribution(i interface{}) error {
	v, ok := i.(InflationDistribution)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.StakingRewards.IsNegative() {
		return errors.New("staking distribution ratio must not be negative")
	}

	if v.CommunityPool.IsNegative() {
		return errors.New("community pool distribution ratio must not be negative")
	}

	totalProportions := v.StakingRewards.Add(v.CommunityPool)
	if !totalProportions.Equal(sdkmath.LegacyNewDec(1)) {
		return errors.New("total distributions ratio should be 1")
	}

	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateExponentialCalculation(p.ExponentialCalculation); err != nil {
		return err
	}
	if err := validateInflationDistribution(p.InflationDistribution); err != nil {
		return err
	}

	return validateBool(p.EnableInflation)
}
