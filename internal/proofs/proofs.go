package proofs

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbDocuments "davensi.com/core/gen/documents"
	pbKyc "davensi.com/core/gen/kyc"
	pbProofs "davensi.com/core/gen/proofs"
	pbProofsconnect "davensi.com/core/gen/proofs/proofsconnect"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/documents"
)

const (
	_package          = "proofs"
	_entityName       = "Proof"
	_entityNamePlural = "Proofs"
)

// ServiceServer implements the ProofsService API
type ServiceServer struct {
	Repo ProofRepository
	pbProofsconnect.ServiceHandler
	db          *pgxpool.Pool
	documentsSS documents.ServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:        *NewProofRepository(db),
		db:          db,
		documentsSS: *documents.GetSingletonServiceServer(db),
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbProofs.CreateRequest],
) (*connect.Response[pbProofs.CreateResponse], error) {
	handleCreateFunc, genErr := s.documentsSS.GenHandleCreationFn(req.Msg.Document)

	if genErr != nil {
		log.Error().Err(genErr.Err)
		return connect.NewResponse(&pbProofs.CreateResponse{
			Response: &pbProofs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    genErr.Code,
					Package: _package,
					Text:    genErr.Err.Error(),
				},
			},
		}), genErr.Err
	}

	var newProof *pbProofs.Proof

	if errExcute := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if handleCreateFunc != nil {
			excutedDocument, errWriteDocument := handleCreateFunc(tx)
			if errWriteDocument != nil || excutedDocument == nil {
				return errWriteDocument
			}
			newProof.Document = excutedDocument
			return nil
		}
		qb, err := s.Repo.QbInsert(req.Msg, newProof.Document)
		if err != nil {
			return err
		}

		sqlStr, args, _ := qb.GenerateSQL()

		excutedProof, err := common.TxWrite(
			ctx,
			tx,
			sqlStr,
			args,
			s.Repo.ScanRow,
		)
		if err != nil {
			return err
		}

		excutedProof.Document = newProof.Document
		newProof = excutedProof

		return nil
	}); errExcute != nil {
		commonErrCreate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_entityName,
			errExcute.Error(),
		)

		log.Error().Err(commonErrCreate.Err)

		return connect.NewResponse(&pbProofs.CreateResponse{
			Response: &pbProofs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErrCreate.Code,
					Package: _package,
					Text:    commonErrCreate.Err.Error(),
				},
			},
		}), commonErrCreate.Err
	}

	log.Info().Msgf("%s created successfully with id = %s",
		_entityName, newProof.GetId())
	return connect.NewResponse(&pbProofs.CreateResponse{
		Response: &pbProofs.CreateResponse_Proof{
			Proof: newProof,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbProofs.UpdateRequest],
) (*connect.Response[pbProofs.UpdateResponse], error) {
	if errQueryUpdate := validateQuery(&pbProofs.Proof{
		Id: req.Msg.Id,
	}); errQueryUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errQueryUpdate.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbProofs.UpdateResponse{
			Response: &pbProofs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbProofs.UpdateResponse{
			Response: &pbProofs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	handleUpdateFunc, _, genErr := s.documentsSS.GenHandleUpdateFn(
		(&documents.UpdateDocumentDto{}).
			FromUpdateDocument(req.Msg.Document, req.Msg.GetId()),
	)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return connect.NewResponse(&pbProofs.UpdateResponse{
			Response: &pbProofs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    genErr.Code,
					Package: _package,
					Text:    genErr.Err.Error(),
				},
			},
		}), genErr.Err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	var updatedProof *pbProofs.Proof

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	if errExcute := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		excutedProof, err := common.TxWrite(
			ctx,
			tx,
			sqlstr,
			sqlArgs,
			s.Repo.ScanRow,
		)
		if err != nil {
			return err
		}
		updatedProof = excutedProof

		if handleUpdateFunc != nil {
			excutedDocument, errWriteDocument := handleUpdateFunc(tx)
			if errWriteDocument != nil || excutedDocument == nil {
				return errWriteDocument
			}
			updatedProof.Document = excutedDocument
			return nil
		}

		return nil
	}); errExcute != nil {
		commonErrCreate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_entityName,
			errExcute.Error(),
		)

		log.Error().Err(commonErrCreate.Err)

		return connect.NewResponse(&pbProofs.UpdateResponse{
			Response: &pbProofs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErrCreate.Code,
					Package: _package,
					Text:    commonErrCreate.Err.Error(),
				},
			},
		}), commonErrCreate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbProofs.UpdateResponse{
		Response: &pbProofs.UpdateResponse_Proof{
			Proof: updatedProof,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbProofs.GetRequest],
) (*connect.Response[pbProofs.GetResponse], error) {
	if errQueryUpdate := validateQuery(&pbProofs.Proof{
		Id: req.Msg.Id,
	}); errQueryUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errQueryUpdate.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbProofs.GetResponse{
			Response: &pbProofs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	qbDocuments := s.documentsSS.Repo.QbGetList(&pbDocuments.GetListRequest{})
	filterDocument, filterArgs := qbDocuments.Filters.GenerateSQL()
	qb := s.Repo.QbGetOne(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON proofs.document_id = documents.id", qbDocuments.TableName)).
		Select(strings.Join(qbDocuments.SelectFields, ", ")).
		Where(filterDocument, filterArgs...)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbProofs.GetResponse{
			Response: &pbProofs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}

	defer rows.Close()

	if !rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbProofs.GetResponse{
			Response: &pbProofs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	proof, err := s.Repo.ScanWithRelationship(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbProofs.GetResponse{
			Response: &pbProofs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	if rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbProofs.GetResponse{
			Response: &pbProofs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbProofs.GetResponse{
		Response: &pbProofs.GetResponse_Proof{
			Proof: proof,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbProofs.GetListRequest],
	res *connect.ServerStream[pbProofs.GetListResponse],
) error {
	if req.Msg.Document == nil {
		req.Msg.Document = &pbDocuments.GetListRequest{}
	}

	qbDocuments := s.documentsSS.Repo.QbGetList(req.Msg.Document)
	filterDocument, filterArgs := qbDocuments.Filters.GenerateSQL()
	qb := s.Repo.QbGetList(req.Msg).
		Join(fmt.Sprintf(
			"LEFT JOIN %s ON %s.document_id = %s.id",
			qbDocuments.TableName,
			_tableName,
			qbDocuments.TableName,
		)).
		Select(strings.Join(qbDocuments.SelectFields, ", ")).
		Where(filterDocument, filterArgs...)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		proof, err := s.Repo.ScanWithRelationship(rows)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}

		if errSend := res.Send(&pbProofs.GetListResponse{
			Response: &pbProofs.GetListResponse_Proof{
				Proof: proof,
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func streamingErr(res *connect.ServerStream[pbProofs.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbProofs.GetListResponse{
		Response: &pbProofs.GetListResponse_Error{
			Error: &pbCommon.Error{
				Code: _errno,
				Text: _err.Error() + " (" + err.Error() + ")",
			},
		},
	}); errSend != nil {
		_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
		_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
		log.Error().Err(errSend).Msg(_errSend.Error())
		_err = _errSend
	}
	return _err
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbProofs.DeleteRequest],
) (*connect.Response[pbProofs.DeleteResponse], error) {
	canceledStatus := pbKyc.Status_STATUS_CANCELED
	proof := &pbProofs.UpdateRequest{
		Id:     req.Msg.GetId(),
		Status: &canceledStatus,
	}
	deleteReq := connect.NewRequest(proof)
	deletedProof, err := s.Update(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbProofs.DeleteResponse{
			Response: &pbProofs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedProof.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedProof.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbProofs.DeleteResponse{
		Response: &pbProofs.DeleteResponse_Proof{
			Proof: deletedProof.Msg.GetProof(),
		},
	}), nil
}
