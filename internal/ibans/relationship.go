package ibans

import (
	"context"

	pbCountries "davensi.com/core/gen/countries"
	"github.com/bufbuild/connect-go"
)

type IncomeRelationships struct {
	country *pbCountries.Country
}

func (s *ServiceServer) GetRelationship(
	selectCountry *pbCountries.Select,
) IncomeRelationships {
	incomceRL := IncomeRelationships{}
	if selectCountry != nil {
		getCountryResponse, err := s.countrySS.Get(context.Background(), &connect.Request[pbCountries.GetRequest]{
			Msg: &pbCountries.GetRequest{
				Select: selectCountry,
			},
		})
		if err == nil {
			incomceRL.country = getCountryResponse.Msg.GetCountry()
		}
	}

	return incomceRL
}
