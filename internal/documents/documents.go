package documents

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbDocuments "davensi.com/core/gen/documents"
	pbDocumentsConnect "davensi.com/core/gen/documents/documentsconnect"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/documentdata"
	"github.com/jackc/pgx/v5"
)

const (
	_package          = "documents"
	_entityName       = "Document"
	_entityNamePlural = "Documents"
)

// ServiceServer implements the DocumentsService API
type ServiceServer struct {
	pbDocumentsConnect.ServiceHandler
	Repo DocumentRepository
	db   *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewDocumentRepository(db),
		db:   db,
	}
}

// For singleton UoMs export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbDocuments.CreateRequest],
) (*connect.Response[pbDocuments.CreateResponse], error) {
	handleCreateFunc, genErr := s.GenHandleCreationFn(req.Msg)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return connect.NewResponse(&pbDocuments.CreateResponse{
			Response: &pbDocuments.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    genErr.Code,
					Package: _package,
					Text:    genErr.Err.Error(),
				},
			},
		}), genErr.Err
	}
	var newDocument *pbDocuments.Document

	if errExcute := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		excutedDocument, errWriteDocument := handleCreateFunc(tx)
		if errWriteDocument == nil {
			newDocument = excutedDocument
			return nil
		}
		return errWriteDocument
	}); errExcute != nil {
		commonErrCreate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_entityName,
			errExcute.Error(),
		)

		log.Error().Err(commonErrCreate.Err)

		return connect.NewResponse(&pbDocuments.CreateResponse{
			Response: &pbDocuments.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErrCreate.Code,
					Package: _package,
					Text:    commonErrCreate.Err.Error(),
				},
			},
		}), commonErrCreate.Err
	}

	log.Info().Msgf(
		"%s created successfully with id = %s",
		_entityName, newDocument.GetId(),
	)
	return connect.NewResponse(&pbDocuments.CreateResponse{
		Response: &pbDocuments.CreateResponse_Document{
			Document: newDocument,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbDocuments.UpdateRequest],
) (*connect.Response[pbDocuments.UpdateResponse], error) {
	if errQueryUpdate := validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbDocuments.UpdateResponse{
			Response: &pbDocuments.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	handleUpdateFunc, sel, genErr := s.GenHandleUpdateFn((&UpdateDocumentDto{}).FromUpdateRequest(req.Msg))
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return connect.NewResponse(&pbDocuments.UpdateResponse{
			Response: &pbDocuments.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    genErr.Code,
					Package: _package,
					Text:    genErr.Err.Error(),
				},
			},
		}), genErr.Err
	}

	var updatedDocument *pbDocuments.Document

	if errExcute := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		excutedDocument, errWriteDocument := handleUpdateFunc(tx)
		if errWriteDocument != nil {
			return errWriteDocument
		}
		updatedDocument = excutedDocument
		return nil
	}); errExcute != nil {
		commonErrUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_entityName,
			errExcute.Error(),
		)

		log.Error().Err(commonErrUpdate.Err)

		return connect.NewResponse(&pbDocuments.UpdateResponse{
			Response: &pbDocuments.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErrUpdate.Code,
					Package: _package,
					Text:    commonErrUpdate.Err.Error(),
				},
			},
		}), commonErrUpdate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbDocuments.UpdateResponse{
		Response: &pbDocuments.UpdateResponse_Document{
			Document: updatedDocument,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbDocuments.GetRequest],
) (*connect.Response[pbDocuments.GetResponse], error) {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"fetching",
		_entityName,
		"",
	)

	if errValidateGet := validateQueryGet(req.Msg); errValidateGet != nil {
		log.Error().Err(errValidateGet.Err)

		return connect.NewResponse(&pbDocuments.GetResponse{
			Response: &pbDocuments.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errValidateGet.Code,
					Package: _package,
					Text:    errValidateGet.Err.Error(),
				},
			},
		}), errValidateGet.Err
	}

	qbDocData := documentdata.QbGetList()
	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.
		Join(fmt.Sprintf("LEFT JOIN %s ON documents.id = documents_data.document_id", qbDocData.TableName)).
		Select(strings.Join(qbDocData.SelectFields, ", ")).
		GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, excuteErr := s.db.Query(ctx, sqlstr, sqlArgs...)
	if excuteErr != nil {
		errGet.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(fmt.Sprintf("%s with error: %s", sel, excuteErr.Error()))
		log.Error().Err(errGet.Err)

		return connect.NewResponse(&pbDocuments.GetResponse{
			Response: &pbDocuments.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errGet.Code,
					Package: _package,
					Text:    errGet.Err.Error(),
				},
			},
		}), errGet.Err
	}

	defer rows.Close()

	var document *pbDocuments.Document

	for rows.Next() {
		docDto, scanErr := s.Repo.ScanWithRelationship(rows)
		if scanErr != nil {
			errGet.
				UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR).
				UpdateMessage(fmt.Sprintf("%s with error: %s", sel, scanErr.Error()))
			log.Error().Err(errGet.Err)

			return connect.NewResponse(&pbDocuments.GetResponse{
				Response: &pbDocuments.GetResponse_Error{
					Error: &pbCommon.Error{
						Code:    errGet.Code,
						Package: _package,
						Text:    errGet.Err.Error(),
					},
				},
			}), errGet.Err
		}

		if document == nil {
			document = &pbDocuments.Document{
				Id:            docDto.DocID,
				File:          docDto.DocFile,
				FileType:      docDto.DocFileType,
				FileTimestamp: docDto.GetDocFileTimestamp(),
				Status:        docDto.DocStatus,
				Data:          map[string]string{},
			}
		}

		if docDto.GetDocDataID() != nil {
			if docDto.GetDocDataField() != nil {
				document.Data[*docDto.GetDocDataField()] = docDto.GetDocDataValue()
			}
		}
	}

	if document == nil {
		errGet.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND).
			UpdateMessage(fmt.Sprintf("%s not found", sel))
		log.Error().Err(errGet.Err)

		return connect.NewResponse(&pbDocuments.GetResponse{
			Response: &pbDocuments.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errGet.Code,
					Package: _package,
					Text:    errGet.Err.Error(),
				},
			},
		}), errGet.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbDocuments.GetResponse{
		Response: &pbDocuments.GetResponse_Document{
			Document: document,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbDocuments.GetListRequest],
	res *connect.ServerStream[pbDocuments.GetListResponse],
) error {
	qb := s.Repo.QbGetList(req.Msg)
	qbDocData := documentdata.QbGetList()
	sqlStr, args, _ := qb.
		Join(fmt.Sprintf("LEFT JOIN %s ON documents.id = documents_data.document_id", qbDocData.TableName)).
		Select(strings.Join(qbDocData.SelectFields, ", ")).
		GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbDocuments.GetListResponse{
					Response: &pbDocuments.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	document := &pbDocuments.Document{
		Id: "",
	}

	for rows.Next() {
		docDto, err := s.Repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbDocuments.GetListResponse{
						Response: &pbDocuments.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if document.Id != docDto.DocID {
			if document.Id != "" {
				if errSend := res.Send(&pbDocuments.GetListResponse{
					Response: &pbDocuments.GetListResponse_Document{
						Document: document,
					},
				}); errSend != nil {
					log.Error().Err(common.CreateErrWithCode(
						pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR,
						"fetching",
						_entityNamePlural,
						errSend.Error(),
					).Err)
				}
			}
			document = &pbDocuments.Document{
				Id:            docDto.DocID,
				File:          docDto.DocFile,
				FileType:      docDto.DocFileType,
				FileTimestamp: docDto.GetDocFileTimestamp(),
				Status:        docDto.DocStatus,
				Data:          map[string]string{},
			}
		}
		if docDto.GetDocDataID() != nil {
			if docDto.GetDocDataField() != nil {
				document.Data[*docDto.GetDocDataField()] = docDto.GetDocDataValue()
			}
		}
	}
	if errSend := res.Send(&pbDocuments.GetListResponse{
		Response: &pbDocuments.GetListResponse_Document{
			Document: document,
		},
	}); errSend != nil {
		log.Error().Err(common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR,
			"fetching",
			_entityNamePlural,
			errSend.Error(),
		).Err)
	}

	return rows.Err()
}

func (s *ServiceServer) SetData(
	ctx context.Context,
	req *connect.Request[pbDocuments.SetDataRequest],
) (*connect.Response[pbDocuments.SetDataResponse], error) {
	errSetData := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
		"setdata",
		_entityName,
		"",
	)
	if validateErr := validateSetData(req.Msg); validateErr != nil {
		log.Error().Err(validateErr.Err)

		return connect.NewResponse(&pbDocuments.SetDataResponse{}), validateErr.Err
	}

	qb, errInsert := documentdata.QbUpsert(req.Msg.Id, req.Msg.Data)
	if errInsert != nil {
		errSetData.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT).
			UpdateMessage(errInsert.Error())

		log.Error().Err(errSetData.Err)

		return connect.NewResponse(&pbDocuments.SetDataResponse{}), errSetData.Err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	newDocData := &documentdata.DocumentData{}
	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")

	if errExcute := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		excutedDocumentData, errWriteDD := common.TxWriteMulti[documentdata.DocumentData](
			ctx,
			tx,
			sqlStr,
			args,
			documentdata.ScanRow,
		)
		if errWriteDD != nil {
			return errWriteDD
		}
		newDocData = excutedDocumentData
		return nil
	}); errExcute != nil {
		errSetData.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errExcute.Error())

		log.Error().Err(errSetData.Err)

		return connect.NewResponse(&pbDocuments.SetDataResponse{}), errSetData.Err
	}

	return connect.NewResponse(&pbDocuments.SetDataResponse{
		Id:   req.Msg.Id,
		Data: newDocData.Data,
	}), nil
}

