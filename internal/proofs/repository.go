package proofs

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbDocuments "davensi.com/core/gen/documents"
	pbKyc "davensi.com/core/gen/kyc"
	pbProofs "davensi.com/core/gen/proofs"
	pbProofsConnect "davensi.com/core/gen/proofs/proofsconnect"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_fields    = "id, section, record_id, document_type, name, document_id, status"
	_tableName = "core.kyc_proofs"
)

type ProofRepository struct {
	pbProofsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewProofRepository(db *pgxpool.Pool) *ProofRepository {
	return &ProofRepository{
		db: db,
	}
}

func (s *ProofRepository) QbInsert(msg *pbProofs.CreateRequest, document *pbDocuments.Document) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleValue := []any{
		msg.GetSection(),
		msg.GetRecordId(),
		msg.GetDocumentType(),
		msg.GetName(),
		msg.GetStatus(),
	}

	// Append required value: type, symbol
	qb.SetInsertField("section", "record_id", "document_type", "name", "status")

	if document != nil {
		qb.SetInsertField("document_id")
		singleValue = append(singleValue, document.GetId())
	}

	_, err := qb.SetInsertValues(singleValue)

	return qb, err
}

func (s *ProofRepository) QbUpdate(msg *pbProofs.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Section != nil {
		qb.SetUpdate("section", msg.GetSection())
	}

	if msg.RecordId != nil {
		qb.SetUpdate("record_id", msg.GetRecordId())
	}

	if msg.DocumentType != nil {
		qb.SetUpdate("document_type", msg.GetDocumentType())
	}

	if msg.Name != nil {
		qb.SetUpdate("name", msg.GetName())
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

func (s *ProofRepository) QbGetOne(msg *pbProofs.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, _tableName))

	qb.Where("id = ?", msg.GetId()).Where("status = ?", pbKyc.Status_STATUS_VALIDATED)

	return qb
}

func (s *ProofRepository) QbGetList(msg *pbProofs.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, _tableName))

	if msg.Section != nil {
		sections := msg.GetSection().GetList()

		if len(sections) > 0 {
			args := []any{}
			for _, v := range sections {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"section IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(sections)), ""), ", "),
				),
				args...,
			)
		}
	}
	if msg.RecordId != nil {
		qb.Where("value LIKE '%' || ? || '%'", msg.GetRecordId())
	}
	if msg.DocumentType != nil {
		qb.Where("value LIKE '%' || ? || '%'", msg.GetDocument())
	}
	if msg.Name != nil {
		qb.Where("value LIKE '%' || ? || '%'", msg.GetName())
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
					"status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(socialStatuses)), ""), ", "),
				),
				args...,
			)
		} else {
			qb.Where("status = ?", pbKyc.Status_STATUS_VALIDATED)
		}
	}

	return qb
}

func (s *ProofRepository) ScanRow(row pgx.Row) (*pbProofs.Proof, error) {
	var (
		id           string
		section      pbKyc.Section
		recordID     string
		documentType string
		name         string
		documentID   string
		status       pbKyc.Status
	)

	err := row.Scan(
		&id,
		&section,
		&recordID,
		&documentType,
		&name,
		&documentID,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbProofs.Proof{
		Id:           id,
		Section:      section,
		RecordId:     recordID,
		DocumentType: &documentType,
		Name:         &name,
		Document: &pbDocuments.Document{
			Id: documentID,
		},
		Status: status,
	}, nil
}

func (s *ProofRepository) ScanWithRelationship(row pgx.Row) (*pbProofs.Proof, error) {
	var (
		id           string
		section      pbKyc.Section
		recordID     string
		documentType string
		name         string
		documentID   string
		status       pbKyc.Status
	)

	var (
		docID            sql.NullString
		docFile          sql.NullString
		docFileType      sql.NullString
		docFileTimestamp sql.NullTime
		docStatus        sql.NullInt32
	)

	err := row.Scan(
		&id,
		&section,
		&recordID,
		&documentType,
		&name,
		&documentID,
		&status,
		&docID,
		&docFile,
		&docFileType,
		&docFileTimestamp,
		&docStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbProofs.Proof{
		Id:           id,
		Section:      section,
		RecordId:     recordID,
		DocumentType: &documentType,
		Name:         &name,
		Document: &pbDocuments.Document{
			Id:            util.GetPointString(util.GetSQLNullString(docID)),
			File:          util.GetPointString(util.GetSQLNullString(docFile)),
			FileType:      util.GetPointString(util.GetSQLNullString(docFileType)),
			FileTimestamp: util.GetSQLNullTime(docFileTimestamp),
			Status:        pbCommon.Status(util.GetPointInt32(util.GetSQLNullInt32(docStatus))),
		},
		Status: status,
	}, nil
}
