package ibans

import (
	"database/sql"
	"errors"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	"davensi.com/core/gen/countries"
	pbIbans "davensi.com/core/gen/ibans"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_fields             = "ibans.id, country_id, valid_from, algorithm, format, weights, modulo, complement, method, ibans.status"
	_tableName          = "core.ibans"
	_tableNameCountries = "core.countries"
)

type IbanRepository struct {
	db *pgxpool.Pool
}

func NewIbanRepository(db *pgxpool.Pool) *IbanRepository {
	return &IbanRepository{
		db: db,
	}
}

func (s *IbanRepository) QbInsert(msg *pbIbans.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleValue := []any{}

	qb.SetInsertField("country_id").SetInsertField("algorithm").SetInsertField("format")
	singleValue = append(
		singleValue,
		msg.GetCountry().GetById(),
		msg.GetAlgorithm(),
		msg.GetFormat(),
	)

	// Append optional fields values
	if msg.Validity != nil {
		qb.SetInsertField("valid_from")
		singleValue = append(singleValue, util.GetDBTimestampValue(msg.GetValidity()))
	}
	if msg.Weights != nil {
		qb.SetInsertField("weights")
		singleValue = append(singleValue, msg.GetWeights())
	}
	if msg.Modulo != nil {
		qb.SetInsertField("modulo")
		singleValue = append(singleValue, msg.GetModulo())
	}
	if msg.Complement != nil {
		qb.SetInsertField("complement")
		singleValue = append(singleValue, msg.GetComplement())
	}
	if msg.Method != nil {
		qb.SetInsertField("method")
		singleValue = append(singleValue, msg.GetMethod())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleValue = append(singleValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleValue)

	qb.SetReturnFields("*")

	return qb, err
}

func (s *IbanRepository) QbUpdate(msg *pbIbans.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Country != nil {
		qb.SetUpdate("country_id", msg.GetCountry().GetById())
	}
	if msg.Algorithm != nil {
		qb.SetUpdate("algorithm", msg.GetAlgorithm())
	}
	if msg.Format != nil {
		qb.SetUpdate("format", msg.GetFormat())
	}
	if msg.Validity != nil {
		qb.SetUpdate("valid_from", util.GetDBTimestampValue(msg.GetValidity()))
	}
	if msg.Weights != nil {
		qb.SetUpdate("weights", msg.GetWeights())
	}
	if msg.Modulo != nil {
		qb.SetUpdate("modulo", msg.GetModulo())
	}
	if msg.Complement != nil {
		qb.SetUpdate("complement", msg.GetComplement())
	}
	if msg.Method != nil {
		qb.SetUpdate("method", msg.GetMethod())
	}
	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch msg.GetSelect().(type) {
	case *pbIbans.UpdateRequest_ById:
		qb.Where("id = ?", msg.GetById())
	case *pbIbans.UpdateRequest_ByCountryValidity:
		// Create country validity filter bracket
		countryValidityFB := createCountryValidityFB(msg.GetByCountryValidity())
		countryValidityBracket, countryValidityArgs := countryValidityFB.GenerateSQL()

		// Append country validity filter bracket to main query builder
		qb.Where(
			countryValidityBracket,
			countryValidityArgs...,
		)
	}

	qb.SetReturnFields("*")

	return qb, nil
}

func (s *IbanRepository) QbGetOne(msg *pbIbans.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	switch msg.GetSelect().Select.(type) {
	case *pbIbans.Select_ById:
		qb.Where("ibans.id = ?", msg.GetSelect().GetById())
	case *pbIbans.Select_ByCountryValidity:
		// Create country validity filter bracket
		countryValidityFB := createCountryValidityFB(msg.GetSelect().GetByCountryValidity())
		countryValidityBracket, countryValidityArgs := countryValidityFB.GenerateSQL()

		// Append country validity filter bracket to main query builder
		qb.Where(
			countryValidityBracket,
			countryValidityArgs...,
		)
	}

	qb.Where("ibans.status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

/**
 * @Todo: add filter by countries
 */
func (s *IbanRepository) QbGetList(msg *pbIbans.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	if msg.Algorithm != nil {
		qb.Where("algorithm = ?", msg.GetAlgorithm())
	}
	if msg.Format != nil {
		qb.Where("format LIKE '%' || ? || '%'", msg.GetFormat())
	}
	if msg.Weights != nil {
		qb.Where("weights LIKE '%' || ? || '%'", msg.GetWeights())
	}
	if msg.Modulo != nil {
		qb.Where("modulo LIKE '%' || ? || '%'", msg.GetModulo())
	}
	if msg.Complement != nil {
		qb.Where("complement LIKE '%' || ? || '%'", msg.GetComplement())
	}
	if msg.Method != nil {
		qb.Where("method LIKE '%' || ? || '%'", msg.GetMethod())
	}
	if msg.Status != nil {
		qb.Where("status = ?", msg.GetStatus())
	}

	return qb
}

func (s *IbanRepository) ScanMainEntity(row pgx.Row) (*pbIbans.IBAN, error) {
	var (
		id         string
		countryID  string
		validFrom  sql.NullTime
		algorithm  pbIbans.Algorithm
		format     sql.NullString
		weights    sql.NullString
		modulo     sql.NullString
		complement sql.NullString
		method     sql.NullString
		status     pbCommon.Status
	)

	err := row.Scan(
		&id,
		&countryID,
		&validFrom,
		&algorithm,
		&format,
		&weights,
		&modulo,
		&complement,
		&method,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbIbans.IBAN{
		Id:         id,
		ValidFrom:  util.GetSQLNullTime(validFrom),
		Algorithm:  algorithm,
		Format:     *util.GetSQLNullString(format),
		Weights:    util.GetSQLNullString(weights),
		Modulo:     util.GetSQLNullString(modulo),
		Complement: util.GetSQLNullString(complement),
		Method:     util.GetSQLNullString(method),
		Status:     status,

		Country: &countries.Country{
			Id: countryID,
		},
	}, nil
}

func (s *IbanRepository) ScanWithRelationship(row pgx.Row) (*pbIbans.IBAN, error) {
	var (
		id         string
		countryID  string
		validFrom  sql.NullTime
		algorithm  pbIbans.Algorithm
		format     sql.NullString
		weights    sql.NullString
		modulo     sql.NullString
		complement sql.NullString
		method     sql.NullString
		status     pbCommon.Status
	)

	var (
		countryCode                   string
		countryName                   sql.NullString
		countryIcon                   sql.NullString
		countryIso3166A3              sql.NullString
		countryIso3166Num             sql.NullString
		countryInternetCctld          sql.NullString
		countryRegion                 sql.NullString
		countrySubRegion              sql.NullString
		countryIntermediateRegion     sql.NullString
		countryIntermediateRegionCode sql.NullString
		countryRegionCode             sql.NullString
		countrySubRegionCode          sql.NullString
		countryStatus                 pbCommon.Status
	)

	err := row.Scan(
		&id,
		&countryID,
		&validFrom,
		&algorithm,
		&format,
		&weights,
		&modulo,
		&complement,
		&method,
		&status,
		&countryID,
		&countryCode,
		&countryName,
		&countryIcon,
		&countryIso3166A3,
		&countryIso3166Num,
		&countryInternetCctld,
		&countryRegion,
		&countrySubRegion,
		&countryIntermediateRegion,
		&countryIntermediateRegionCode,
		&countryRegionCode,
		&countrySubRegionCode,
		&countryStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbIbans.IBAN{
		Id:         id,
		ValidFrom:  util.GetSQLNullTime(validFrom),
		Algorithm:  algorithm,
		Format:     *util.GetSQLNullString(format),
		Weights:    util.GetSQLNullString(weights),
		Modulo:     util.GetSQLNullString(modulo),
		Complement: util.GetSQLNullString(complement),
		Method:     util.GetSQLNullString(method),
		Status:     status,

		Country: &countries.Country{
			Id:                     countryID,
			Code:                   countryCode,
			Name:                   util.GetSQLNullString(countryName),
			Icon:                   util.GetSQLNullString(countryIcon),
			Iso3166A3:              util.GetSQLNullString(countryIso3166A3),
			Iso3166Num:             util.GetSQLNullString(countryIso3166Num),
			InternetCctld:          util.GetSQLNullString(countryInternetCctld),
			Region:                 util.GetSQLNullString(countryRegion),
			SubRegion:              util.GetSQLNullString(countrySubRegion),
			IntermediateRegion:     util.GetSQLNullString(countryIntermediateRegion),
			IntermediateRegionCode: util.GetSQLNullString(countryIntermediateRegionCode),
			RegionCode:             util.GetSQLNullString(countryRegionCode),
			SubRegionCode:          util.GetSQLNullString(weights),
			Status:                 countryStatus,
			Fiats:                  nil,
			Cryptos:                nil,
		},
	}, nil
}

func createCountryValidityFB(
	countryValidity *pbIbans.CountryValidity,
) *util.FilterBracket {
	// Create a countries qb
	countriesQb := util.CreateQueryBuilder(util.Select, _tableNameCountries)
	countriesQb.Select("countries.id")
	switch countryValidity.GetCountry().Select.(type) {
	case *countries.Select_ByCode:
		countriesQb.Where(
			"code = ?",
			countryValidity.GetCountry().GetByCode(),
		)
	case *countries.Select_ById:
		countriesQb.Where(
			"id = ?",
			countryValidity.GetCountry().GetById(),
		)
	}
	countriesSQL, countriesArgs, _ := countriesQb.GenerateSQL()

	// Create country validity filter bracket
	countryValidityFB := util.CreateFilterBracket("AND")
	countryValidityFB.SetFilter(
		fmt.Sprintf("countries.id IN (%s)", countriesSQL),
		countriesArgs...,
	)
	countryValidityFB.SetFilter("valid_from = ?", util.GetDBTimestampValue(countryValidity.GetValidFrom()))

	return countryValidityFB
}
