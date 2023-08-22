package documents

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbDocuments "davensi.com/core/gen/documents"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_fields    = "id, file, file_type, file_timestamp, status"
	_tableName = "core.documents"
)

type DocumentRepository struct {
	db *pgxpool.Pool
}

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{
		db: db,
	}
}

func (s *DocumentRepository) QbInsert(msg *pbDocuments.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleValue := []any{}

	// Append required value
	qb.SetInsertField("file")
	file := msg.GetFile()
	singleValue = append(singleValue, file)

	qb.SetInsertField("file_type")
	fileType := filepath.Ext(file)
	singleValue = append(singleValue, fileType)

	// Append optional fields values
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleValue = append(singleValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleValue)

	return qb, err
}

func (s *DocumentRepository) QbUpdate(msg *UpdateDocumentDto) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.File != nil {
		file := msg.GetFile()
		qb.SetUpdate("file", file)

		if msg.FileType == nil {
			fileType := filepath.Ext(file)
			qb.SetUpdate("file_type", fileType)
		}
	}

	if msg.FileType != nil {
		qb.SetUpdate("file_type", msg.GetFileType())
	}

	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("id = ?", msg.ID)

	return qb, nil
}

func (s *DocumentRepository) QbGetOne(msg *pbDocuments.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, _tableName))

	qb.Where("documents.id = ?", msg.GetId())

	return qb
}

func (s *DocumentRepository) QbGetList(msg *pbDocuments.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_fields, _tableName))

	if msg.File != nil {
		qb.Where("documents.file LIKE '%' || ? || '%'", msg.GetFile())
	}

	if msg.FileType != nil {
		qb.Where("documents.file_type LIKE '%' || ? || '%'", msg.GetFileType())
	}

	if msg.Status != nil {
		statuses := msg.GetStatus().GetList()

		if len(statuses) > 0 {
			args := []any{}
			for _, v := range statuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"documents.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(statuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *DocumentRepository) ScanMainEntity(row pgx.Row) (*pbDocuments.Document, error) {
	var (
		id            string
		file          string
		fileType      string
		fileTimestamp sql.NullTime
		status        pbCommon.Status
	)

	err := row.Scan(
		&id,
		&file,
		&fileType,
		&fileTimestamp,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbDocuments.Document{
		Id:            id,
		File:          file,
		FileType:      fileType,
		FileTimestamp: util.GetSQLNullTime(fileTimestamp),
		Status:        status,
	}, nil
}

func (s *DocumentRepository) ScanWithRelationship(row pgx.Rows) (*DocumentDto, error) {
	var docDto DocumentDto

	err := row.Scan(
		&docDto.DocID,
		&docDto.DocFile,
		&docDto.DocFileType,
		&docDto.DocFileTimestamp,
		&docDto.DocStatus,
		&docDto.DocDataID,
		&docDto.DocDataField,
		&docDto.DocDataValue,
		&docDto.DocDataStatus,
	)
	if err != nil {
		return nil, err
	}

	return &docDto, nil
}
