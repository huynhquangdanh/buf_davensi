package usercontacts

import (
	"reflect"
	"strings"
	"sync"

	"errors"

	pbContacts "davensi.com/core/gen/contacts"
	pbKyc "davensi.com/core/gen/kyc"
	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName          = "core.users_contacts"
	_userContactsFields = "user_id, label, contact_id, status"
)

type UserContactsRepository struct {
	db *pgxpool.Pool
}

var (
	singleRepo *UserContactsRepository
	once       sync.Once
)

func NewUserContactRepository(db *pgxpool.Pool) *UserContactsRepository {
	return &UserContactsRepository{
		db: db,
	}
}

func GetSingletonRepository(db *pgxpool.Pool) *UserContactsRepository {
	once.Do(func() {
		singleRepo = NewUserContactRepository(db)
	})
	return singleRepo
}

func (*UserContactsRepository) QbUpsertUserContacts(
	userID string,
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
	qb.SetInsertField("user_id", "label", "contact_id", "status")

	for _, contact := range contacts {
		_, err := qb.SetInsertValues([]any{
			userID,
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
func (*UserContactsRepository) QbBulkInsert(userID string, valueSlices []*pbContacts.SetLabeledContact) (*util.QueryBuilder, error) {
	// Split the string using the comma as a delimiter
	fieldNames := strings.Split(_userContactFields, ",")

	// Trim any leading or trailing whitespaces from the field names
	for i, fieldName := range fieldNames {
		fieldNames[i] = strings.TrimSpace(fieldName)
	}
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	qb.SetInsertField(fieldNames...)
	if !reflect.ValueOf(valueSlices).IsNil() {
		for _, value := range valueSlices {
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
	return nil, errors.New("value slices is not supported for bulk insert")
}

func (*UserContactsRepository) ScanRow(rows pgx.Rows) (*pbContacts.LabeledContact, error) {
	var (
		userID      string
		contactID   pgtype.Text
		label       string
		mainContact bool
		status      uint32
	)

	err := rows.Scan(
		&userID,
		&label,
		&contactID,
		&mainContact,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbContacts.LabeledContact{
		Label: label,
		Contact: &pbContacts.Contact{
			Id: contactID.String,
		},
		Status: pbKyc.Status(status),
	}, nil
}

func (*UserContactsRepository) ScanMultiRows(rows pgx.Rows) (*pbContacts.LabeledContactList, error) {
	var (
		userID      string
		contactID   pgtype.Text
		label       string
		mainContact bool
		status      uint32
	)

	userContacts := []*pbContacts.LabeledContact{}

	for rows.Next() {
		err := rows.Scan(
			&userID,
			&label,
			&contactID,
			&mainContact,
			&status,
		)
		if err != nil {
			return nil, err
		}
		userContacts = append(userContacts, &pbContacts.LabeledContact{
			Label: label,
			Contact: &pbContacts.Contact{
				Id: contactID.String,
			},
			MainContact: &mainContact,
			Status:      pbKyc.Status(status),
		})
	}
	return &pbContacts.LabeledContactList{
		List: userContacts,
	}, nil
}

func (*UserContactsRepository) ScanLabeledContactRows(rows pgx.Rows) ([]*pbContacts.LabeledContact, error) {
	var (
		userID      string
		contactID   pgtype.Text
		label       string
		mainContact bool
		status      uint32
	)

	userContacts := []*pbContacts.LabeledContact{}

	for rows.Next() {
		err := rows.Scan(
			&userID,
			&label,
			&contactID,
			&mainContact,
			&status,
		)
		if err != nil {
			return nil, err
		}
		userContacts = append(userContacts, &pbContacts.LabeledContact{
			Label: label,
			Contact: &pbContacts.Contact{
				Id: contactID.String,
			},
			MainContact: &mainContact,
			Status:      pbKyc.Status(status),
		})
	}
	return userContacts, nil
}
func (*UserContactsRepository) ScanLabelContactSingleRow(row pgx.Row) (*pbContacts.LabeledContact, error) {
	var (
		userID      string
		contactID   pgtype.Text
		label       string
		mainContact bool
		status      uint32
	)

	err := row.Scan(
		&userID,
		&label,
		&contactID,
		&mainContact,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbContacts.LabeledContact{
		Label: label,
		Contact: &pbContacts.Contact{
			Id: contactID.String,
		},
		MainContact: &mainContact,
		Status:      pbKyc.Status(status),
	}, nil
}
