package contacts

import (
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	pbContactsConnect "davensi.com/core/gen/contacts/contactsconnect"
	pbKyc "davensi.com/core/gen/kyc"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_fields    = "id, type, value, status"
	_tableName = "core.contacts"
)

type ContactRepository struct {
	pbContactsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewContactRepository(db *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{
		db: db,
	}
}

func (s *ContactRepository) QbInsert(msg *pbContacts.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleSocialValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("type").SetInsertField("value")
	singleSocialValue = append(singleSocialValue, msg.GetType(), msg.GetValue())

	// Append optional fields values
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleSocialValue = append(singleSocialValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleSocialValue)

	return qb, err
}

func (s *ContactRepository) QbBulkInsert(msg *pbContacts.SetLabeledContactList) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	qb.SetInsertField(
		"type", "value", "status",
	)
	for _, labelContact := range msg.GetList() {
		if labelContact.GetContact() != nil {
			createReq := labelContact.GetContact()
			_, err := qb.SetInsertValues([]any{
				createReq.GetType(),
				createReq.GetValue(),
				createReq.GetStatus(),
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return qb, nil
}

func (s *ContactRepository) QbUpdate(msg *pbContacts.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Type != nil {
		qb.SetUpdate("type", msg.GetType())
	}

	if msg.Value != nil {
		qb.SetUpdate("value", msg.GetValue())
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

func (s *ContactRepository) QbDeleteMany(ids []string) (qb *util.QueryBuilder, err error) {
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

func (s *ContactRepository) QbGetOne(msg *pbContacts.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	qb.Where("id = ?", msg.GetId())
	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *ContactRepository) QbGetList(msg *pbContacts.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, "contacts"))

	if msg.Type != nil {
		contactTypes := msg.GetType().GetList()

		if len(contactTypes) > 0 {
			args := []any{}
			for _, v := range contactTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"contacts.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(contactTypes)), ""), ", "),
				),
				args...,
			)
		}
	}
	if msg.Value != nil {
		qb.Where("contacts.value LIKE '%' || ? || '%'", msg.GetValue())
	}
	if msg.Status != nil {
		socialStatuses := msg.GetStatus().GetList()

		if len(socialStatuses) > 0 {
			args := []any{}
			for _, v := range socialStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"contacts.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(socialStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *ContactRepository) QbBulkInsertMany(
	valueSlices []*pbContacts.CreateRequest, controlContractIDs []string) (*util.QueryBuilder, error) {
	var qbErr error
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	if controlContractIDs != nil {
		qb.SetInsertField("id", "type", "value", "status")
	} else {
		qb.SetInsertField("type", "value", "status")
	}
	for index, value := range valueSlices {
		if controlContractIDs != nil && len(valueSlices) != len(controlContractIDs) {
			return nil, errors.New("length of valueSlices must be equal to length of controlContractIDs")
		}
		if controlContractIDs != nil {
			_, qbErr = qb.SetInsertValues([]any{
				controlContractIDs[index],
				value.GetType(),
				value.GetValue(),
				value.GetStatus().Enum(),
			})
		} else {
			_, qbErr = qb.SetInsertValues([]any{
				value.GetType(),
				value.GetValue(),
				value.GetStatus().Enum(),
			})
		}

		if qbErr != nil {
			return nil, qbErr
		}
	}
	return qb, nil
}

func (s *ContactRepository) ScanRow(row pgx.Row) (*pbContacts.Contact, error) {
	var (
		id          string
		contactType pbContacts.Type
		value       string
		status      pbCommon.Status
	)

	err := row.Scan(
		&id,
		&contactType,
		&value,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbContacts.Contact{
		Id:     id,
		Type:   contactType,
		Value:  value,
		Status: status,
	}, nil
}

func (s *ContactRepository) QbInsertWithUUID(msg *pbContacts.CreateRequest, contactUUID string) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleContactValue := []any{}

	// Append required value: type, value
	qb.SetInsertField("type").SetInsertField("value")
	singleContactValue = append(singleContactValue, msg.GetType(), msg.GetValue())

	qb.SetInsertField("id")
	singleContactValue = append(singleContactValue, contactUUID)

	// Append optional fields values
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleContactValue = append(singleContactValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleContactValue)

	return qb, err
}

func (s *ContactRepository) ScanMultiRows(rows pgx.Rows) ([]*pbContacts.Contact, error) {
	var (
		id          string
		contactType pbContacts.Type
		value       string
		status      pbCommon.Status
	)

	contactsList := []*pbContacts.Contact{}
	rowValues, rowValuesErr := rows.Values()
	if rowValuesErr != nil {
		return nil, rowValuesErr
	}
	for i := 0; i < len(rowValues); i++ {
		err := rows.Scan(
			&id,
			&contactType,
			&value,
			&status,
		)
		if err != nil {
			return nil, err
		}
		contactsList = append(contactsList, &pbContacts.Contact{
			Id:     id,
			Type:   contactType,
			Value:  value,
			Status: status,
		})
	}
	return contactsList, nil
}