func (s *ServiceServer) UpdateData(
	ctx context.Context,
	req *connect.Request[pbDocuments.UpdateDataRequest],
) (*connect.Response[pbDocuments.UpdateDataResponse], error) {
	errUpdateData := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
		"updatedata",
		_entityName,
		"",
	)
	if validateErr := validateUpdateData(req.Msg); validateErr != nil {
		log.Error().Err(validateErr.Err)

		return connect.NewResponse(&pbDocuments.UpdateDataResponse{}), validateErr.Err
	}

	qb, errUpsert := documentdata.QbUpsert(req.Msg.Id, req.Msg.Data)
	if errUpsert != nil {
		errUpdateData.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT).
			UpdateMessage(errUpsert.Error())

		log.Error().Err(errUpdateData.Err)

		return connect.NewResponse(&pbDocuments.UpdateDataResponse{}), errUpdateData.Err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	updateDocData := &documentdata.DocumentData{
		Data: map[string]string{},
	}
	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")

	if errExcute := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		excutedDocumentData, errWriteDD := common.TxWriteMulti[documentdata.DocumentData](
			ctx,
			tx,
			sqlStr,
			args,
			documentdata.ScanRow,
		)
		if errWriteDD == nil {
			updateDocData = excutedDocumentData
			return nil
		}
		return errWriteDD
	}); errExcute != nil {
		errUpdateData.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errExcute.Error())

		log.Error().Err(errUpdateData.Err)

		return connect.NewResponse(&pbDocuments.UpdateDataResponse{}), errUpdateData.Err
	}

	log.Info().Msgf(
		"%s update data successfully with id = %s",
		_entityName, req.Msg.GetId(),
	)
	return connect.NewResponse(&pbDocuments.UpdateDataResponse{
		Id:   req.Msg.Id,
		Data: updateDocData.Data,
	}), nil
}

