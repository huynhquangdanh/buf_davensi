package dvsubaccounts

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbDvSubAccounts "davensi.com/core/gen/dvsubaccounts"
	pbDvSubAccountsConnect "davensi.com/core/gen/dvsubaccounts/dvsubaccountsconnect"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbOrgs "davensi.com/core/gen/orgs"
	pbRecipients "davensi.com/core/gen/recipients"
	pbUoms "davensi.com/core/gen/uoms"
	pbUsers "davensi.com/core/gen/users"
	"davensi.com/core/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DvSubAccountRepository struct {
	pbDvSubAccountsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewDvSubAccountRepository(db *pgxpool.Pool) *DvSubAccountRepository {
	return &DvSubAccountRepository{
		db: db,
	}
}

func (s *DvSubAccountRepository) QbInsert(
	msg *pbDvSubAccounts.CreateRequest,
	recipient *pbRecipients.Recipient,
) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _table)
	singleDvbotValue := []any{}
	qb.SetInsertField("id")
	singleDvbotValue = append(singleDvbotValue, recipient.GetId())
	// Append optional fields values
	qb.SetInsertField("subaccount_type")
	singleDvbotValue = append(singleDvbotValue, msg.GetSubaccountType())

	if msg.Address != "" {
		qb.SetInsertField("address")
		singleDvbotValue = append(singleDvbotValue, msg.GetAddress())
	}

	_, err := qb.SetInsertValues(singleDvbotValue)
	return qb, err
}

func (s *DvSubAccountRepository) QbUpdate(msg *pbDvSubAccounts.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _table)
	if msg.GetSubaccountType() != pbDvSubAccounts.Type_TYPE_UNSPECIFIED {
		qb.SetUpdate("subaccount_type", msg.GetSubaccountType())
	}
	if msg.GetAddress() != "" {
		qb.SetUpdate("address", msg.GetAddress())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}
	switch msg.GetRecipient().GetSelect().GetSelect().(type) {
	case *pbRecipients.Select_ById:
		qb.Where("core.dvsubaccounts.id = ? ", msg.GetRecipient().GetSelect().GetById())
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		qb.Where(
			"core.dvsubaccounts.id = (SELECT core.recipients.id FROM core.recipients "+
				"WHERE core.recipients.legalentity_id = ? "+
				"AND core.recipients.user_id = ? "+
				"AND core.recipients.label = ?)",
			msg.GetRecipient().GetSelect().GetByLegalEntityUserLabel().GetLegalEntity().GetById(),
			msg.GetRecipient().GetSelect().GetByLegalEntityUserLabel().GetUser().GetById(),
			msg.GetRecipient().GetSelect().GetByLegalEntityUserLabel().GetLabel(),
		)
	}

	return qb, nil
}

func (s *DvSubAccountRepository) QbGetOne(_ *pbRecipients.GetRequest, recipientsQb *util.QueryBuilder) *util.QueryBuilder {
	recipientsQb.Join(fmt.Sprintf("LEFT JOIN %s ON dvsubaccounts.id = recipients.id", _table))
	fields := strings.Split(_fields, ",")
	for i := 0; i < len(fields); i++ {
		fields[i] = _table + "." + fields[i]
	}
	newFields := strings.Join(fields, ",")
	recipientsQb.Select(newFields)
	return recipientsQb
}

func (s *DvSubAccountRepository) QbGetList(msg *pbDvSubAccounts.GetListRequest, recipientsQb *util.QueryBuilder) *util.QueryBuilder {
	recipientsQb.Join(fmt.Sprintf("LEFT JOIN %s ON dvbots.id = recipients.id", _table))
	fields := strings.Split(_fields, ",")
	for i := 0; i < len(fields); i++ {
		fields[i] = _table + "." + fields[i]
	}
	newFields := strings.Join(fields, ",")
	recipientsQb.Select(newFields)
	recipientsQb.Where("core.recipients.type = ?", pbRecipients.Type_TYPE_DV_BOT)
	if msg.SubaccountType != nil {
		recipientsQb.Where("core.dvsubaccounts.subaccount_type = ?", msg.GetSubaccountType())
	}
	if msg.Address != nil {
		recipientsQb.Where("core.dvsubaccounts.address = ?", msg.GetAddress())
	}

	return recipientsQb
}

