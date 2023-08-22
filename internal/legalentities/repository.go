package legalentities

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	pbCountries "davensi.com/core/gen/countries"
	pbKyc "davensi.com/core/gen/kyc"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbLegalEntitiesConnect "davensi.com/core/gen/legalentities/legalentitiesconnect"
	pbUoMs "davensi.com/core/gen/uoms"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	_tableName    = "core.legalentities"
	_tableContact = "core.legalentities_contacts"
	_fields       = "id, name, type, incorporation_country_id, incorporation_locality" +
		", business_registration_no, business_registration_alt_no," +
		" valid_until, tax_id, currency1_id, currency2_id, currency3_id, status"
	_fieldsContact = "legalentity_id, label, contact_id, status"
)

type LegalEntitiesRepository struct {
	pbLegalEntitiesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewLegalEntitiesRepository(db *pgxpool.Pool) *LegalEntitiesRepository {
	return &LegalEntitiesRepository{
		db: db,
	}
}

func SetQBBySelect(selectPrice *pbLegalEntities.Select, qb *util.QueryBuilder) {
	switch selectPrice.GetSelect().(type) {
	case *pbLegalEntities.Select_ById:
		qb.Where("legalentities.id = ?", selectPrice.GetById())
	case *pbLegalEntities.Select_ByName:
		qb.Where("legalentities.name = ?", selectPrice.GetByName())
	}
}

func (s *LegalEntitiesRepository) QbInsert(msg *pbLegalEntities.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleLedgerValue := []any{}

	qb.SetInsertField("name").
		SetInsertField("type").
		SetInsertField("currency1_id").
		SetInsertField("incorporation_country_id")

	singleLedgerValue = append(singleLedgerValue,
		msg.GetName(),
		msg.GetType(),
		msg.GetIncorporationCountry().GetById(),
		msg.GetCurrency1().GetById(),
	)

	// Append optional fields values
	if msg.IncorporationLocality != nil {
		qb.SetInsertField("incorporation_locality")
		singleLedgerValue = append(singleLedgerValue, msg.GetIncorporationLocality())
	}
	if msg.BusinessRegistrationNo != nil {
		qb.SetInsertField("business_registration_no")
		singleLedgerValue = append(singleLedgerValue, msg.GetBusinessRegistrationNo())
	}
	if msg.BusinessRegistrationAltNo != nil {
		qb.SetInsertField("business_registration_alt_no")
		singleLedgerValue = append(singleLedgerValue, msg.GetBusinessRegistrationAltNo())
	}
	if msg.ValidUntil != nil {
		qb.SetInsertField("valid_until")
		singleLedgerValue = append(singleLedgerValue, util.GetDBTimestampValue(msg.GetValidUntil()))
	}
	if msg.TaxId != nil {
		qb.SetInsertField("tax_id")
		singleLedgerValue = append(singleLedgerValue, msg.GetTaxId())
	}
	if msg.Currency2 != nil && msg.GetCurrency2().GetById() != "" {
		qb.SetInsertField("currency2_id")
		singleLedgerValue = append(singleLedgerValue, msg.GetCurrency2().GetById())
	}
	if msg.Currency3 != nil && msg.GetCurrency3().GetById() != "" {
		qb.SetInsertField("currency3_id")
		singleLedgerValue = append(singleLedgerValue, msg.GetCurrency3().GetById())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleLedgerValue = append(singleLedgerValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleLedgerValue)

	return qb, err
}

func handleTypicalLegalEntityFields(msg *pbLegalEntities.LegalEntity, handleFn func(field string, value any)) {
	if msg.IncorporationLocality != nil {
		handleFn("incorporation_locality", msg.GetIncorporationLocality())
	}
	if msg.BusinessRegistrationNo != nil {
		handleFn("business_registration_no", msg.GetBusinessRegistrationNo())
	}
	if msg.BusinessRegistrationAltNo != nil {
		handleFn("business_registration_alt_no", msg.GetBusinessRegistrationAltNo())
	}
	if msg.TaxId != nil {
		handleFn("tax_id", msg.GetTaxId())
	}
}

func (s *LegalEntitiesRepository) QbUpdate(msg *pbLegalEntities.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Name != nil {
		qb.SetUpdate("name", msg.GetName())
	}
	if msg.Type != nil {
		qb.SetUpdate("type", msg.GetType())
	}
	if msg.IncorporationCountry != nil {
		qb.SetUpdate("incorporation_country_id", msg.GetIncorporationCountry().GetById())
	}
	if msg.ValidUntil != nil {
		qb.SetUpdate("valid_until", util.GetDBTimestampValue(msg.GetValidUntil()))
	}
	if msg.Currency1 != nil {
		qb.SetUpdate("currency1_id", msg.GetCurrency1().GetById())
	}
	if msg.Currency2 != nil {
		qb.SetUpdate("currency2_id", msg.GetCurrency2().GetById())
	}
	if msg.Currency3 != nil {
		qb.SetUpdate("currency3_id", msg.GetCurrency3().GetById())
	}
	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}
	handleTypicalLegalEntityFields(
		&pbLegalEntities.LegalEntity{
			IncorporationLocality:     msg.IncorporationLocality,
			BusinessRegistrationNo:    msg.BusinessRegistrationNo,
			BusinessRegistrationAltNo: msg.BusinessRegistrationAltNo,
			TaxId:                     msg.TaxId,
		},
		func(field string, value any) {
			qb.SetUpdate(field, value)
		},
	)
	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	SetQBBySelect(msg.GetSelect(), qb)

	return qb, nil
}

func (s *LegalEntitiesRepository) QbGetOne(msg *pbLegalEntities.GetRequest) *util.QueryBuilder {
	qb := util.
		CreateQueryBuilder(util.Select, _tableName).
		Select(util.GetFieldsWithTableName(_fields, _tableName)).
		Where("legalentities.status = ?", pbCommon.Status_STATUS_ACTIVE)

	SetQBBySelect(msg.GetSelect(), qb)

	return qb
}

func appendFilterByRangeValue(msg *pbLegalEntities.GetListRequest, qb *util.QueryBuilder) {
	if msg.Type != nil {
		legalEntityTypes := msg.GetType().GetList()

		if len(legalEntityTypes) > 0 {
			args := []any{}
			for _, v := range legalEntityTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"legalentities.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(legalEntityTypes)), ""), ", "),
				),
				args...,
			)
		}
	}

	if msg.Status != nil {
		legalEntityStatuses := msg.GetStatus().GetList()

		if len(legalEntityStatuses) > 0 {
			args := []any{}
			for _, v := range legalEntityStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"legalentities.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(legalEntityStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}
}

func (s *LegalEntitiesRepository) QbGetList(msg *pbLegalEntities.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, "legalentities"))

	appendFilterByRangeValue(msg, qb)

	if msg.Name != nil {
		qb.Where("legalentities.name LIKE '%' || ? || '%'", msg.GetName())
	}
	if msg.IncorporationLocality != nil {
		qb.Where("legalentities.incorporation_locality LIKE '%' || ? || '%'", msg.GetIncorporationLocality())
	}
	if msg.BusinessRegistrationNo != nil {
		qb.Where("legalentities.business_registration_no LIKE '%' || ? || '%'", msg.GetBusinessRegistrationNo())
	}
	if msg.BusinessRegistrationAltNo != nil {
		qb.Where("legalentities.business_registration_alt_no LIKE '%' || ? || '%'", msg.GetBusinessRegistrationAltNo())
	}
	if msg.ValidUntilFrom != nil && len(msg.GetValidUntilFrom().GetList()) > 0 {
		decimalFilter := GetDecimalsFB(msg.GetValidUntilFrom(), "legalentities.valid_until")
		sqlStr, args := decimalFilter.GenerateSQL()

		qb.Where(sqlStr, args...)
	}
	if msg.ValidUntilTo != nil && len(msg.GetValidUntilTo().GetList()) > 0 {
		decimalFilter := GetDecimalsFB(msg.GetValidUntilTo(), "legalentities.valid_until")
		sqlStr, args := decimalFilter.GenerateSQL()

		qb.Where(sqlStr, args...)
	}
	if msg.TaxId != nil {
		qb.Where("legalentities.tax_id LIKE '%' || ? || '%'", msg.GetTaxId())
	}

	return qb
}

