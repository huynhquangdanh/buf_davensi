package recipients

import (
	"errors"
	"fmt"
	"strings"

	"davensi.com/core/internal/legalentities"
	"davensi.com/core/internal/users"
	"davensi.com/core/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbOrgs "davensi.com/core/gen/orgs"
	pbRecipients "davensi.com/core/gen/recipients"
	pbRecipientsConnect "davensi.com/core/gen/recipients/recipientsconnect"
	pbUoms "davensi.com/core/gen/uoms"
	pbUsers "davensi.com/core/gen/users"
)

const (
	_tableName        = "core.recipients"
	_recipientsFields = "id, legalentity_id, user_id, label, type, org_id, status"
)

type Repository struct {
	pbRecipientsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewRecipientsRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (s *Repository) QbInsert(
	req *pbRecipients.CreateRequest,
	pkRes RecipientRelationshipIds,
) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	var singleRecipientValue []any

	qb.SetInsertField("label")
	singleRecipientValue = append(singleRecipientValue, req.GetLabel())

	// Append optional fields values
	if pkRes.UserID != nil {
		qb.SetInsertField("user_id")
		singleRecipientValue = append(singleRecipientValue, pkRes.UserID.String())
	}
	if pkRes.LegalEntityID != nil {
		qb.SetInsertField("legalentity_id")
		singleRecipientValue = append(singleRecipientValue, pkRes.LegalEntityID.String())
	}
	if pkRes.OrgID != nil {
		qb.SetInsertField("org_id")
		singleRecipientValue = append(singleRecipientValue, pkRes.OrgID.String())
	}
	if req.Type != nil {
		qb.SetInsertField("type")
		singleRecipientValue = append(singleRecipientValue, req.GetType())
	} else {
		qb.SetInsertField("type")
		singleRecipientValue = append(singleRecipientValue, pbRecipients.Type_TYPE_UNSPECIFIED)
	}
	if req.Status != nil {
		qb.SetInsertField("status")
		singleRecipientValue = append(singleRecipientValue, req.GetStatus())
	} else {
		qb.SetInsertField("status")
		singleRecipientValue = append(singleRecipientValue, pbCommon.Status_STATUS_UNSPECIFIED)
	}

	_, err := qb.SetInsertValues(singleRecipientValue)

	return qb, err
}