func (s *DvSubAccountRepository) ScanRow(row pgx.Row) (*pbDvSubAccounts.DVSubAccount, error) {
	var (
		botID             string
		botType           sql.NullInt32
		defaultParamsName sql.NullString
		botStateValue     sql.NullInt32
		_status           sql.NullInt32
	)

	err := row.Scan(
		&botID,
		&botType,
		&defaultParamsName,
		&botStateValue,
		&_status,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing

	return &pbDvSubAccounts.DVSubAccount{
		Recipient: &pbRecipients.Recipient{
			Id: botID,
		},
		SubaccountType: 0,
		Address:        "",
	}, nil
}
func (s *DvSubAccountRepository) ScanFullRow(row pgx.Row) (*pbDvSubAccounts.DVSubAccount, error) {
	var (
		recipientID                       string
		nullableLegalentityID             pgtype.Text
		nullableUserID                    pgtype.Text
		recipientLabel                    string
		recipientType                     pbRecipients.Type
		nullableOrgID                     pgtype.Text
		recipientStatus                   pbCommon.Status
		userLogin                         string
		userType                          pbUsers.Type
		nullableScreenName                pgtype.Text
		nullableAvatar                    pgtype.Text
		userStatus                        pbCommon.Status
		orgName                           string
		orgType                           pbOrgs.Type
		orgStatus                         pbCommon.Status
		legalEntityName                   string
		legalEntityType                   pbLegalEntities.Type
		legalEntityIncorporationCountryID string
		nullableIncorporationLocality     pgtype.Text
		nullableBusinessRegistrationNo    pgtype.Text
		nullableBusinessRegistrationAltNo pgtype.Text
		nullableValidUntil                pgtype.Timestamp
		nullableTaxID                     pgtype.Text
		legalEntityCurrency1ID            string
		nullableCurrency2ID               pgtype.Text
		nullableCurrency3ID               pgtype.Text
		legalEntityStatus                 pbCommon.Status
		botID                             string
		botType                           sql.NullInt32
		defaultParamsName                 sql.NullString
		botStateValue                     sql.NullInt32
		_status                           sql.NullInt32
	)

	err := row.Scan(
		&recipientID,
		&nullableLegalentityID,
		&nullableUserID,
		&recipientLabel,
		&recipientType,
		&nullableOrgID,
		&recipientStatus,
		&userLogin,
		&userType,
		&nullableScreenName,
		&nullableAvatar,
		&userStatus,
		&legalEntityName,
		&legalEntityType,
		&legalEntityIncorporationCountryID,
		&nullableIncorporationLocality,
		&nullableBusinessRegistrationNo,
		&nullableBusinessRegistrationAltNo,
		&nullableValidUntil,
		&nullableTaxID,
		&legalEntityCurrency1ID,
		&nullableCurrency2ID,
		&nullableCurrency3ID,
		&legalEntityStatus,
		&orgName,
		&orgType,
		&orgStatus,
		&botID,
		&botType,
		&defaultParamsName,
		&botStateValue,
		&_status,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing

	return &pbDvSubAccounts.DVSubAccount{
		Recipient: &pbRecipients.Recipient{
			Id:    recipientID,
			Label: recipientLabel,
			LegalEntity: &pbLegalEntities.LegalEntity{
				Id:   nullableLegalentityID.String,
				Name: legalEntityName,
				Type: legalEntityType,
				IncorporationCountry: &pbCountries.Country{
					Id: legalEntityIncorporationCountryID,
				},
				IncorporationLocality:     &nullableIncorporationLocality.String,
				BusinessRegistrationNo:    &nullableBusinessRegistrationNo.String,
				BusinessRegistrationAltNo: &nullableBusinessRegistrationAltNo.String,
				ValidUntil:                timestamppb.New(nullableValidUntil.Time),
				TaxId:                     &nullableTaxID.String,
				Currency1: &pbUoms.UoM{
					Id: legalEntityCurrency1ID,
				},
				Currency2: &pbUoms.UoM{
					Id: nullableCurrency2ID.String,
				},
				Currency3: &pbUoms.UoM{
					Id: nullableCurrency3ID.String,
				},
				Status: legalEntityStatus,
			},
			User: &pbUsers.User{
				Id:         nullableUserID.String,
				Login:      userLogin,
				Type:       userType,
				ScreenName: &nullableScreenName.String,
				Avatar:     &nullableAvatar.String,
				Status:     userStatus,
			},
			Org: &pbOrgs.Org{
				Id:     nullableOrgID.String,
				Name:   orgName,
				Type:   orgType,
				Status: orgStatus,
			},
			Type:   recipientType,
			Status: recipientStatus,
		},
		SubaccountType: 0,
		Address:        "",
	}, nil
}
func (s *DvSubAccountRepository) ScanGetRow(row pgx.Row) (*pbDvSubAccounts.DVSubAccount, error) {
	var (
		recipientID           string
		nullableLegalentityID pgtype.Text
		nullableUserID        pgtype.Text
		recipientLabel        string
		recipientType         pbRecipients.Type
		nullableOrgID         pgtype.Text
		recipientStatus       pbCommon.Status
		botID                 string
		botType               sql.NullInt32
		defaultParamsName     sql.NullString
		botStateValue         sql.NullInt32
		_status               sql.NullInt32
	)

	err := row.Scan(
		&recipientID,
		&nullableLegalentityID,
		&nullableUserID,
		&recipientLabel,
		&recipientType,
		&nullableOrgID,
		&recipientStatus,
		&botID,
		&botType,
		&defaultParamsName,
		&botStateValue,
		&_status,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing

	return &pbDvSubAccounts.DVSubAccount{
		Recipient: &pbRecipients.Recipient{
			Id:    recipientID,
			Label: recipientLabel,
			LegalEntity: &pbLegalEntities.LegalEntity{
				Id: nullableLegalentityID.String,
			},
			User: &pbUsers.User{
				Id: nullableUserID.String,
			},
			Org: &pbOrgs.Org{
				Id: nullableOrgID.String,
			},
			Type:   recipientType,
			Status: recipientStatus,
		},
		SubaccountType: 0,
		Address:        "",
	}, nil
}

func (s *DvSubAccountRepository) ScanListRow(row pgx.Row) (*pbDvSubAccounts.DVSubAccount, error) {
	var (
		recipientID           sql.NullString
		nullableLegalentityID sql.NullString
		nullableUserID        sql.NullString
		recipientLabel        sql.NullString
		recipientType         pbRecipients.Type
		nullableOrgID         sql.NullString
		recipientStatus       pbCommon.Status
		subAccountType        sql.NullInt32
		address               sql.NullString
	)

	err := row.Scan(
		&recipientID,
		&nullableLegalentityID,
		&nullableUserID,
		&recipientLabel,
		&recipientType,
		&nullableOrgID,
		&recipientStatus,
		&subAccountType,
		address,
	)
	if err != nil {
		return nil, err
	}

	// Nullable field processing
	if !subAccountType.Valid {
		subAccountType = sql.NullInt32{
			Int32: 0,
			Valid: true,
		}
	}

	return &pbDvSubAccounts.DVSubAccount{
		Recipient: &pbRecipients.Recipient{
			Id:    getNullableString(recipientID),
			Label: getNullableString(recipientLabel),
			LegalEntity: &pbLegalEntities.LegalEntity{
				Id: getNullableString(nullableLegalentityID),
			},
			User: &pbUsers.User{
				Id: getNullableString(nullableUserID),
			},
			Org: &pbOrgs.Org{
				Id: getNullableString(nullableOrgID),
			},
			Type:   recipientType,
			Status: recipientStatus,
		},
		SubaccountType: pbDvSubAccounts.Type(*util.GetSQLNullInt32(subAccountType)),
		Address:        getNullableString(address),
	}, nil
}

func (s *DvSubAccountRepository) ScanUpdateRow(row pgx.Row) (*pbDvSubAccounts.DVSubAccount, error) {
	var (
		botID             sql.NullString
		botType           sql.NullInt32
		defaultParamsName sql.NullString
		botStateValue     sql.NullInt32
		_status           sql.NullInt32
	)

	err := row.Scan(
		&botID,
		&botType,
		&defaultParamsName,
		&botStateValue,
		&_status,
	)
	if err != nil {
		return nil, err
	}
	// Nullable field processing
	if !botType.Valid {
		botType = sql.NullInt32{
			Int32: 0,
			Valid: true,
		}
	}
	if !botStateValue.Valid {
		botStateValue = sql.NullInt32{
			Int32: 1,
			Valid: true,
		}
	}

	return &pbDvSubAccounts.DVSubAccount{
		Recipient: &pbRecipients.Recipient{
			Id: getNullableString(botID),
		},
		SubaccountType: 0,
		Address:        "",
	}, nil
}

func getNullableString(s sql.NullString) string {
	if s.Valid {
		return *util.GetSQLNullString(s)
	}
	return *util.GetSQLNullString(sql.NullString{
		String: "",
		Valid:  true,
	})
}
