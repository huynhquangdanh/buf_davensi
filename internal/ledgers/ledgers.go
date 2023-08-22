package ledgers

import (
	"context"
	"fmt"

	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbLedgers "davensi.com/core/gen/ledgers"
	pbLedgersConnect "davensi.com/core/gen/ledgers/ledgersconnect"
)

const (
	_package          = "ledgers"
	_entityName       = "Ledger"
	_entityNamePlural = "Ledgers"
)

// ServiceServer implements the UoMsService API
type ServiceServer struct {
	repo LedgerRepository
	pbLedgersConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo: *NewLedgerRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbLedgers.CreateRequest],
) (*connect.Response[pbLedgers.CreateResponse], error) {
	if errCreation := s.validateCreate(req.Msg); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbLedgers.CreateResponse{
			Response: &pbLedgers.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	qb, err := s.repo.QbInsert(req.Msg)
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"create",
			_package,
			err.Error(),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbLedgers.CreateResponse{
			Response: &pbLedgers.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	newLedger, err := common.ExecuteTxWrite[pbLedgers.Ledger](
		ctx,
		s.db,
		sqlStr,
		args,
		s.repo.ScanRow,
	)
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_package,
			fmt.Sprintf(
				"name = '%s'",
				req.Msg.GetName(),
			),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbLedgers.CreateResponse{
			Response: &pbLedgers.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	log.Info().Msgf("%s with name = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetName(), newLedger.Id)
	return connect.NewResponse(&pbLedgers.CreateResponse{
		Response: &pbLedgers.CreateResponse_Ledger{
			Ledger: newLedger,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbLedgers.UpdateRequest],
) (*connect.Response[pbLedgers.UpdateResponse], error) {
	if errQueryUpdate := validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbLedgers.UpdateResponse{
			Response: &pbLedgers.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getLedgerResponse, err := s.getOldLedgerToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbLedgers.UpdateResponse{
			Response: &pbLedgers.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getLedgerResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getLedgerResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	ledgerBeforeUpdate := getLedgerResponse.Msg.GetLedger()

	if errUpdateValue := s.validateUpdateValue(ledgerBeforeUpdate, req.Msg); errUpdateValue != nil {
		log.Error().Err(errUpdateValue.Err)
		return connect.NewResponse(&pbLedgers.UpdateResponse{
			Response: &pbLedgers.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdateValue.Code,
					Package: _package,
					Text:    errUpdateValue.Err.Error(),
				},
			},
		}), errUpdateValue.Err
	}

	qb, err := s.repo.QbUpdate(req.Msg)
	if err != nil {
		errGenSQL := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			err.Error(),
		)
		log.Error().Err(errGenSQL.Err)
		return connect.NewResponse(&pbLedgers.UpdateResponse{
			Response: &pbLedgers.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errGenSQL.Code,
					Package: _package,
					Text:    errGenSQL.Err.Error(),
				},
			},
		}), errGenSQL.Err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	updatedLedger, err := common.ExecuteTxWrite[pbLedgers.Ledger](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.repo.ScanRow,
	)
	if err != nil {
		errUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errUpdate.Err)
		return connect.NewResponse(&pbLedgers.UpdateResponse{
			Response: &pbLedgers.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), errUpdate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbLedgers.UpdateResponse{
		Response: &pbLedgers.UpdateResponse_Ledger{
			Ledger: updatedLedger,
		},
	}), nil
}

func (s *ServiceServer) getOldLedgerToUpdate(msg *pbLedgers.UpdateRequest) (*connect.Response[pbLedgers.GetResponse], error) {
	var getUomRequest *pbLedgers.GetRequest
	switch msg.GetSelect().(type) {
	case *pbLedgers.UpdateRequest_ById:
		getUomRequest = &pbLedgers.GetRequest{
			Select: &pbLedgers.GetRequest_ById{
				ById: msg.GetById(),
			},
		}
	case *pbLedgers.UpdateRequest_ByName:
		getUomRequest = &pbLedgers.GetRequest{
			Select: &pbLedgers.GetRequest_ByName{
				ByName: msg.GetByName(),
			},
		}
	}

	return s.Get(context.Background(), &connect.Request[pbLedgers.GetRequest]{
		Msg: getUomRequest,
	})
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbLedgers.GetRequest],
) (*connect.Response[pbLedgers.GetResponse], error) {
	if errQueryGet := validateQueryGet(req.Msg); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbLedgers.GetResponse{
			Response: &pbLedgers.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qb := s.repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbLedgers.GetResponse{
			Response: &pbLedgers.GetResponse_Error{
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
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
			"fetching",
			_package,
			sel,
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbLedgers.GetResponse{
			Response: &pbLedgers.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}
	ledger, err := s.repo.ScanRow(rows)
	if err != nil {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
			"fetching",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbLedgers.GetResponse{
			Response: &pbLedgers.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}
	if rows.Next() {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND,
			"fetching",
			_entityNamePlural,
			sel,
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbLedgers.GetResponse{
			Response: &pbLedgers.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbLedgers.GetResponse{
		Response: &pbLedgers.GetResponse_Ledger{
			Ledger: ledger,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbLedgers.GetListRequest],
	res *connect.ServerStream[pbLedgers.GetListResponse],
) error {
	qb := s.repo.QbGetList(req.Msg)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbLedgers.GetListResponse{
					Response: &pbLedgers.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		ledger, err := s.repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbLedgers.GetListResponse{
						Response: &pbLedgers.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbLedgers.GetListResponse{
			Response: &pbLedgers.GetListResponse_Ledger{
				Ledger: ledger,
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbLedgers.DeleteRequest],
) (*connect.Response[pbLedgers.DeleteResponse], error) {
	var updateRequest pbLedgers.UpdateRequest
	switch req.Msg.Select.(type) {
	case *pbLedgers.DeleteRequest_ById:
		updateRequest.Select = &pbLedgers.UpdateRequest_ById{
			ById: req.Msg.GetById(),
		}
	case *pbLedgers.DeleteRequest_ByName:
		updateRequest.Select = &pbLedgers.UpdateRequest_ByName{
			ByName: req.Msg.GetByName(),
		}
	}
	updateRequest.Status = pbCommon.Status_STATUS_TERMINATED.Enum()

	updateRes, err := s.Update(ctx, connect.NewRequest(&updateRequest))
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		log.Error().Err(err).Msg(err.Error())
		return connect.NewResponse(&pbLedgers.DeleteResponse{
			Response: &pbLedgers.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, "id = "+updateRes.Msg.GetLedger().Id)
	return connect.NewResponse(&pbLedgers.DeleteResponse{
		Response: &pbLedgers.DeleteResponse_Ledger{
			Ledger: updateRes.Msg.GetLedger(),
		},
	}), nil
}
