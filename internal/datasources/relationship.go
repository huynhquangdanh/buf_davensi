package datasources

import (
	"context"

	pbFsproviders "davensi.com/core/gen/fsproviders"
	"github.com/bufbuild/connect-go"
)

type DatasourceRelationships struct {
	fsprovider *pbFsproviders.FSProvider
}

func (s *ServiceServer) GetRelationship(
	selectProvider *pbFsproviders.Select,
) DatasourceRelationships {
	datasourceRl := DatasourceRelationships{}
	if selectProvider != nil {
		getProviderResponse, err := s.providerSS.Get(context.Background(), &connect.Request[pbFsproviders.GetRequest]{
			Msg: &pbFsproviders.GetRequest{
				Select: selectProvider,
			},
		})
		if err == nil {
			datasourceRl.fsprovider = getProviderResponse.Msg.GetFsprovider()
		}
	}

	return datasourceRl
}
