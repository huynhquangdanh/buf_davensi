package prices

import (
	"context"

	pbDataSources "davensi.com/core/gen/datasources"
	pbMarkets "davensi.com/core/gen/markets"
	"github.com/bufbuild/connect-go"
)

type PriceRelationships struct {
	Market     *pbMarkets.Market
	DataSource *pbDataSources.DataSource
}

func (s *ServiceServer) GetRelationship(
	selectMarket *pbMarkets.Select,
	selectDataSource *pbDataSources.Select,
) PriceRelationships {
	marketChan := make(chan *pbMarkets.Market)
	dataSourceChan := make(chan *pbDataSources.DataSource)

	go func() {
		var existMarket *pbMarkets.Market
		if selectMarket != nil {
			getMarketResponse, err := s.marketsSS.Get(context.Background(), &connect.Request[pbMarkets.GetRequest]{
				Msg: &pbMarkets.GetRequest{
					Select: selectMarket,
				},
			})
			if err == nil {
				existMarket = getMarketResponse.Msg.GetMarket()
			}
		}
		marketChan <- existMarket
	}()
	go func() {
		var existDataSource *pbDataSources.DataSource
		if selectMarket != nil {
			getMarketResponse, err := s.dataSourcesSS.GetMainEntity(context.Background(), &connect.Request[pbDataSources.GetRequest]{
				Msg: &pbDataSources.GetRequest{
					Select: selectDataSource,
				},
			})
			if err == nil {
				existDataSource = getMarketResponse.Msg.GetDatasource()
			}
		}
		dataSourceChan <- existDataSource
	}()

	return PriceRelationships{
		Market:     <-marketChan,
		DataSource: <-dataSourceChan,
	}
}
