package defiwallets

import (
	"context"

	pbBlockchains "davensi.com/core/gen/blockchains"

	"github.com/bufbuild/connect-go"
)

type DefiwalletRelationships struct {
	blockchain *pbBlockchains.Blockchain
}

func (s *ServiceServer) GetRelationship(
	selectBlockchain *pbBlockchains.Select,
) DefiwalletRelationships {
	datasourceRl := DefiwalletRelationships{}
	if selectBlockchain != nil {
		getProviderResponse, err := s.blockchainSS.Get(context.Background(), &connect.Request[pbBlockchains.GetRequest]{
			Msg: &pbBlockchains.GetRequest{
				Select: selectBlockchain,
			},
		})
		if err == nil {
			datasourceRl.blockchain = getProviderResponse.Msg.GetBlockchain()
		}
	}

	return datasourceRl
}