func (s *ServiceServer) RemoveData(
	ctx context.Context,
	req *connect.Request[pbDocuments.RemoveDataRequest],
) (*connect.Response[pbDocuments.RemoveDataResponse], error) {
	errRemoveData := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
		"setdata",
		_entityName,
		"",
	)
	if validateErr := validateRemoveData(req.Msg); validateErr != nil {
		log.Error().Err(validateErr.Err)

		return connect.NewResponse(&pbDocuments.RemoveDataResponse{}), validateErr.Err
	}

	qb := documentdata.QbRemove(req.Msg.Id, req.Msg.Keys.GetList())
	sqlStr, args, _ := qb.GenerateSQL()

	updatedDocummentData := &documentdata.DocumentData{
		Data: map[string]string{},
	}
	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")

	if errExcute := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		excutedDocumentData, errWriteDD := common.TxWriteMulti[documentdata.DocumentData](
			ctx,
			tx,
			sqlStr,
			args,
			documentdata.ScanRow,
		)
		if errWriteDD != nil {
			return errWriteDD
		}
		if excutedDocumentData != nil {
			updatedDocummentData = excutedDocumentData
		}
		return nil
	}); errExcute != nil {
		errRemoveData.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errExcute.Error())

		log.Error().Err(errRemoveData.Err)

		return connect.NewResponse(&pbDocuments.RemoveDataResponse{}), errRemoveData.Err
	}

	log.Info().Msgf(
		"%s delete successfully with id = %s",
		_entityName, req.Msg.GetId(),
	)
	return connect.NewResponse(&pbDocuments.RemoveDataResponse{
		Id:   req.Msg.Id,
		Data: updatedDocummentData.Data,
	}), nil
}
