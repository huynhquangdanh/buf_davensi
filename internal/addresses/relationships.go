package addresses

import (
	"context"

	pbCountries "davensi.com/core/gen/countries"
	"github.com/bufbuild/connect-go"
)

type AddressRelationships struct {
	Country *pbCountries.Country
}

// For Country field
func (s *ServiceServer) GetRelationship(
	selectCountry *pbCountries.Select,
) AddressRelationships {
	countryChan := make(chan *pbCountries.Country)

	// Country field
	go func() {
		var existCountry *pbCountries.Country
		if selectCountry != nil {
			getCountryResponse, err := s.countryService.Get(context.Background(), &connect.Request[pbCountries.GetRequest]{
				Msg: &pbCountries.GetRequest{
					Select: selectCountry,
				},
			})
			if err == nil {
				existCountry = getCountryResponse.Msg.GetCountry()
			}
		}
		countryChan <- existCountry
	}()

	return AddressRelationships{
		Country: <-countryChan,
	}
}
