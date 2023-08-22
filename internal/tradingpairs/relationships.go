package tradingpairs

import (
	"context"

	pbUoms "davensi.com/core/gen/uoms"
	"github.com/bufbuild/connect-go"
)

type TradingPairRelationships struct {
	qUom *pbUoms.UoM
	pUom *pbUoms.UoM
}

func (s *ServiceServer) GetRelationship(
	selectQuantityUom *pbUoms.Select, selectPriceUom *pbUoms.Select,
) TradingPairRelationships {
	relationship := TradingPairRelationships{}

	if selectQuantityUom != nil {
		getQuantityUomResponse, err := s.uomSS.Get(context.Background(), &connect.Request[pbUoms.GetRequest]{
			Msg: &pbUoms.GetRequest{
				Select: selectQuantityUom,
			},
		})
		if err == nil {
			relationship.qUom = getQuantityUomResponse.Msg.GetUom()
		}
	}

	if selectPriceUom != nil {
		getPriceUomResponse, err := s.uomSS.Get(context.Background(), &connect.Request[pbUoms.GetRequest]{
			Msg: &pbUoms.GetRequest{
				Select: selectPriceUom,
			},
		})
		if err == nil {
			relationship.pUom = getPriceUomResponse.Msg.GetUom()
		}
	}

	return relationship
}
