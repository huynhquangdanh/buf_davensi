package documents

import (
	"database/sql"

	pbCommon "davensi.com/core/gen/common"
	pbDocuments "davensi.com/core/gen/documents"

	"davensi.com/core/internal/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DocumentDto struct {
	DocID            string          `db:"id"`
	DocFile          string          `db:"file"`
	DocFileType      string          `db:"file_type"`
	DocFileTimestamp sql.NullTime    `db:"file_timestamp"`
	DocStatus        pbCommon.Status `db:"status"`
	DocDataID        sql.NullString  `db:"document_id"`
	DocDataField     sql.NullString  `db:"field"`
	DocDataValue     sql.NullString  `db:"value"`
	DocDataStatus    sql.NullInt32   `db:"status"`
}

func (dto *DocumentDto) GetDocFileTimestamp() *timestamppb.Timestamp {
	return util.GetSQLNullTime(dto.DocFileTimestamp)
}

func (dto *DocumentDto) GetDocDataID() *string {
	return util.GetSQLNullString(dto.DocDataID)
}

func (dto *DocumentDto) GetDocDataField() *string {
	return util.GetSQLNullString(dto.DocDataField)
}

func (dto *DocumentDto) GetDocDataValue() string {
	if value := util.GetSQLNullString(dto.DocDataValue); value != nil {
		return *value
	}
	return ""
}

func (dto *DocumentDto) GetDocDataStatus() pbCommon.Status {
	statusNumber := util.GetSQLNullInt32(dto.DocDataStatus)
	if statusNumber != nil {
		return pbCommon.Status(*statusNumber)
	}
	return pbCommon.Status_STATUS_UNSPECIFIED
}

type UpdateDocumentDto struct {
	ID       string
	File     *string
	FileType *string
	Status   *pbCommon.Status
}

func (dto *UpdateDocumentDto) FromUpdateRequest(msg *pbDocuments.UpdateRequest) *UpdateDocumentDto {
	dto.ID = msg.Id
	dto.File = msg.File
	dto.Status = msg.Status

	return dto
}

func (dto *UpdateDocumentDto) FromUpdateDocument(msg *pbDocuments.UpdateDocument, documentID string) *UpdateDocumentDto {
	dto.ID = documentID
	dto.File = msg.File
	dto.FileType = msg.FileType
	dto.Status = msg.Status

	return dto
}

func (dto *UpdateDocumentDto) GetFile() string {
	if dto.File == nil {
		return ""
	}
	return *dto.File
}

func (dto *UpdateDocumentDto) GetFileType() string {
	if dto.FileType == nil {
		return ""
	}
	return *dto.FileType
}

func (dto *UpdateDocumentDto) GetStatus() pbCommon.Status {
	if dto.Status == nil {
		return pbCommon.Status_STATUS_UNSPECIFIED
	}
	return *dto.Status
}
