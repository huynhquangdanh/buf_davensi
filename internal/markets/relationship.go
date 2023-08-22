package markets

import (
	"context"

	pbTradingpairs "davensi.com/core/gen/tradingpairs"
	"github.com/bufbuild/connect-go"
)

type MarketRelationships struct {
	tradingpair *pbTradingpairs.TradingPair
}

func (s *ServiceServer) GetRelationship(
	selectProvider *pbTradingpairs.Select,
) MarketRelationships {
	marketRl := MarketRelationships{}
	if selectProvider != nil {
		getProviderResponse, err := s.tradingpairSS.Get(context.Background(), &connect.Request[pbTradingpairs.GetRequest]{
			Msg: &pbTradingpairs.GetRequest{
				Select: selectProvider,
			},
		})
		if err == nil {
			marketRl.tradingpair = getProviderResponse.Msg.GetTradingpair()
		}
	}

	return marketRl
}
