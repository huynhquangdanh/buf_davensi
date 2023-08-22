package countries

import (
	pbCryptos "davensi.com/core/gen/cryptos"
	pbFiats "davensi.com/core/gen/fiats"
	pbUoMs "davensi.com/core/gen/uoms"
)

type CountryRelationship struct {
	Fiats   []*pbFiats.Fiat
	Cryptos []*pbCryptos.Crypto
}

func (s *ServiceServer) GetRelationship(
	selectCryptos *pbUoMs.SelectList,
	selectFiats *pbUoMs.SelectList,
) *CountryRelationship {
	relationship := &CountryRelationship{
		Fiats:   []*pbFiats.Fiat{},
		Cryptos: []*pbCryptos.Crypto{},
	}

	if selectCryptos != nil {
		cryptos := make(chan []*pbCryptos.Crypto)
		go func() {
			cryptos <- []*pbCryptos.Crypto{}
		}()
		relationship.Cryptos = <-cryptos
	}

	if selectFiats != nil {
		fiats := make(chan []*pbFiats.Fiat)
		go func() {
			fiats <- []*pbFiats.Fiat{}
		}()
		relationship.Fiats = <-fiats
	}

	return relationship
}
