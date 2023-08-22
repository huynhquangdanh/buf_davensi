package userids

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	pbContacts "davensi.com/core/gen/contacts"
	pbCredentials "davensi.com/core/gen/credentials"
	pbIncomes "davensi.com/core/gen/incomes"
	pbKyc "davensi.com/core/gen/kyc"
	pbLiveliness "davensi.com/core/gen/liveliness"
	pbPhysiques "davensi.com/core/gen/physiques"
	pbSocials "davensi.com/core/gen/socials"
	pbUserIDs "davensi.com/core/gen/userids"
	pbUserIDsConnect "davensi.com/core/gen/userids/useridsconnect"
	pbUsers "davensi.com/core/gen/users"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type UserIDsRepository struct {
	pbUserIDsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}
type SupportedValueSlices struct {
	UserContactValueList []*pbContacts.SetLabeledContact
	UserIncomeValueList  []*pbIncomes.SetLabeledIncome
}

func NewUserIDRepository(db *pgxpool.Pool) *UserIDsRepository {
	return &UserIDsRepository{
		db: db,
	}
}

func (s *UserIDsRepository) QbInsert(
	userID string, status pbKyc.Status, tableName string, optionalCol ...any) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, tableName)
	singleUserIDValue := []any{}
	log.Info().Msgf("inside QbInsert of userid")
	for _, col := range optionalCol {
		switch convertedMsg := col.(type) {
		case *pbCredentials.Credentials:
			qb.SetInsertField("credential_id")
			singleUserIDValue = append(singleUserIDValue, convertedMsg.Id)

		case *pbLiveliness.Liveliness:
			qb.SetInsertField("liveliness_id")
			singleUserIDValue = append(singleUserIDValue, convertedMsg.Id)

		case *pbSocials.Social:
			qb.SetInsertField("social_id")
			singleUserIDValue = append(singleUserIDValue, convertedMsg.Id)

		case *pbPhysiques.Physique:
			qb.SetInsertField("physique_id")
			singleUserIDValue = append(singleUserIDValue, convertedMsg.Id)
		}
	}
	// Append required value: user_id, key
	qb.SetInsertField("user_id")
	singleUserIDValue = append(singleUserIDValue, userID)

	// Append optional fields values
	qb.SetInsertField("status")
	singleUserIDValue = append(singleUserIDValue, status)

	_, err := qb.SetInsertValues(singleUserIDValue)

	return qb, err
}

func (s *UserIDsRepository) QbBulkInsert(tableName, fields, userID string, valueSlices SupportedValueSlices) (*util.QueryBuilder, error) {
	// Split the string using the comma as a delimiter
	fieldNames := strings.Split(fields, ",")

	// Trim any leading or trailing whitespaces from the field names
	for i, fieldName := range fieldNames {
		fieldNames[i] = strings.TrimSpace(fieldName)
	}
	qb := util.CreateQueryBuilder(util.Insert, tableName)
	qb.SetInsertField(fieldNames...)
	if !reflect.ValueOf(valueSlices.UserContactValueList).IsNil() {
		for _, value := range valueSlices.UserContactValueList {
			_, err := qb.SetInsertValues([]any{
				userID,
				value.GetLabel(),
				value.GetId(),
				value.GetMainContact(),
				value.GetStatus().Enum(),
			})
			if err != nil {
				return nil, err
			}
		}
		return qb, nil
	}
	if !reflect.ValueOf(valueSlices.UserIncomeValueList).IsNil() {
		for _, value := range valueSlices.UserIncomeValueList {
			_, err := qb.SetInsertValues([]any{
				userID,
				value.GetLabel(),
				value.GetId(),
				value.GetStatus().Enum(),
			})
			if err != nil {
				return nil, err
			}
		}
		return qb, nil
	}
	return nil, errors.New("value slices is not supported for bulk insert")
}

