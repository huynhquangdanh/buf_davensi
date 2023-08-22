package countries

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbCountriesConnect "davensi.com/core/gen/countries/countriesconnect"
	"davensi.com/core/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_fields = "id, code, name, icon, iso3166_a3, " +
		"iso3166_num, internet_cctld, region, sub_region" +
		", intermediate_region, intermediate_region_code, region_code, sub_region_code, status"
	_tableName = "core.countries"
)

type CountryRepository struct {
	pbCountriesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewCountryRepository(db *pgxpool.Pool) *CountryRepository {
	return &CountryRepository{
		db: db,
	}
}

func SetQBBySelect(selectCountries *pbCountries.Select, qb *util.QueryBuilder, alias string) {
	if selectCountries == nil {
		return
	}

	if alias != "" {
		alias += "."
	}

	switch selectCountries.Select.(type) {
	case *pbCountries.Select_ById:
		qb.Where(fmt.Sprintf("%sid = ?", alias), selectCountries.GetById())
	case *pbCountries.Select_ByCode:
		qb.Where(fmt.Sprintf("%scode = ?", alias), selectCountries.GetByCode())
	}
}

func handleTypicalCountryFields(msg *pbCountries.Country, handleFn func(field string, value any)) {
	if msg.Name != nil {
		handleFn("countries.name", msg.GetName())
	}
	if msg.Icon != nil {
		handleFn("countries.icon", msg.GetIcon())
	}
	if msg.Iso3166A3 != nil {
		handleFn("countries.iso3166_a3", msg.GetIso3166A3())
	}
	if msg.Iso3166Num != nil {
		handleFn("countries.iso3166_num", msg.GetIso3166Num())
	}
	if msg.InternetCctld != nil {
		handleFn("countries.internet_cctld", msg.GetInternetCctld())
	}
	if msg.Region != nil {
		handleFn("countries.region", msg.GetRegion())
	}
	if msg.SubRegion != nil {
		handleFn("countries.sub_region", msg.GetSubRegion())
	}
	if msg.IntermediateRegion != nil {
		handleFn("countries.intermediate_region", msg.GetIntermediateRegion())
	}
	if msg.IntermediateRegionCode != nil {
		handleFn("countries.intermediate_region_code", msg.GetIntermediateRegionCode())
	}
	if msg.RegionCode != nil {
		handleFn("countries.region_code", msg.GetRegionCode())
	}
	if msg.SubRegionCode != nil {
		handleFn("countries.sub_region_code", msg.GetSubRegionCode())
	}
}

func (s *CountryRepository) QbInsert(msg *pbCountries.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("code")
	singleValue = append(singleValue, msg.GetCode())

	// Append optional fields values
	handleTypicalCountryFields(
		&pbCountries.Country{
			Name:                   msg.Name,
			Icon:                   msg.Icon,
			Iso3166A3:              msg.Iso3166A3,
			Iso3166Num:             msg.Iso3166Num,
			InternetCctld:          msg.InternetCctld,
			Region:                 msg.Region,
			SubRegion:              msg.SubRegion,
			IntermediateRegion:     msg.IntermediateRegion,
			IntermediateRegionCode: msg.IntermediateRegionCode,
			RegionCode:             msg.RegionCode,
			SubRegionCode:          msg.SubRegionCode,
		},
		func(field string, value any) {
			qb.SetInsertField(field)
			singleValue = append(singleValue, value)
		},
	)
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleValue = append(singleValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleValue)

	return qb, err
}

func (s *CountryRepository) QbUpdate(msg *pbCountries.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Code != nil {
		qb.SetUpdate("code", msg.GetCode())
	}
	handleTypicalCountryFields(
		&pbCountries.Country{
			Name:                   msg.Name,
			Icon:                   msg.Icon,
			Iso3166A3:              msg.Iso3166A3,
			Iso3166Num:             msg.Iso3166Num,
			InternetCctld:          msg.InternetCctld,
			Region:                 msg.Region,
			SubRegion:              msg.SubRegion,
			IntermediateRegion:     msg.IntermediateRegion,
			IntermediateRegionCode: msg.IntermediateRegionCode,
			RegionCode:             msg.RegionCode,
			SubRegionCode:          msg.SubRegionCode,
		},
		func(field string, value any) {
			qb.SetUpdate(field, value)
		},
	)
	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	SetQBBySelect(msg.Select, qb, "")

	return qb, nil
}

func (s *CountryRepository) QbGetOne(msg *pbCountries.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	SetQBBySelect(msg.Select, qb, "")

	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *CountryRepository) QbGetList(msg *pbCountries.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, "countries"))

	if msg.Code != nil {
		qb.Where("countries.code LIKE '%' || ? || '%'", msg.GetCode())
	}
	handleTypicalCountryFields(
		&pbCountries.Country{
			Name:                   msg.Name,
			Icon:                   msg.Icon,
			Iso3166A3:              msg.Iso3166A3,
			Iso3166Num:             msg.Iso3166Num,
			InternetCctld:          msg.InternetCctld,
			Region:                 msg.Region,
			SubRegion:              msg.SubRegion,
			IntermediateRegion:     msg.IntermediateRegion,
			IntermediateRegionCode: msg.IntermediateRegionCode,
			RegionCode:             msg.RegionCode,
			SubRegionCode:          msg.SubRegionCode,
		},
		func(field string, value any) {
			qb.Where(
				fmt.Sprintf("%s LIKE '%%' || ? || '%%'", field),
				msg.GetCode(),
			)
		},
	)
	if msg.Status != nil {
		statuses := msg.GetStatus().GetList()

		if len(statuses) > 0 {
			args := []any{}
			for _, v := range statuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"countries.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(statuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *CountryRepository) ScanRow(row pgx.Row) (*pbCountries.Country, error) {
	var (
		id                     string
		code                   string
		name                   sql.NullString
		icon                   sql.NullString
		iso3166A3              sql.NullString
		iso3166Num             sql.NullString
		internetCctld          sql.NullString
		region                 sql.NullString
		subRegion              sql.NullString
		intermediateRegion     sql.NullString
		intermediateRegionCode sql.NullString
		regionCode             sql.NullString
		subRegionCode          sql.NullString
		status                 pbCommon.Status
	)

	err := row.Scan(
		&id,
		&code,
		&name,
		&icon,
		&iso3166A3,
		&iso3166Num,
		&internetCctld,
		&region,
		&subRegion,
		&intermediateRegion,
		&intermediateRegionCode,
		&regionCode,
		&subRegionCode,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbCountries.Country{
		Id:                     id,
		Code:                   code,
		Name:                   util.GetSQLNullString(name),
		Icon:                   util.GetSQLNullString(icon),
		Iso3166A3:              util.GetSQLNullString(iso3166A3),
		Iso3166Num:             util.GetSQLNullString(iso3166Num),
		InternetCctld:          util.GetSQLNullString(internetCctld),
		Region:                 util.GetSQLNullString(region),
		SubRegion:              util.GetSQLNullString(subRegion),
		IntermediateRegion:     util.GetSQLNullString(intermediateRegion),
		IntermediateRegionCode: util.GetSQLNullString(intermediateRegionCode),
		RegionCode:             util.GetSQLNullString(regionCode),
		SubRegionCode:          util.GetSQLNullString(subRegionCode),
		Status:                 status,
	}, nil
}