func (s *Repository) QbUpdate(
	req *pbRecipients.UpdateRequest,
	pkResOld RecipientRelationshipIds,
	pkResNew RecipientRelationshipIds,
) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if req.Label != nil {
		qb.SetUpdate("label", req.GetLabel())
	}

	if req.User != nil {
		if pkResNew.UserID == nil {
			var nilUUID uuid.NullUUID
			qb.SetUpdate("user_id", nilUUID)
		} else {
			qb.SetUpdate("user_id", pkResNew.UserID)
		}
	}

	if req.LegalEntity != nil {
		if pkResNew.LegalEntityID == nil {
			var nilUUID uuid.NullUUID
			qb.SetUpdate("legalentity_id", nilUUID)
		} else {
			qb.SetUpdate("legalentity_id", pkResNew.LegalEntityID)
		}
	}

	if req.Org != nil {
		if pkResNew.OrgID == nil {
			var nilUUID uuid.NullUUID
			qb.SetUpdate("org_id", nilUUID)
		} else {
			qb.SetUpdate("org_id", pkResNew.OrgID)
		}
	}

	if req.Type != nil {
		qb.SetUpdate("type", req.GetType())
	}

	if req.Status != nil {
		qb.SetUpdate("status", req.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch req.GetSelect().GetSelect().(type) {
	case *pbRecipients.Select_ById:
		qb.Where("id = ?", req.GetSelect().GetById())
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		qb.Where(
			"label = ?",
			req.GetSelect().GetByLegalEntityUserLabel().GetLabel(),
		)
		if pkResOld.UserID != nil {
			qb.Where(
				"user_id = ?",
				pkResOld.UserID.String(),
			)
		}
		if pkResOld.LegalEntityID != nil {
			qb.Where(
				"legalentity_id = ?",
				pkResOld.LegalEntityID.String(),
			)
		}
	}

	return qb, nil
}

func (s *Repository) QbGetOne(
	req *pbRecipients.GetRequest,
	pkRes RecipientRelationshipIds,
	mainQuery bool,
	addTableNameToSelect bool,
) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	var selectFields string
	if addTableNameToSelect {
		fields := strings.Split(_recipientsFields, ", ")
		for i := 0; i < len(fields); i++ {
			fields[i] = _tableName + "." + fields[i]
		}
		selectFields = strings.Join(fields, ",")
	} else {
		selectFields = _recipientsFields
	}
	if mainQuery {
		qb.SuperSelect(selectFields)
		fieldUser := "login, type, screen_name, avatar, status"
		fieldLegalEntity := "name, type, incorporation_country_id, incorporation_locality," +
			"business_registration_no, business_registration_alt_no, valid_until," +
			"tax_id, currency1_id, currency2_id, currency3_id, status"
		fieldOrg := "name, type, status"
		qb.SuperJoin("LEFT JOIN core.users ON core.users.id = core.recipients.user_id", fieldUser, "core.users")
		qb.SuperJoin("LEFT JOIN core.legalentities ON core.legalentities.id = core.recipients.legalentity_id",
			fieldLegalEntity, "core.legalentities")
		qb.SuperJoin("LEFT JOIN core.orgs ON core.orgs.id = core.recipients.org_id", fieldOrg, "core.orgs")
	} else {
		qb.Select(selectFields)
	}

	switch req.GetSelect().GetSelect().(type) {
	case *pbRecipients.Select_ById:
		qb.Where(_tableName+".id = ? AND "+_tableName+".status = 1", req.GetSelect().GetById())
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		qb.Where(
			_tableName+".label = ? AND "+_tableName+".status = 1",
			req.GetSelect().GetByLegalEntityUserLabel().GetLabel(),
		)
		if req.GetSelect().GetByLegalEntityUserLabel().GetUser() != nil {
			inputUserID := req.GetSelect().GetByLegalEntityUserLabel().GetUser().GetById()
			if pkRes.UserID != nil {
				inputUserID = pkRes.UserID.String()
			}
			qb.Where(
				_tableName+".user_id = ?",
				inputUserID,
			)
		}
		if req.GetSelect().GetByLegalEntityUserLabel().GetLegalEntity() != nil {
			inputLegalEntityID := req.GetSelect().GetByLegalEntityUserLabel().GetLegalEntity().GetById()
			if pkRes.LegalEntityID != nil {
				inputLegalEntityID = pkRes.LegalEntityID.String()
			}
			qb.Where(
				_tableName+".legalentity_id = ?",
				inputLegalEntityID,
			)
		}
	}
	return qb
}

func (s *Repository) QbGetList(req *pbRecipients.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	fields := strings.Split(_recipientsFields, ", ")
	for i := 0; i < len(fields); i++ {
		fields[i] = _tableName + "." + fields[i]
	}
	newFields := strings.Join(fields, ",")
	qb.Select(newFields)

	if req.Label != nil {
		qb.Where("recipients.label LIKE '%' || ? || '%'", req.GetLabel())
	}

	if req.Type != nil {
		recipientTypes := req.GetType().GetList()

		if len(recipientTypes) > 0 {
			var args []any
			for _, v := range recipientTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"recipients.type = ANY(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(recipientTypes)), ""), ", "),
				),
				args...,
			)
		}
	}

	if req.Status != nil {
		recipientStatuses := req.GetStatus().GetList()

		if len(recipientStatuses) == 1 {
			qb.Where(fmt.Sprintf("recipients.status = %d", recipientStatuses[0]))
		}
		if len(recipientStatuses) > 1 {
			var args []any
			for _, v := range recipientStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"recipients.status = ANY(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(recipientStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *Repository) QbDelete(
	req *pbRecipients.DeleteRequest,
	pkRes RecipientRelationshipIds,
) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Delete, _tableName)
	qb.Select(_recipientsFields)

	switch req.GetSelect().GetSelect().(type) {
	case *pbRecipients.Select_ById:
		qb.Where("id = ? AND status = 1", req.GetSelect().GetById())
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		qb.Where(
			"label = ? AND status = 1",
			req.GetSelect().GetByLegalEntityUserLabel().GetLabel(),
		)
		if pkRes.UserID != nil {
			qb.Where(
				_tableName+".user_id = ?",
				pkRes.UserID.String(),
			)
		}
		if pkRes.LegalEntityID != nil {
			qb.Where(
				_tableName+".legalentity_id = ?",
				pkRes.LegalEntityID.String(),
			)
		}
	}

	return qb
}