func (s *UserIDsRepository) QbUserAdditionInfoUpdateMany(
	tableName, userID string,
	valueSlices SupportedValueSlices,
) ([]*util.QueryBuilder, error) {
	var qbList []*util.QueryBuilder
	if !reflect.ValueOf(valueSlices.UserContactValueList).IsNil() {
		for _, value := range valueSlices.UserContactValueList {
			qb := util.CreateQueryBuilder(util.Update, tableName)
			if value.Label != "" {
				qb.SetUpdate("label", value.Label)
			}
			if value.GetStatus() != pbKyc.Status_STATUS_UNSPECIFIED {
				qb.SetUpdate("status", value.GetStatus())
			}
			if !reflect.ValueOf(value.MainContact).IsNil() {
				qb.SetUpdate("main_contact", value.MainContact)
			}
			if qb.IsUpdatable() {
				qb.Where("user_id = ? AND contact_id = ? AND status != ?", userID, value.GetId(), pbKyc.Status_STATUS_CANCELED)
				qbList = append(qbList, qb)
			} else {
				return nil, errors.New("cannot update without new value")
			}
		}
		return qbList, nil
	}
	if !reflect.ValueOf(valueSlices.UserIncomeValueList).IsNil() {
		for _, value := range valueSlices.UserIncomeValueList {
			qb := util.CreateQueryBuilder(util.Update, tableName)
			if value.Label != "" {
				qb.SetUpdate("label", value.Label)
			}
			if value.GetStatus() != pbKyc.Status_STATUS_UNSPECIFIED {
				qb.SetUpdate("status", value.GetStatus())
			}
			if qb.IsUpdatable() {
				qb.Where("user_id = ? AND income_id = ? AND status != ?", userID, value.GetId(), pbKyc.Status_STATUS_CANCELED)
				qbList = append(qbList, qb)
			} else {
				return nil, errors.New("cannot update without new value")
			}
		}
		return qbList, nil
	}
	return nil, errors.New("value slices is not supported for bulk insert")
}

func (s *UserIDsRepository) QbUpdate(
	userID string, status pbKyc.Status, tableName string, optionalCol ...any,
) (qb *util.QueryBuilder, err error) {
	// TODO: all fields are relationship fields, do later

	qb = util.CreateQueryBuilder(util.Update, tableName)
	for _, col := range optionalCol {
		switch convertedMsg := col.(type) {
		case *pbCredentials.Credentials:
			qb.SetUpdate("credential_id", convertedMsg.Id)
		case *pbLiveliness.Liveliness:
			qb.SetUpdate("liveliness_id", convertedMsg.Id)
		case *pbSocials.Social:
			qb.SetUpdate("social_id", convertedMsg.Id)
		case *pbPhysiques.Physique:
			qb.SetUpdate("physique_id", convertedMsg.Id)
		case *pbContacts.UpdateLabeledContactRequest:
			if convertedMsg.GetStatus() != pbKyc.Status_STATUS_UNSPECIFIED {
				qb.SetUpdate("status", convertedMsg.GetStatus())
			}
			if !reflect.ValueOf(convertedMsg.MainContact).IsNil() {
				qb.SetUpdate("main_contact", convertedMsg.GetMainContact())
			}
			if !reflect.ValueOf(convertedMsg.Contact).IsNil() && !reflect.ValueOf(convertedMsg.Contact.Label).IsNil() {
				qb.SetUpdate("label", convertedMsg.GetContact().GetLabel())
			}
			qb.Where("label = ?", convertedMsg.GetByLabel())
		}
	}

	if status != pbKyc.Status_STATUS_UNSPECIFIED {
		qb.SetUpdate("status", status)
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("user_id = ?", userID)

	return qb, nil
}

func (s *UserIDsRepository) QbUpdateMany(
	userID string, labels []string, status *pbKyc.Status, tableName string, optionalCol ...any,
) (qb *util.QueryBuilder, err error) {
	// TODO: all fields are relationship fields, do later

	qb = util.CreateQueryBuilder(util.Update, tableName)
	for _, col := range optionalCol {
		switch convertedMsg := col.(type) {
		case *pbCredentials.Credentials:
			qb.SetUpdate("credential_id", convertedMsg.Id)
		case *pbLiveliness.Liveliness:
			qb.SetUpdate("liveliness_id", convertedMsg.Id)
		case *pbSocials.Social:
			qb.SetUpdate("social_id", convertedMsg.Id)
		case *pbPhysiques.Physique:
			qb.SetUpdate("physique_id", convertedMsg.Id)
		}
	}

	if status != nil {
		qb.SetUpdate("status", status)
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}
	qb.Where("user_id = ?", userID)
	if len(labels) > 0 {
		qb.Where("label IN ?", labels)
	}
	return qb, nil
}

func (s *UserIDsRepository) QbGetOne(
	userID, tableName, field string, labels []string,
) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, tableName)

	qb.Select(field)
	qb.Where("user_id = ?", userID)
	if len(labels) > 0 {
		args := []any{}
		for _, v := range labels {
			args = append(args, v)
		}
		qb.Where(
			fmt.Sprintf(
				"label IN (%s)",
				strings.Join(strings.Split(strings.Repeat("?", len(labels)), ""), ", "),
			),
			args...,
		)
	}

	return qb
}

