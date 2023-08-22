package documents

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbDocuments "davensi.com/core/gen/documents"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/documentdata"
	"github.com/jackc/pgx/v5"
)

func (s *ServiceServer) GenHandleCreationFn(msg *pbDocuments.CreateRequest) (
	handleFn func(tx pgx.Tx) (*pbDocuments.Document, error),
	err *common.ErrWithCode,
) {
	if msg == nil {
		return nil, nil
	}

	errGenFn := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
		"creating",
		_entityName,
		"",
	)

	if validateErr := validateCreation(msg); validateErr != nil {
		return nil, validateErr
	}

	qb, errInsert := s.Repo.QbInsert(msg)
	if errInsert != nil {
		return nil, errGenFn.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT).
			UpdateMessage(errInsert.Error())
	}

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")

	return func(tx pgx.Tx) (*pbDocuments.Document, error) {
		excutedDocument, errWriteDocument := common.TxWrite[pbDocuments.Document](
			context.Background(),
			tx,
			sqlStr,
			args,
			s.Repo.ScanMainEntity,
		)
		if errWriteDocument != nil {
			return nil, errWriteDocument
		}
		qbDD, errInsertDData := documentdata.QbUpsert(excutedDocument.Id, msg.GetData())
		if errInsertDData != nil {
			return nil, errInsertDData
		}
		sqlDDStr, ddArgs, _ := qbDD.GenerateSQL()
		fmt.Println(sqlDDStr, ddArgs)
		excutedDocumentData, errWriteDD := common.TxWriteMulti[documentdata.DocumentData](
			context.Background(),
			tx,
			sqlDDStr,
			ddArgs,
			documentdata.ScanRow,
		)
		if errWriteDD != nil {
			return nil, errWriteDD
		}

		excutedDocument.Data = excutedDocumentData.Data

		return excutedDocument, nil
	}, nil
}

func (s *ServiceServer) GenHandleUpdateFn(msg *UpdateDocumentDto) (
	handleFn func(tx pgx.Tx) (*pbDocuments.Document, error),
	sel string,
	err *common.ErrWithCode,
) {
	if msg == nil {
		return nil, "", nil
	}

	errGen := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
		"updating",
		_entityName,
		"",
	)

	qb, errGenUpdate := s.Repo.QbUpdate(msg)
	if errGenUpdate != nil {
		errGen.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT).
			UpdateMessage(errGenUpdate.Error())

		return nil, "", errGen
	}

	sqlStr, args, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	return func(tx pgx.Tx) (*pbDocuments.Document, error) {
		return common.TxWrite(
			context.Background(),
			tx,
			sqlStr,
			args,
			s.Repo.ScanMainEntity,
		)
	}, sel, nil
}
