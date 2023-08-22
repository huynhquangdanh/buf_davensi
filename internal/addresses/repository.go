package addresses

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	pbAddresses "davensi.com/core/gen/addresses"
	pbAddressesConnect "davensi.com/core/gen/addresses/addressesconnect"
	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbKyc "davensi.com/core/gen/kyc"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_fields = "id, type, country_id, building, floor, unit," +
		" street_num, street_name, district, locality, zip_code, region, state, status"
	_tableName = "core.addresses"
)

type AddressRepository struct {
	pbAddressesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

var (
	singleRepo *AddressRepository
	onceRepo   sync.Once
)

func NewAddressRepository(db *pgxpool.Pool) *AddressRepository {
	return &AddressRepository{
		db: db,
	}
}

func GetSingletonRepository(db *pgxpool.Pool) *AddressRepository {
	onceRepo.Do(func() {
		singleRepo = NewAddressRepository(db)
	})
	return singleRepo
}

func (s *AddressRepository) QbInsertManyAddress(msg *pbAddresses.SetLabeledAddressList) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	qb.SetInsertField(
		"type", "country_id", "building", "floor", "unit", "street_num",
		"street_name", "district", "locality", "zip_code", "region",
		"state", "status",
	)
	for _, labelAddress := range msg.GetList() {
		if labelAddress.GetAddress() != nil {
			createReq := labelAddress.GetAddress()
			_, err := qb.SetInsertValues([]any{
				createReq.GetType(),
				createReq.GetCountry().GetById(),
				createReq.GetBuilding(),
				createReq.GetFloor(),
				createReq.GetUnit(),
				createReq.GetStreetNum(),
				createReq.GetStreetName(),
				createReq.GetDistrict(),
				createReq.GetLocality(),
				createReq.GetZipCode(),
				createReq.GetRegion(),
				createReq.GetState(),
				createReq.GetStatus(),
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return qb, nil
}

func (s *AddressRepository) QbInsert(msg *pbAddresses.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleAddressValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("type").SetInsertField("country_id")
	singleAddressValue = append(singleAddressValue, msg.GetType(), msg.GetCountry().GetById())

	// Append optional fields values
	if msg.Building != nil {
		qb.SetInsertField("building")
		singleAddressValue = append(singleAddressValue, msg.GetBuilding())
	}
	if msg.Floor != nil {
		qb.SetInsertField("floor")
		singleAddressValue = append(singleAddressValue, msg.GetFloor())
	}
	if msg.Unit != nil {
		qb.SetInsertField("unit")
		singleAddressValue = append(singleAddressValue, msg.GetUnit())
	}
	if msg.StreetNum != nil {
		qb.SetInsertField("street_num")
		singleAddressValue = append(singleAddressValue, msg.GetStreetNum())
	}
	if msg.StreetName != nil {
		qb.SetInsertField("street_name")
		singleAddressValue = append(singleAddressValue, msg.GetStreetName())
	}
	if msg.District != nil {
		qb.SetInsertField("district")
		singleAddressValue = append(singleAddressValue, msg.GetDistrict())
	}
	if msg.Locality != nil {
		qb.SetInsertField("locality")
		singleAddressValue = append(singleAddressValue, msg.GetLocality())
	}
	if msg.ZipCode != nil {
		qb.SetInsertField("zip_code")
		singleAddressValue = append(singleAddressValue, msg.GetZipCode())
	}
	if msg.Region != nil {
		qb.SetInsertField("region")
		singleAddressValue = append(singleAddressValue, msg.GetRegion())
	}
	if msg.State != nil {
		qb.SetInsertField("state")
		singleAddressValue = append(singleAddressValue, msg.GetState())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleAddressValue = append(singleAddressValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleAddressValue)

	return qb, err
}

func (s *AddressRepository) QbUpdate(msg *pbAddresses.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Type != nil {
		qb.SetUpdate("type", msg.GetType())
	}
	if msg.Country != nil {
		qb.SetUpdate("country_id", msg.GetCountry().GetById())
	}
	if msg.Building != nil {
		qb.SetUpdate("building", msg.GetBuilding())
	}
	if msg.Floor != nil {
		qb.SetUpdate("floor", msg.GetFloor())
	}
	if msg.Unit != nil {
		qb.SetUpdate("unit", msg.GetUnit())
	}
	if msg.StreetNum != nil {
		qb.SetUpdate("street_num", msg.GetStreetNum())
	}
	if msg.StreetName != nil {
		qb.SetUpdate("street_name", msg.GetStreetName())
	}
	if msg.District != nil {
		qb.SetUpdate("district", msg.GetDistrict())
	}
	if msg.Locality != nil {
		qb.SetUpdate("locality", msg.GetLocality())
	}
	if msg.ZipCode != nil {
		qb.SetUpdate("zip_code", msg.GetZipCode())
	}
	if msg.Region != nil {
		qb.SetUpdate("region", msg.GetRegion())
	}
	if msg.State != nil {
		qb.SetUpdate("state", msg.GetState())
	}
	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("id = ?", msg.GetId())

	return qb, nil
}

func (s *AddressRepository) QbGetOne(msg *pbAddresses.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	qb.Where("id = ?", msg.GetId())
	qb.Where("status = ?", pbKyc.Status_STATUS_VALIDATED)

	return qb
}

func handleTypicalAddressFields(msg *pbAddresses.Address, handleFn func(field string, value any)) {
	if msg.Building != nil {
		handleFn("addresses.building", msg.GetBuilding())
	}
	if msg.Floor != nil {
		handleFn("addresses.floor", msg.GetFloor())
	}
	if msg.Unit != nil {
		handleFn("addresses.unit", msg.GetUnit())
	}
	if msg.StreetNum != nil {
		handleFn("addresses.street_num", msg.GetStreetNum())
	}
	if msg.District != nil {
		handleFn("addresses.district", msg.GetDistrict())
	}
	if msg.Locality != nil {
		handleFn("addresses.locality", msg.GetLocality())
	}
	if msg.ZipCode != nil {
		handleFn("addresses.zip_code", msg.GetZipCode())
	}
	if msg.Region != nil {
		handleFn("addresses.region", msg.GetRegion())
	}
	if msg.State != nil {
		handleFn("addresses.state", msg.GetState())
	}
}

func (s *AddressRepository) QbGetList(msg *pbAddresses.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, "addresses"))

	if msg.Type != nil {
		addressTypes := msg.GetType().GetList()

		if len(addressTypes) > 0 {
			args := []any{}
			for _, v := range addressTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"addresses.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(addressTypes)), ""), ", "),
				),
				args...,
			)
		}
	}
	handleTypicalAddressFields(
		&pbAddresses.Address{
			Building:   msg.Building,
			Floor:      msg.Floor,
			Unit:       msg.Unit,
			StreetNum:  msg.StreetNum,
			StreetName: msg.StreetName,
			District:   msg.District,
			Locality:   msg.Locality,
			ZipCode:    msg.ZipCode,
			Region:     msg.Region,
			State:      msg.State,
		},
		func(field string, value any) {
			qb.Where(fmt.Sprintf("%s LIKE '%%' || ? || '%%'", field), value)
		},
	)
	// TODO: country condition when implement relationship
	if msg.Status != nil {
		socialStatuses := msg.GetStatus().GetList()

		if len(socialStatuses) > 0 {
			args := []any{}
			for _, v := range socialStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"addresses.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(socialStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *AddressRepository) QbDeleteMany(ids []string) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)
	qb.SetUpdate("status", pbKyc.Status_STATUS_CANCELED)

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	args := []any{}
	for _, v := range ids {
		args = append(args, v)
	}
	qb.Where(
		fmt.Sprintf(
			"id IN (%s)",
			strings.Join(strings.Split(strings.Repeat("?", len(ids)), ""), ", "),
		),
		args...,
	)

	return qb, nil
}

func (s *AddressRepository) ScanRow(row pgx.Row) (*pbAddresses.Address, error) {
	var (
		id                 string
		addressType        pbAddresses.Type
		countryID          string
		nullableBuilding   sql.NullString
		building           *string
		nullableFloor      sql.NullString
		floor              *string
		nullableUnit       sql.NullString
		unit               *string
		nullableStreetNum  sql.NullString
		streetNum          *string
		nullableStreetName sql.NullString
		streetName         *string
		nullableDistrict   sql.NullString
		district           *string
		nullableLocality   sql.NullString
		locality           *string
		nullableZipCode    sql.NullString
		zipCode            *string
		nullableRegion     sql.NullString
		region             *string
		nullableState      sql.NullString
		state              *string
		status             pbCommon.Status
	)

	err := row.Scan(
		&id,
		&addressType,
		&countryID,
		&nullableBuilding,
		&nullableFloor,
		&nullableUnit,
		&nullableStreetNum,
		&nullableStreetName,
		&nullableDistrict,
		&nullableLocality,
		&nullableZipCode,
		&nullableRegion,
		&nullableState,
		&status,
	)
	if err != nil {
		return nil, err
	}

	if nullableBuilding.Valid {
		building = &nullableBuilding.String
	}
	if nullableFloor.Valid {
		floor = &nullableFloor.String
	}
	if nullableUnit.Valid {
		unit = &nullableUnit.String
	}
	if nullableStreetNum.Valid {
		streetNum = &nullableStreetNum.String
	}
	if nullableStreetName.Valid {
		streetName = &nullableStreetName.String
	}
	if nullableDistrict.Valid {
		district = &nullableDistrict.String
	}
	if nullableLocality.Valid {
		locality = &nullableLocality.String
	}
	if nullableZipCode.Valid {
		zipCode = &nullableZipCode.String
	}
	if nullableRegion.Valid {
		region = &nullableRegion.String
	}
	if nullableState.Valid {
		state = &nullableState.String
	}

	return &pbAddresses.Address{
		Id:         id,
		Type:       addressType,
		Country:    &pbCountries.Country{Id: countryID},
		Building:   building,
		Floor:      floor,
		Unit:       unit,
		StreetNum:  streetNum,
		StreetName: streetName,
		District:   district,
		Locality:   locality,
		ZipCode:    zipCode,
		Region:     region,
		State:      state,
		Status:     status,
	}, nil
}

func (s *AddressRepository) ScanWithRelationship(row pgx.Row) (*pbAddresses.Address, error) {
	var (
		id          string
		addressType pbAddresses.Type
		building    sql.NullString
		floor       sql.NullString
		unit        sql.NullString
		streetNum   sql.NullString
		streetName  sql.NullString
		district    sql.NullString
		locality    sql.NullString
		zipCode     sql.NullString
		region      sql.NullString
		state       sql.NullString
		status      pbCommon.Status
	)

	var (
		countryID                     string
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
		&addressType,
		&countryID,
		&building,
		&floor,
		&unit,
		&streetNum,
		&streetName,
		&district,
		&locality,
		&zipCode,
		&region,
		&state,
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

	return &pbAddresses.Address{
		Id:   id,
		Type: addressType,
		Country: &pbCountries.Country{
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
			SubRegionCode:          util.GetSQLNullString(countrySubRegionCode),
			Status:                 countryStatus,
		},
		Building:   util.GetSQLNullString(building),
		Floor:      util.GetSQLNullString(floor),
		Unit:       util.GetSQLNullString(unit),
		StreetNum:  util.GetSQLNullString(streetNum),
		StreetName: util.GetSQLNullString(streetName),
		District:   util.GetSQLNullString(district),
		Locality:   util.GetSQLNullString(locality),
		ZipCode:    util.GetSQLNullString(zipCode),
		Region:     util.GetSQLNullString(region),
		State:      util.GetSQLNullString(state),
		Status:     status,
	}, nil
}

func (s *AddressRepository) QbInsertWithUUID(msg *pbAddresses.CreateRequest, uuid string) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleAddressValue := []any{}

	// Append required field
	qb.SetInsertField("id")
	singleAddressValue = append(singleAddressValue, uuid)

	qb.SetInsertField("type").SetInsertField("country_id")
	singleAddressValue = append(singleAddressValue, msg.GetType(), msg.GetCountry().GetById())

	// Append optional fields values
	if msg.Building != nil {
		qb.SetInsertField("building")
		singleAddressValue = append(singleAddressValue, msg.GetBuilding())
	}
	if msg.Floor != nil {
		qb.SetInsertField("floor")
		singleAddressValue = append(singleAddressValue, msg.GetFloor())
	}
	if msg.Unit != nil {
		qb.SetInsertField("unit")
		singleAddressValue = append(singleAddressValue, msg.GetUnit())
	}
	if msg.StreetNum != nil {
		qb.SetInsertField("street_num")
		singleAddressValue = append(singleAddressValue, msg.GetStreetNum())
	}
	if msg.StreetName != nil {
		qb.SetInsertField("street_name")
		singleAddressValue = append(singleAddressValue, msg.GetStreetName())
	}
	if msg.District != nil {
		qb.SetInsertField("district")
		singleAddressValue = append(singleAddressValue, msg.GetDistrict())
	}
	if msg.Locality != nil {
		qb.SetInsertField("locality")
		singleAddressValue = append(singleAddressValue, msg.GetLocality())
	}
	if msg.ZipCode != nil {
		qb.SetInsertField("zip_code")
		singleAddressValue = append(singleAddressValue, msg.GetZipCode())
	}
	if msg.Region != nil {
		qb.SetInsertField("region")
		singleAddressValue = append(singleAddressValue, msg.GetRegion())
	}
	if msg.State != nil {
		qb.SetInsertField("state")
		singleAddressValue = append(singleAddressValue, msg.GetState())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleAddressValue = append(singleAddressValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleAddressValue)

	return qb, err
}
