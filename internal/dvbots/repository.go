package dvbots

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbOrgs "davensi.com/core/gen/orgs"
	pbUoms "davensi.com/core/gen/uoms"
	pbUsers "davensi.com/core/gen/users"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbDvbots "davensi.com/core/gen/dvbots"
	pbDvbotConnect "davensi.com/core/gen/dvbots/dvbotsconnect"
	pbRecipients "davensi.com/core/gen/recipients"
	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DvbotRepository struct {
	pbDvbotConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewDvbotRepository(db *pgxpool.Pool) *DvbotRepository {
	return &DvbotRepository{
		db: db,
	}
}

func (s *DvbotRepository) QbInsert(msg *pbDvbots.CreateRequest, recipient *pbRecipients.Recipient) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _table)
	singleDvbotValue := []any{}
	qb.SetInsertField("id")
	singleDvbotValue = append(singleDvbotValue, recipient.GetId())
	// Append optional fields values
	qb.SetInsertField("bot_type")
	singleDvbotValue = append(singleDvbotValue, msg.GetBotType())

	if msg.DefaultParamsName != nil {
		qb.SetInsertField("default_params_name")
		singleDvbotValue = append(singleDvbotValue, msg.GetDefaultParamsName())
	}

	if msg.BotStatus != nil {
		qb.SetInsertField("bot_state")
		singleDvbotValue = append(singleDvbotValue, msg.GetBotStatus())
	}

	_, err := qb.SetInsertValues(singleDvbotValue)
	return qb, err
}

func (s *DvbotRepository) QbUpdate(msg *pbDvbots.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _table)
	if msg.GetBotType() != pbDvbots.Type_TYPE_UNSPECIFIED {
		qb.SetUpdate("bot_type", msg.GetBotType())
	}
	if msg.GetDefaultParamsName() != "" {
		qb.SetUpdate("default_params_name", msg.GetDefaultParamsName())
	}
	if msg.GetBotStatus() != pbDvbots.BotState_BOT_STATE_UNSPECIFIED {
		qb.SetUpdate("bot_state", msg.GetBotStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}
	switch msg.GetRecipient().GetSelect().GetSelect().(type) {
	case *pbRecipients.Select_ById:
		qb.Where("core.dvbots.id = ? ", msg.GetRecipient().GetSelect().GetById())
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		qb.Where(
			"core.dvbots.id = (SELECT core.recipients.id FROM core.recipients "+
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

func (s *DvbotRepository) QbGetOne(_ *pbRecipients.GetRequest, recipientsQb *util.QueryBuilder) *util.QueryBuilder {
	recipientsQb.Join(fmt.Sprintf("LEFT JOIN %s ON dvbots.id = recipients.id", _table))
	recipientsQb.Where("core.dvbots.bot_state != ?", pbDvbots.BotState_BOT_STATE_STOPPED)
	fields := strings.Split(_fields, ",")
	for i := 0; i < len(fields); i++ {
		fields[i] = _table + "." + fields[i]
	}
	newFields := strings.Join(fields, ",")
	recipientsQb.Select(newFields)
	return recipientsQb
}
func (s *DvbotRepository) QbGetList(msg *pbDvbots.GetListRequest, recipientsQb *util.QueryBuilder) *util.QueryBuilder {
	recipientsQb.Join(fmt.Sprintf("LEFT JOIN %s ON dvbots.id = recipients.id", _table))
	fields := strings.Split(_fields, ",")
	for i := 0; i < len(fields); i++ {
		fields[i] = _table + "." + fields[i]
	}
	newFields := strings.Join(fields, ",")
	recipientsQb.Select(newFields)
	recipientsQb.Where("core.recipients.type = ?", pbRecipients.Type_TYPE_DV_BOT)
	if msg.BotType != nil {
		recipientsQb.Where("core.dvbots.bot_type = ?", msg.GetBotType())
	}
	if msg.DefaultParamsName != nil {
		recipientsQb.Where("core.dvbots.default_params_name = ?", msg.GetDefaultParamsName())
	}
	if msg.BotStatus != nil {
		recipientsQb.Where("core.dvbots.bot_status = ?", msg.GetBotStatus())
	}

	return recipientsQb
}

func (s *DvbotRepository) ScanRow(row pgx.Row) (*pbDvbots.DVBot, error) {
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

	return &pbDvbots.DVBot{
		Recipient: &pbRecipients.Recipient{
			Id: botID,
		},
		BotType:           pbDvbots.Type(*util.GetSQLNullInt32(botType)),
		DefaultParamsName: *util.GetSQLNullString(defaultParamsName),
		BotState:          pbDvbots.BotState(*util.GetSQLNullInt32(botStateValue)),
	}, nil
}
func (s *DvbotRepository) ScanFullRow(row pgx.Row) (*pbDvbots.DVBot, error) {
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

	return &pbDvbots.DVBot{
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
		BotType:           pbDvbots.Type(*util.GetSQLNullInt32(botType)),
		DefaultParamsName: *util.GetSQLNullString(defaultParamsName),
		BotState:          pbDvbots.BotState(*util.GetSQLNullInt32(botStateValue)),
	}, nil
}
func (s *DvbotRepository) ScanGetRow(row pgx.Row) (*pbDvbots.DVBot, error) {
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

	return &pbDvbots.DVBot{
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
		BotType:           pbDvbots.Type(*util.GetSQLNullInt32(botType)),
		DefaultParamsName: *util.GetSQLNullString(defaultParamsName),
		BotState:          pbDvbots.BotState(*util.GetSQLNullInt32(botStateValue)),
	}, nil
}

func (s *DvbotRepository) ScanListRow(row pgx.Row) (*pbDvbots.DVBot, error) {
	var (
		recipientID           sql.NullString
		nullableLegalentityID sql.NullString
		nullableUserID        sql.NullString
		recipientLabel        sql.NullString
		recipientType         pbRecipients.Type
		nullableOrgID         sql.NullString
		recipientStatus       pbCommon.Status
		botID                 sql.NullString
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

	return &pbDvbots.DVBot{
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
		BotType:           pbDvbots.Type(*util.GetSQLNullInt32(botType)),
		DefaultParamsName: getNullableString(defaultParamsName),
		BotState:          pbDvbots.BotState(*util.GetSQLNullInt32(botStateValue)),
	}, nil
}

func (s *DvbotRepository) ScanUpdateRow(row pgx.Row) (*pbDvbots.DVBot, error) {
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

	return &pbDvbots.DVBot{
		Recipient: &pbRecipients.Recipient{
			Id: getNullableString(botID),
		},
		BotType:           pbDvbots.Type(*util.GetSQLNullInt32(botType)),
		DefaultParamsName: getNullableString(defaultParamsName),
		BotState:          pbDvbots.BotState(*util.GetSQLNullInt32(botStateValue)),
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
