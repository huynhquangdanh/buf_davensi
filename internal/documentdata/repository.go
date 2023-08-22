package documentdata

import (
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
)

const (
	DDTableName = "core.documents_data"
	DDFields    = "document_id, field, value, status"
)

func QbUpsert(documentID string, documentData map[string]string) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Insert, DDTableName).
		SetInsertField("document_id", "field", "value", "status")

	for field, value := range documentData {
		_, err = qb.SetInsertValues([]any{
			documentID,
			field,
			value,
			pbCommon.Status_STATUS_ACTIVE.Enum(),
		})
		if err != nil {
			break
		}
	}

	if err != nil {
		return qb, err
	}

	return qb.
		OnConflict("document_id", "field").
		SetUpdate("value", nil).
		SetUpdate("status", pbCommon.Status_STATUS_ACTIVE), err
}

func QbGetList() (qb *util.QueryBuilder) {
	qb = util.CreateQueryBuilder(util.Select, DDTableName).
		Select(util.GetFieldsWithTableName(DDFields, DDTableName))

	return qb
}

func QbUpdate(documentID string, documentData map[string]string) (qb *util.QueryBuilder, err error) {
	tempDocData := []string{}

	for field, value := range documentData {
		tempDocData = append(
			tempDocData,
			fmt.Sprintf(
				`{"tempField": %q, "tempValue": %q}`,
				strings.ReplaceAll(field, "'", ""),
				strings.ReplaceAll(value, "'", ""),
			),
		)
	}

	if len(tempDocData) == 0 {
		return qb, errors.New("update value must be specified")
	}

	qb = util.CreateQueryBuilder(util.Update, DDTableName).
		Where("document_id", documentID).
		Where("field = (tempDocData->'tempField')").
		SetUpdateFrom("jsonb_array_elements(?::jsonb) AS tempDocData", tempDocData).
		SetUpdate("value = (tempDocData->'tempValue')", nil)

	return qb, err
}

func QbRemove(documentID string, fields []string) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Update, DDTableName).
		Where("document_id", documentID).
		SetUpdate("status", pbCommon.Status_STATUS_TERMINATED)

	if len(fields) > 0 {
		args := []any{}
		for _, field := range fields {
			args = append(args, field)
		}
		qb.Where(
			fmt.Sprintf("fields in (%s)", strings.Join(strings.Split(strings.Repeat("?", len(fields)), ""), ", ")),
			args...,
		)
	}

	return qb
}

type DocumentData struct {
	Data map[string]string
}

func ScanRow(rows pgx.Rows) (*DocumentData, error) {
	documentData := &DocumentData{
		Data: map[string]string{},
	}

	for rows.Next() {
		var (
			documentID string
			field      string
			value      string
			status     pbCommon.Status
		)
		err := rows.Scan(
			&documentID,
			&field,
			&value,
			&status,
		)
		if err != nil {
			return nil, err
		}
		documentData.Data[field] = value
	}

	return documentData, nil
}