func (s *UserIDsRepository) QbGetList(_ *pbUserIDs.SetContactsRequest, userID string) *util.QueryBuilder {
	// TODO later, all relationship fields, too complicated
	qb := util.CreateQueryBuilder(util.Select, _table)
	qb.Select(_userIDsFields)

	qb.Where("user_id = ?", userID)

	// if req.Status != nil {
	// 	userIDStatuses := req.GetStatus().GetList()

	// 	if len(userIDStatuses) > 0 {
	// 		args := []any{}
	// 		for _, v := range userIDStatuses {
	// 			args = append(args, v)
	// 		}

	// 		qb.Where(
	// 			fmt.Sprintf(
	// 				"status IN (%s)",
	// 				strings.Join(strings.Split(strings.Repeat("?", len(userIDStatuses)), ""), ", "),
	// 			),
	// 			args...,
	// 		)
	// 	}
	// }

	return qb
}

func (s *UserIDsRepository) QbDelete(
	userID string, tableName string, labels []string,
) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, tableName)

	qb.SetUpdate("status", pbKyc.Status_STATUS_UNSPECIFIED)

	qb.Where("user_id = ?", userID)
	if len(labels) > 0 {
		args := []any{}
		for _, v := range labels {
			args = append(args, v)
		}
		qb.Where(
			fmt.Sprintf(
				"label IN (%s)",
				strings.Join(strings.Split(strings.Repeat("?", len(labels)), ""), ", "),
			),
			args...,
		)
	}

	return qb, nil
}

func (s *UserIDsRepository) ScanID(row pgx.Row) (string, error) {
	var userID string
	if err := row.Scan(&userID); err != nil {
		return "", err
	}
	return userID, nil
}

func (s *UserIDsRepository) ScanRow(row pgx.Row) (*pbUserIDs.UserId, error) {
	var (
		userID       string
		credentialID pgtype.Text
		physiqueID   pgtype.Text
		livelinessID pgtype.Text
		socialID     pgtype.Text
		status       uint32
	)

	err := row.Scan(
		&userID,
		&credentialID,
		&physiqueID,
		&livelinessID,
		&socialID,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbUserIDs.UserId{
		User: &pbUsers.User{
			Id: userID,
		},
		Credentials: &pbCredentials.Credentials{
			Id: credentialID.String,
		},
		Physique: &pbPhysiques.Physique{
			Id: physiqueID.String,
		},
		Liveliness: &pbLiveliness.Liveliness{
			Id: livelinessID.String,
		},
		Social: &pbSocials.Social{
			Id: socialID.String,
		},
		Status: pbKyc.Status(status),
	}, nil
}
func (s *UserIDsRepository) ScanMultiLabelRows(rows pgx.Rows) (*pbIncomes.LabeledIncome, error) {
	var (
		userID   string
		label    pgtype.Text
		incomeID pgtype.Text
		status   uint32
	)
	err := rows.Scan(
		&userID,
		&label,
		&incomeID,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbIncomes.LabeledIncome{
		Label: label.String,
		Income: &pbIncomes.Income{
			Id: incomeID.String,
		},
		Status: pbKyc.Status(status),
	}, nil
}