func (s *LegalEntitiesRepository) ScanID(row pgx.Row) (string, error) {
	var id string
	err := row.Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *LegalEntitiesRepository) ScanMainEntity(row pgx.Row) (*pbLegalEntities.LegalEntity, error) {
	var (
		id                        string
		name                      string
		legalEntitytype           pbLegalEntities.Type
		incorporationCountryID    string
		incorporationLocality     sql.NullString
		businessRegistrationNo    sql.NullString
		businessRegistrationAltNo sql.NullString
		validUntil                sql.NullTime
		taxID                     sql.NullString
		currency1ID               string
		currency2ID               sql.NullString
		currency3ID               sql.NullString
		status                    pbCommon.Status
	)

	err := row.Scan(
		&id,
		&name,
		&legalEntitytype,
		&incorporationCountryID,
		&incorporationLocality,
		&businessRegistrationNo,
		&businessRegistrationAltNo,
		&validUntil,
		&taxID,
		&currency1ID,
		&currency2ID,
		&currency3ID,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbLegalEntities.LegalEntity{
		Id:                        id,
		Name:                      name,
		Type:                      legalEntitytype,
		IncorporationCountry:      &pbCountries.Country{Id: incorporationCountryID},
		IncorporationLocality:     util.GetSQLNullString(incorporationLocality),
		BusinessRegistrationNo:    util.GetSQLNullString(businessRegistrationNo),
		BusinessRegistrationAltNo: util.GetSQLNullString(businessRegistrationAltNo),
		ValidUntil:                util.GetSQLNullTime(validUntil),
		TaxId:                     util.GetSQLNullString(taxID),
		Currency1:                 &pbUoMs.UoM{Id: currency1ID},
		Currency2:                 &pbUoMs.UoM{Id: util.GetPointString(util.GetSQLNullString(currency2ID))},
		Currency3:                 &pbUoMs.UoM{Id: util.GetPointString(util.GetSQLNullString(currency3ID))},
		Status:                    status,
		Addresses:                 nil,
		Contacts:                  nil,
	}, nil
}

func setRangeCondition(
	expression, field string,
	value *pbCommon.TimestampBoundary,
	filterBracket *util.FilterBracket,
) {
	switch value.GetBoundary().(type) {
	case *pbCommon.TimestampBoundary_Incl:
		filterBracket.SetFilter(
			fmt.Sprintf("%s %s= ?", field, expression),
			util.GetDBTimestampValue(value.GetBoundary().(*pbCommon.TimestampBoundary_Incl).Incl),
		)
	case *pbCommon.TimestampBoundary_Excl:
		filterBracket.SetFilter(
			fmt.Sprintf("%s %s ?", field, expression),
			util.GetDBTimestampValue(value.GetBoundary().(*pbCommon.TimestampBoundary_Excl).Excl),
		)
	}
}

func GetDecimalsFB(
	list *pbCommon.TimestampValueList,
	field string,
) *util.FilterBracket {
	filterBracket := util.CreateFilterBracket("OR")

	for _, v := range list.GetList() {
		switch v.GetSelect().(type) {
		case *pbCommon.TimestampValues_Single:
			filterBracket.SetFilter(
				fmt.Sprintf("%s = ?", field),
				util.GetDBTimestampValue(v.GetSelect().(*pbCommon.TimestampValues_Single).Single),
			)
		case *pbCommon.TimestampValues_Range:
			from := v.GetSelect().(*pbCommon.TimestampValues_Range).Range.From
			to := v.GetSelect().(*pbCommon.TimestampValues_Range).Range.To
			if v.GetSelect().(*pbCommon.TimestampValues_Range).Range.From != nil && v.GetSelect().(*pbCommon.TimestampValues_Range).Range.To != nil {
				setRangeCondition("<", field, from, filterBracket)
				setRangeCondition(">", field, to, filterBracket)
			} else if v.GetSelect().(*pbCommon.TimestampValues_Range).Range.From != nil {
				setRangeCondition(">", field, to, filterBracket)
			} else if v.GetSelect().(*pbCommon.TimestampValues_Range).Range.To != nil {
				setRangeCondition("<", field, from, filterBracket)
			}
		}
	}

	return filterBracket
}

func (s *LegalEntitiesRepository) ScanWithRelationship(row pgx.Row) (*pbLegalEntities.LegalEntity, error) {
	var ( // main table fields
		id                        string
		name                      string
		legalEntityType           uint32
		incorporationCountryID    string
		incorporationLocality     pgtype.Text
		businessRegistrationNo    pgtype.Text
		businessRegistrationAltNo pgtype.Text
		validUntil                pgtype.Timestamp
		taxID                     pgtype.Text
		currency1ID               string
		currency2ID               pgtype.Text
		currency3ID               pgtype.Text
		status                    uint32
	)

	var ( // country fields
		countryID                     pgtype.Text
		countryCode                   pgtype.Text
		countryName                   pgtype.Text
		countryIcon                   pgtype.Text
		countryIso3166A3              pgtype.Text
		countryIso3166Num             pgtype.Text
		countryInternetCctld          pgtype.Text
		countryRegion                 pgtype.Text
		countrySubRegion              pgtype.Text
		countryIntermediateRegion     pgtype.Text
		countryIntermediateRegionCode pgtype.Text
		countryRegionCode             pgtype.Text
		countrySubRegionCode          pgtype.Text
		countryStatus                 pgtype.Int2
	)

	var ( // uom1 fields
		uom1ID                pgtype.Text
		uom1Type              pgtype.Int2
		uom1Symbol            pgtype.Text
		uom1Name              pgtype.Text
		uom1Icon              pgtype.Text
		uom1ManagedDecimals   pgtype.Int2
		uom1DisplayedDecimals pgtype.Int2
		uom1ReportingUnit     pgtype.Bool
		uom1Status            pgtype.Int2
	)

	var ( // uom2 fields
		uom2ID                pgtype.Text
		uom2Type              pgtype.Int2
		uom2Symbol            pgtype.Text
		uom2Name              pgtype.Text
		uom2Icon              pgtype.Text
		uom2ManagedDecimals   pgtype.Int2
		uom2DisplayedDecimals pgtype.Int2
		uom2ReportingUnit     pgtype.Bool
		uom2Status            pgtype.Int2
	)

	var ( // uom3 fields
		uom3ID                pgtype.Text
		uom3Type              pgtype.Int2
		uom3Symbol            pgtype.Text
		uom3Name              pgtype.Text
		uom3Icon              pgtype.Text
		uom3ManagedDecimals   pgtype.Int2
		uom3DisplayedDecimals pgtype.Int2
		uom3ReportingUnit     pgtype.Bool
		uom3Status            pgtype.Int2
	)

	err := row.Scan(
		&id,
		&name,
		&legalEntityType,
		&incorporationCountryID,
		&incorporationLocality,
		&businessRegistrationNo,
		&businessRegistrationAltNo,
		&validUntil,
		&taxID,
		&currency1ID,
		&currency2ID,
		&currency3ID,
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
		&uom1ID,
		&uom1Type,
		&uom1Symbol,
		&uom1Name,
		&uom1Icon,
		&uom1ManagedDecimals,
		&uom1DisplayedDecimals,
		&uom1ReportingUnit,
		&uom1Status,
		&uom2ID,
		&uom2Type,
		&uom2Symbol,
		&uom2Name,
		&uom2Icon,
		&uom2ManagedDecimals,
		&uom2DisplayedDecimals,
		&uom2ReportingUnit,
		&uom2Status,
		&uom3ID,
		&uom3Type,
		&uom3Symbol,
		&uom3Name,
		&uom3Icon,
		&uom3ManagedDecimals,
		&uom3DisplayedDecimals,
		&uom3ReportingUnit,
		&uom3Status,
	)
	if err != nil {
		return nil, err
	}

	return &pbLegalEntities.LegalEntity{
		Id:   id,
		Name: name,
		Type: pbLegalEntities.Type(legalEntityType),
		IncorporationCountry: &pbCountries.Country{
			Id:                     countryID.String,
			Code:                   countryCode.String,
			Name:                   &countryName.String,
			Icon:                   &countryIcon.String,
			Iso3166A3:              &countryIso3166A3.String,
			Iso3166Num:             &countryIso3166Num.String,
			InternetCctld:          &countryInternetCctld.String,
			Region:                 &countryRegion.String,
			SubRegion:              &countrySubRegion.String,
			IntermediateRegion:     &countryIntermediateRegion.String,
			IntermediateRegionCode: &countryIntermediateRegionCode.String,
			RegionCode:             &countryRegion.String,
			SubRegionCode:          &countrySubRegionCode.String,
			Status:                 pbCommon.Status(countryStatus.Int16),
		},
		IncorporationLocality:     &incorporationLocality.String,
		BusinessRegistrationNo:    &businessRegistrationNo.String,
		BusinessRegistrationAltNo: &businessRegistrationAltNo.String,
		ValidUntil:                timestamppb.New(validUntil.Time),
		TaxId:                     &taxID.String,
		Currency1: &pbUoMs.UoM{
			Id:                uom1ID.String,
			Type:              pbUoMs.Type(uom1Type.Int16),
			Symbol:            uom1Symbol.String,
			Name:              &uom1Name.String,
			Icon:              &uom1Icon.String,
			ManagedDecimals:   uint32(uom1ManagedDecimals.Int16),
			DisplayedDecimals: uint32(uom1DisplayedDecimals.Int16),
			ReportingUnit:     uom1ReportingUnit.Bool,
			Status:            pbCommon.Status(uom1Status.Int16),
		},
		Currency2: &pbUoMs.UoM{
			Id:                uom2ID.String,
			Type:              pbUoMs.Type(uom2Type.Int16),
			Symbol:            uom2Symbol.String,
			Name:              &uom2Name.String,
			Icon:              &uom2Icon.String,
			ManagedDecimals:   uint32(uom2ManagedDecimals.Int16),
			DisplayedDecimals: uint32(uom2DisplayedDecimals.Int16),
			ReportingUnit:     uom2ReportingUnit.Bool,
			Status:            pbCommon.Status(uom2Status.Int16),
		},
		Currency3: &pbUoMs.UoM{
			Id:                uom3ID.String,
			Type:              pbUoMs.Type(uom3Type.Int16),
			Symbol:            uom3Symbol.String,
			Name:              &uom3Name.String,
			Icon:              &uom3Icon.String,
			ManagedDecimals:   uint32(uom3ManagedDecimals.Int16),
			DisplayedDecimals: uint32(uom3DisplayedDecimals.Int16),
			ReportingUnit:     uom3ReportingUnit.Bool,
			Status:            pbCommon.Status(uom3Status.Int16),
		},
		Status: pbCommon.Status(status),
	}, nil
}

// The query for GETLIST() use multiple JOINs on one same table, thus the need to modify the query after generating it
func generateAndTransformSQL(qb *util.QueryBuilder) (sqlStr string, sqlArgs []any) {
	replaceCount := 9 // number of times EACH 'uoms.{column}' apppears, to replace with aliases later

	sqlStr, sqlArgs, _ = qb.GenerateSQL()
	// Transform sqlstr to give aliases to core.uoms fields because there are multiple left-joins on that table
	sqlStr = strings.Replace(sqlStr, "uoms", "currency1", replaceCount)
	sqlStr = strings.Replace(sqlStr, "uoms", "currency2", replaceCount)
	sqlStr = strings.Replace(sqlStr, "uoms", "currency3", replaceCount)
	sqlStr = strings.Replace(sqlStr, "core.uoms ON legalentities.currency1_id = uoms.id",
		"core.uoms currency1 ON legalentities.currency1_id = currency1.id", 1)
	sqlStr = strings.Replace(sqlStr, "core.uoms ON legalentities.currency2_id = uoms.id",
		"core.uoms currency2 ON legalentities.currency2_id = currency2.id", 1)
	sqlStr = strings.Replace(sqlStr, "core.uoms ON legalentities.currency3_id = uoms.id",
		"core.uoms currency3 ON legalentities.currency3_id = currency3.id", 1)

	sqlStr = strings.Replace(sqlStr, "SELECT ", "SELECT DISTINCT ", 1)

	return sqlStr, sqlArgs
}

func (s *LegalEntitiesRepository) QbUpsertContacts(
	legalEntityID string,
	contacts []*pbContacts.SetLabeledContact,
	command string,
) (*util.QueryBuilder, error) {
	var queryType util.QueryType
	if command == "upsert" {
		queryType = util.Upsert
	} else {
		queryType = util.Insert
	}
	qb := util.CreateQueryBuilder(queryType, _tableName)
	qb.SetInsertField("legalentity_id", "label", "address_id", "status")

	for _, contact := range contacts {
		_, err := qb.SetInsertValues([]any{
			legalEntityID,
			contact.GetLabel(),
			contact.GetId(),
			contact.GetStatus().Enum(),
		})
		if err != nil {
			return nil, err
		}
	}

	return qb, nil
}

func (s *LegalEntitiesRepository) ScanRow(row pgx.Row) (*pbContacts.LabeledContact, error) {
	var (
		legalEntityID string
		label         string
		contactID     string
		mainContact   bool
		status        *pbKyc.Status
	)
	err := row.Scan(&legalEntityID, &label, &contactID, &mainContact, &status)
	if err != nil {
		return nil, err
	}
	return &pbContacts.LabeledContact{
		Label: label,
		Contact: &pbContacts.Contact{
			Id: contactID,
		},
		MainContact: &mainContact,
		Status:      *status.Enum(),
	}, nil
}

func (s *LegalEntitiesRepository) QbUpdateLegalEntitiesContact(
	legalEntityID string,
	req *pbContacts.UpdateLabeledContactRequest,
) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Update, _tableContact)

	if req.Status != nil {
		qb.SetUpdate("status", req.GetStatus()).
			Where("label = ? AND legalentity_id = ?", req.GetByLabel(), legalEntityID)
		return qb, nil
	}

	return nil, errors.New("cannot update without new value")
}

func (s *LegalEntitiesRepository) QbRemoveLegalEntitiesContacts(
	legalEntityID string,
	req *pbLegalEntities.LabelList,
) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Update, _tableContact)

	qb.SetUpdate("status", pbKyc.Status_STATUS_CANCELED)
	if !qb.IsUpdatable() {
		return qb, errors.New("cannot remove")
	}

	labels := req.GetList()

	args := []any{}
	for _, v := range labels {
		args = append(args, v)
	}

	qb.Where(
		fmt.Sprintf(
			"label IN(%s)",
			strings.Join(strings.Split(strings.Repeat("?", len(req.GetList())), ""), ", "),
		),
		args...,
	)

	qb.Where("legalentity_id = ?", legalEntityID)

	return qb, nil
}

func (s *LegalEntitiesRepository) QbGetOneContact(leID, label string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableContact)
	qb.Select(_fieldsContact)

	qb.Where("legalentity_id = ? AND label = ?", leID, label)

	return qb
}
