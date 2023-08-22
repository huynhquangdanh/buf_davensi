package ohlcvt

import (
	"context"

	pbDataSources "davensi.com/core/gen/datasources"
	pbMarkets "davensi.com/core/gen/markets"

	"github.com/bufbuild/connect-go"
)

type OHLCVTRelationships struct {
	dataSource *pbDataSources.DataSource
	market     *pbMarkets.Market
}

func (s *ServiceServer) GetRelationship(
	selectSource *pbDataSources.Select,
	selectMarket *pbMarkets.Select,
) OHLCVTRelationships {
	datasourceRl := OHLCVTRelationships{}
	if selectSource != nil {
		getSourceResponse, err := s.srcService.Get(context.Background(), &connect.Request[pbDataSources.GetRequest]{
			Msg: &pbDataSources.GetRequest{
				Select: selectSource,
			},
		})
		if err == nil {
			datasourceRl.dataSource = getSourceResponse.Msg.GetDatasource()
		}
	}
	if selectMarket != nil {
		getMarketResponse, err := s.marketService.Get(context.Background(), &connect.Request[pbMarkets.GetRequest]{
			Msg: &pbMarkets.GetRequest{
				Select: selectMarket,
			},
		})
		if err == nil {
			datasourceRl.market = getMarketResponse.Msg.GetMarket()
		}
	}

	return datasourceRl
}
