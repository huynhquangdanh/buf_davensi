package blockchains

import (
	"context"

	pbCryptos "davensi.com/core/gen/cryptos"
	pbUoMs "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/cryptos"
)

func (s *ServiceServer) GetCryptosSelectList(selectUoMs *pbUoMs.SelectList) []*pbCryptos.Crypto {
	result := []*pbCryptos.Crypto{}
	sqlstr, args, _ := uomsRepo.
		QbGetBySelect(selectUoMs).
		Where("uoms.type = ?", pbUoMs.Type_TYPE_CRYPTO).
		Join("LEFT JOIN core.cryptos ON uoms.id = cryptos.id").
		GenerateSQL()

	rows, err := s.db.Query(context.Background(), sqlstr, args)

	if err != nil {
		return result
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		crypto, err := cryptos.ScanCryptoUom(rows)
		if err != nil {
			continue
		}

		result = append(result, crypto)
	}

	return result
}
