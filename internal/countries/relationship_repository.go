package countries

import (
	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbCryptos "davensi.com/core/gen/cryptos"
	pbFiats "davensi.com/core/gen/fiats"
	pbUoMs "davensi.com/core/gen/uoms"
	"github.com/jackc/pgx/v5"

	"davensi.com/core/internal/cryptos"
	"davensi.com/core/internal/fiats"
	"davensi.com/core/internal/uoms"
	"davensi.com/core/internal/util"
)

type CountryUom struct {
	CountryID string          `db:"country_id"`
	UomID     string          `db:"uom_id"`
	Status    pbCommon.Status `db:"status"`
}

const (
	_countriesUomsTableName = "core.countries_uoms"
)

var uomsRepo = uoms.NewUoMRepository(nil)
var fiatsRepo = fiats.NewFiatRepository(nil)
var cryptosRepo = cryptos.NewCryptoRepository(nil)

func QbUpsertUoms(country *pbCountries.Country, selectUoms *pbUoMs.SelectList) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Upsert, _countriesUomsTableName)

	qb.SetInsertField("country_id", "uom_id", "status")

	for _, selectUom := range selectUoms.GetList() {
		_, err := qb.SetInsertValues(
			[]any{
				country.GetId(),
				selectUom.GetById(),
				pbCommon.Status_STATUS_ACTIVE,
			},
		)

		if err != nil {
			return nil, err
		}
	}

	return qb, nil
}

func ScanCountriesUoms(rows pgx.Rows) ([]*CountryUom, error) {
	countriesUoms := []*CountryUom{}

	for rows.Next() {
		countryUom := CountryUom{}
		err := rows.Scan(&countryUom.CountryID, &countryUom.UomID, &countryUom.Status)

		if err != nil {
			return nil, err
		}

		countriesUoms = append(countriesUoms, &countryUom)
	}

	return countriesUoms, nil
}

func genUomFilter(selectUoms *pbUoMs.SelectList) *util.FilterBracket {
	uomFilter := util.CreateFilterBracket("or")

	for _, selectUoM := range selectUoms.GetList() {
		switch selectUoM.GetSelect().(type) {
		case *pbUoMs.Select_ById:
			uomFilter.SetFilter("uoms.id = ?", selectUoM.GetById())
		case *pbUoMs.Select_ByTypeSymbol:
			uomFilter.SetFilter(
				"(uoms.type = ? AND uoms.symbol = ?)",
				selectUoM.GetByTypeSymbol().GetType(),
				selectUoM.GetByTypeSymbol().GetSymbol(),
			)
		}
	}
	return uomFilter
}

func QbGetListFiats(selectFiats *pbUoMs.SelectList) *util.QueryBuilder {
	qb := fiatsRepo.QbGetList(&pbFiats.GetListRequest{}, uomsRepo.QbGetList(&pbUoMs.GetListRequest{}))

	uomFilter := genUomFilter(selectFiats)

	uomQuery, uomArgs := uomFilter.GenerateSQL()

	return qb.Where(uomQuery, uomArgs...).Where("uoms.status = ?", pbCommon.Status_STATUS_ACTIVE)
}

func QbGetListCryptos(selectFiats *pbUoMs.SelectList) *util.QueryBuilder {
	qb := cryptosRepo.QbGetList(&pbCryptos.GetListRequest{}, uomsRepo.QbGetList(&pbUoMs.GetListRequest{}))

	uomFilter := genUomFilter(selectFiats)

	uomQuery, uomArgs := uomFilter.GenerateSQL()

	return qb.Where(uomQuery, uomArgs...).Where("uoms.status = ?", pbCommon.Status_STATUS_ACTIVE)
}
