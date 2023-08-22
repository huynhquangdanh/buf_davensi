package countries

import (
	"context"

	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbUoMs "davensi.com/core/gen/uoms"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/uoms"
	"github.com/jackc/pgx/v5"
)

func (s *ServiceServer) GenCreationHandleFn() {

}

func (s *ServiceServer) GenUpdateHandleFn() {

}

func (s *ServiceServer) GenUpsertUomHandleFn(country *pbCountries.Country, selectUoms *pbUoMs.SelectList) (
	handleFn func(tx pgx.Tx) ([]*CountryUom, error),
	err *common.ErrWithCode,
) {
	errGenFn := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
		"upsert",
		_entityName,
		"",
	)

	if errSelectUomsUpsert := validateUpsertCountriesUoms(country, selectUoms); errSelectUomsUpsert != nil {
		return nil, errSelectUomsUpsert
	}

	qb, errGenUpsertFiat := QbUpsertUoms(country, selectUoms)

	if errGenUpsertFiat != nil {
		return nil, errGenFn.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT).
			UpdateMessage(errGenUpsertFiat.Error())
	}

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")

	return func(tx pgx.Tx) ([]*CountryUom, error) {
		return common.TxBulkWrite(
			context.Background(),
			tx,
			sqlStr,
			args,
			ScanCountriesUoms,
		)
	}, nil
}

func (s *ServiceServer) GenSetFiatsHandleFn(country *pbCountries.Country, selectFiats *pbUoMs.SelectList) (
	handleFn func(tx pgx.Tx) ([]*CountryUom, error),
	err *common.ErrWithCode,
) {
	if errValidate := uoms.ValidateSelectList(selectFiats, "set fiats"); errValidate != nil {
		return nil, errValidate
	}

	if errValidate := s.validateUpsertCountriesRelationship(nil, selectFiats); errValidate != nil {
		return nil, errValidate
	}

	return s.GenUpsertUomHandleFn(country, selectFiats)
}

func (s *ServiceServer) GenAddFiatsHandleFn(country *pbCountries.Country, selectFiats *pbUoMs.SelectList) (
	handleFn func(tx pgx.Tx) ([]*CountryUom, error),
	err *common.ErrWithCode,
) {
	if errValidate := uoms.ValidateSelectList(selectFiats, "add fiats"); errValidate != nil {
		return nil, errValidate
	}

	if errValidate := s.validateUpsertCountriesRelationship(nil, selectFiats); errValidate != nil {
		return nil, errValidate
	}

	return s.GenUpsertUomHandleFn(country, selectFiats)
}

func (s *ServiceServer) GenSetCryptosHandleFn(country *pbCountries.Country, selectCryptos *pbUoMs.SelectList) (
	handleFn func(tx pgx.Tx) ([]*CountryUom, error),
	err *common.ErrWithCode,
) {
	if errValidate := uoms.ValidateSelectList(selectCryptos, "set cryptos"); errValidate != nil {
		return nil, errValidate
	}

	if errValidate := s.validateUpsertCountriesRelationship(selectCryptos, nil); errValidate != nil {
		return nil, errValidate
	}

	return s.GenUpsertUomHandleFn(country, selectCryptos)
}

func (s *ServiceServer) GenAddCryptosHandleFn(country *pbCountries.Country, selectCryptos *pbUoMs.SelectList) (
	handleFn func(tx pgx.Tx) ([]*CountryUom, error),
	err *common.ErrWithCode,
) {
	if errValidate := uoms.ValidateSelectList(selectCryptos, "add cryptos"); errValidate != nil {
		return nil, errValidate
	}

	if errValidate := s.validateUpsertCountriesRelationship(selectCryptos, nil); errValidate != nil {
		return nil, errValidate
	}

	return s.GenUpsertUomHandleFn(country, selectCryptos)
}

func (s *ServiceServer) GenSetMarketsHandleFn() {

}

func (s *ServiceServer) GenAddMarketsHandleFn() {

}