func (s *Repository) ScanRow(row pgx.Row) (*pbRecipients.Recipient, error) {
	var (
		id                    string
		nullableLegalentityID pgtype.Text
		nullableUserID        pgtype.Text
		label                 string
		recipientType         pbRecipients.Type
		nullableOrgID         pgtype.Text
		status                pbCommon.Status
	)

	err := row.Scan(
		&id,
		&nullableLegalentityID,
		&nullableUserID,
		&label,
		&recipientType,
		&nullableOrgID,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbRecipients.Recipient{
		Id:          id,
		Label:       label,
		LegalEntity: &pbLegalEntities.LegalEntity{Id: nullableLegalentityID.String},
		User:        &pbUsers.User{Id: nullableUserID.String},
		Org:         &pbOrgs.Org{Id: nullableOrgID.String},
		Type:        recipientType,
		Status:      status,
	}, nil
}

func (s *Repository) SuperScanRow(row pgx.Row) (*pbRecipients.Recipient, error) {
	var (
		recipientID                       string
		nullableLegalentityID             pgtype.Text
		nullableUserID                    pgtype.Text
		recipientLabel                    string
		recipientType                     pbRecipients.Type
		nullableOrgID                     pgtype.Text
		recipientStatus                   pbCommon.Status
		userLogin                         pgtype.Text
		userType                          pgtype.Int2
		nullableScreenName                pgtype.Text
		nullableAvatar                    pgtype.Text
		userStatus                        pgtype.Int2
		legalEntityName                   pgtype.Text
		legalEntityType                   pgtype.Int2
		legalEntityIncorporationCountryID pgtype.Text
		nullableIncorporationLocality     pgtype.Text
		nullableBusinessRegistrationNo    pgtype.Text
		nullableBusinessRegistrationAltNo pgtype.Text
		nullableValidUntil                pgtype.Timestamp
		nullableTaxID                     pgtype.Text
		legalEntityCurrency1ID            pgtype.Text
		nullableCurrency2ID               pgtype.Text
		nullableCurrency3ID               pgtype.Text
		legalEntityStatus                 pgtype.Int2
		orgName                           pgtype.Text
		orgType                           pgtype.Int2
		orgStatus                         pgtype.Int2
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
	)
	if err != nil {
		return nil, err
	}

	return &pbRecipients.Recipient{
		Id:    recipientID,
		Label: recipientLabel,
		LegalEntity: &pbLegalEntities.LegalEntity{
			Id:   nullableLegalentityID.String,
			Name: legalEntityName.String,
			Type: pbLegalEntities.Type(legalEntityType.Int16),
			IncorporationCountry: &pbCountries.Country{
				Id: legalEntityIncorporationCountryID.String,
			},
			IncorporationLocality:     &nullableIncorporationLocality.String,
			BusinessRegistrationNo:    &nullableBusinessRegistrationNo.String,
			BusinessRegistrationAltNo: &nullableBusinessRegistrationAltNo.String,
			ValidUntil:                timestamppb.New(nullableValidUntil.Time),
			TaxId:                     &nullableTaxID.String,
			Currency1: &pbUoms.UoM{
				Id: legalEntityCurrency1ID.String,
			},
			Currency2: &pbUoms.UoM{
				Id: nullableCurrency2ID.String,
			},
			Currency3: &pbUoms.UoM{
				Id: nullableCurrency3ID.String,
			},
			Status: pbCommon.Status(legalEntityStatus.Int16),
		},
		User: &pbUsers.User{
			Id:         nullableUserID.String,
			Login:      userLogin.String,
			Type:       pbUsers.Type(userType.Int16),
			ScreenName: &nullableScreenName.String,
			Avatar:     &nullableAvatar.String,
			Status:     pbCommon.Status(userStatus.Int16),
		},
		Org: &pbOrgs.Org{
			Id:     nullableOrgID.String,
			Name:   orgName.String,
			Type:   pbOrgs.Type(orgType.Int16),
			Status: pbCommon.Status(orgStatus.Int16),
		},
		Type:   recipientType,
		Status: recipientStatus,
	}, nil
}

func SetQBBySelect(selectRecipient *pbRecipients.Select, qb *util.QueryBuilder) {
	switch selectRecipient.GetSelect().(type) {
	case *pbRecipients.Select_ById:
		qb.Where("recipients.id = ?", selectRecipient.GetById())
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		filterArgs := []any{}

		qbGetLegalEntity := util.CreateQueryBuilder(util.Select, "core.legalentities")
		legalentities.SetQBBySelect(
			selectRecipient.GetByLegalEntityUserLabel().GetLegalEntity(),
			qbGetLegalEntity.Select("id"),
		)
		legalEntitySQL, legalEntityArgs, _ := qbGetLegalEntity.GenerateSQL()

		qbGetUser := util.CreateQueryBuilder(util.Select, "core.users")
		users.SetQBBySelect(
			selectRecipient.GetByLegalEntityUserLabel().GetUser(),
			qbGetUser.Select("id"),
		)
		userSQL, userArgs, _ := qbGetUser.GenerateSQL()

		filterArgs = append(filterArgs, legalEntityArgs...)
		filterArgs = append(filterArgs, userArgs...)
		filterArgs = append(filterArgs, selectRecipient.GetByLegalEntityUserLabel().GetLabel())

		qb.Where(fmt.Sprintf(
			"recipients.legalentity_id IN (%s) AND recipients.user_id IN (%s) AND recipients.label = ?", legalEntitySQL, userSQL,
		), filterArgs...)
	}
}

func (s *Repository) QbInsertWithUUID(
	req *pbRecipients.CreateRequest,
	pkRes RecipientRelationshipIds,
	recipientUUID string,
) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	var singleRecipientValue []any

	qb.SetInsertField("id")
	singleRecipientValue = append(singleRecipientValue, recipientUUID)

	qb.SetInsertField("label")
	singleRecipientValue = append(singleRecipientValue, req.GetLabel())

	// Append optional fields values
	if pkRes.UserID != nil {
		qb.SetInsertField("user_id")
		singleRecipientValue = append(singleRecipientValue, pkRes.UserID.String())
	}
	if pkRes.LegalEntityID != nil {
		qb.SetInsertField("legalentity_id")
		singleRecipientValue = append(singleRecipientValue, pkRes.LegalEntityID.String())
	}
	if pkRes.OrgID != nil {
		qb.SetInsertField("org_id")
		singleRecipientValue = append(singleRecipientValue, pkRes.OrgID.String())
	}
	if req.Type != nil {
		qb.SetInsertField("type")
		singleRecipientValue = append(singleRecipientValue, req.GetType())
	} else {
		qb.SetInsertField("type")
		singleRecipientValue = append(singleRecipientValue, pbRecipients.Type_TYPE_UNSPECIFIED)
	}
	if req.Status != nil {
		qb.SetInsertField("status")
		singleRecipientValue = append(singleRecipientValue, req.GetStatus())
	} else {
		qb.SetInsertField("status")
		singleRecipientValue = append(singleRecipientValue, pbCommon.Status_STATUS_UNSPECIFIED)
	}

	_, err := qb.SetInsertValues(singleRecipientValue)

	return qb, err
}

