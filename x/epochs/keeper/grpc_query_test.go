package keeper_test

import (
	"fmt"
	"time"

	"github.com/EpixZone/epix/x/epochs/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func (suite *KeeperTestSuite) TestEpochInfo() {
	var (
		req    *types.QueryEpochsInfoRequest
		expRes *types.QueryEpochsInfoResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default EpochInfos",
			func() {
				req = &types.QueryEpochsInfoRequest{}

				day := types.EpochInfo{
					Identifier:              types.DayEpochID,
					StartTime:               time.Time{},
					Duration:                time.Hour * 24,
					CurrentEpoch:            0,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    false,
				}
				day.StartTime = suite.ctx.BlockTime()
				day.CurrentEpochStartHeight = suite.ctx.BlockHeight()

				week := types.EpochInfo{
					Identifier:              types.WeekEpochID,
					StartTime:               time.Time{},
					Duration:                time.Hour * 24 * 7,
					CurrentEpoch:            0,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    false,
				}
				week.StartTime = suite.ctx.BlockTime()
				week.CurrentEpochStartHeight = suite.ctx.BlockHeight()

				expRes = &types.QueryEpochsInfoResponse{
					Epochs: []types.EpochInfo{day, week},
					Pagination: &query.PageResponse{
						NextKey: nil,
						Total:   uint64(2),
					},
				}
			},
			true,
		},
		{
			"set epoch info",
			func() {
				day := types.EpochInfo{
					Identifier:              types.DayEpochID,
					StartTime:               time.Time{},
					Duration:                time.Hour * 24,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    true,
				}
				day.StartTime = suite.ctx.BlockTime()
				day.CurrentEpochStartTime = suite.ctx.BlockTime()
				day.CurrentEpochStartHeight = suite.ctx.BlockHeight()

				week := types.EpochInfo{
					Identifier:              types.WeekEpochID,
					StartTime:               time.Time{},
					Duration:                time.Hour * 24 * 7,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    true,
				}
				week.StartTime = suite.ctx.BlockTime()
				week.CurrentEpochStartTime = suite.ctx.BlockTime()
				week.CurrentEpochStartHeight = suite.ctx.BlockHeight()

				quarter := types.EpochInfo{
					Identifier:              "quarter",
					StartTime:               time.Time{},
					Duration:                time.Hour * 24 * 7 * 13,
					CurrentEpoch:            1,
					CurrentEpochStartHeight: 1,
					CurrentEpochStartTime:   time.Time{},
					EpochCountingStarted:    true,
				}
				quarter.StartTime = suite.ctx.BlockTime()
				quarter.CurrentEpochStartTime = suite.ctx.BlockTime()
				quarter.CurrentEpochStartHeight = suite.ctx.BlockHeight()
				suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, quarter)
				suite.Commit()

				req = &types.QueryEpochsInfoRequest{}
				expRes = &types.QueryEpochsInfoResponse{
					Epochs: []types.EpochInfo{day, quarter, week},
					Pagination: &query.PageResponse{
						NextKey: nil,
						Total:   uint64(3),
					},
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := suite.ctx
			tc.malleate()

			res, err := suite.queryClient.EpochInfos(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCurrentEpoch() {
	var (
		req    *types.QueryCurrentEpochRequest
		expRes *types.QueryCurrentEpochResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"unknown identifier",
			func() {
				defaultCurrentEpoch := int64(0)
				req = &types.QueryCurrentEpochRequest{Identifier: "second"}
				expRes = &types.QueryCurrentEpochResponse{
					CurrentEpoch: defaultCurrentEpoch,
				}
			},
			false,
		},
		{
			"week - default currentEpoch",
			func() {
				defaultCurrentEpoch := int64(0)
				req = &types.QueryCurrentEpochRequest{Identifier: types.WeekEpochID}
				expRes = &types.QueryCurrentEpochResponse{
					CurrentEpoch: defaultCurrentEpoch,
				}
			},
			true,
		},
		{
			"day - default currentEpoch",
			func() {
				defaultCurrentEpoch := int64(0)
				req = &types.QueryCurrentEpochRequest{Identifier: types.DayEpochID}
				expRes = &types.QueryCurrentEpochResponse{
					CurrentEpoch: defaultCurrentEpoch,
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			res, err := suite.queryClient.CurrentEpoch(suite.ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