func (s *Repository) ScanWithRelationships(row pgx.Row) (*pbRecipients.Recipient, error) {
	var (
		id            string
		legalEntityID pgtype.Text
		userID        pgtype.Text
		label         pgtype.Text
		recipientType pgtype.Int2
		orgID         pgtype.Text
		status        pgtype.Int2

		legalEntityLegalEntityID             pgtype.Text
		legalEntityName                      pgtype.Text
		legalEntityType                      pgtype.Int2
		legalEntityIncorporationCountryID    pgtype.Text
		legalEntityIncorporationLocality     pgtype.Text
		legalEntityBusinessRegistrationNo    pgtype.Text
		legalEntityBusinessRegistrationAltNo pgtype.Text
		legalEntityValidUntil                pgtype.Timestamp
		legalEntityTaxID                     pgtype.Text
		legalEntityCurrency1ID               pgtype.Text
		legalEntityCurrency2ID               pgtype.Text
		legalEntityCurrency3ID               pgtype.Text
		legalEntityStatus                    pgtype.Int2

		userUserID     pgtype.Text
		userLogin      pgtype.Text
		userType       pgtype.Int2
		userScreenName pgtype.Text
		userAvatar     pgtype.Text
		userStatus     pgtype.Int2

		orgOrgID  pgtype.Text
		orgName   pgtype.Text
		orgType   pgtype.Int2
		orgStatus pgtype.Int2
	)

	err := row.Scan(
		&id,
		&legalEntityID,
		&userID,
		&label,
		&recipientType,
		&orgID,
		&status,
		&legalEntityLegalEntityID,
		&legalEntityName,
		&legalEntityType,
		&legalEntityIncorporationCountryID,
		&legalEntityIncorporationLocality,
		&legalEntityBusinessRegistrationNo,
		&legalEntityBusinessRegistrationAltNo,
		&legalEntityValidUntil,
		&legalEntityTaxID,
		&legalEntityCurrency1ID,
		&legalEntityCurrency2ID,
		&legalEntityCurrency3ID,
		&legalEntityStatus,
		&userUserID,
		&userLogin,
		&userType,
		&userScreenName,
		&userAvatar,
		&userStatus,
		&orgOrgID,
		&orgName,
		&orgType,
		&orgStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbRecipients.Recipient{
		Id: id,
		LegalEntity: &pbLegalEntities.LegalEntity{
			Status: pbCommon.Status(legalEntityStatus.Int16),
			IncorporationCountry: &pbCountries.Country{
				Id: legalEntityIncorporationCountryID.String,
			},
			IncorporationLocality:     &legalEntityIncorporationLocality.String,
			BusinessRegistrationAltNo: &legalEntityBusinessRegistrationAltNo.String,
			Currency2: &pbUoms.UoM{
				Id: legalEntityCurrency2ID.String,
			},
			ValidUntil: timestamppb.New(legalEntityValidUntil.Time),
			TaxId:      &legalEntityTaxID.String,
			Currency1: &pbUoms.UoM{
				Id: legalEntityCurrency1ID.String,
			},
			BusinessRegistrationNo: &legalEntityBusinessRegistrationNo.String,
			Id:                     legalEntityID.String,
			Type:                   pbLegalEntities.Type(legalEntityType.Int16),
			Currency3: &pbUoms.UoM{
				Id: legalEntityCurrency3ID.String,
			},
			Name: legalEntityName.String,
		},
		User: &pbUsers.User{
			Id:         userID.String,
			Login:      userLogin.String,
			Type:       pbUsers.Type(userType.Int16),
			ScreenName: &userScreenName.String,
			Avatar:     &userAvatar.String,
			Status:     pbCommon.Status(userStatus.Int16),
		},
		Label: label.String,
		Org: &pbOrgs.Org{
			Id:     orgID.String,
			Name:   orgName.String,
			Type:   pbOrgs.Type(orgType.Int16),
			Status: pbCommon.Status(orgStatus.Int16),
		},
		Type:   pbRecipients.Type(recipientType.Int16),
		Status: pbCommon.Status(status.Int16),
	}, nil
}
