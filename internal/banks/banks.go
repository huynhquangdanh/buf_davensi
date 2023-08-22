package banks

import (
	"context"
	"fmt"
	"sync"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbBanks "davensi.com/core/gen/banks"
	pbBanksConnect "davensi.com/core/gen/banks/banksconnect"
	pbCommon "davensi.com/core/gen/common"

	"davensi.com/core/internal/common"
)

const (
	_package          = "banks"
	_entityName       = "Bank"
	_entityNamePlural = "Banks"
)

// ServiceServer implements the BanksService API
type ServiceServer struct {
	Repo BankRepository
	pbBanksConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewBankRepository(db),
		db:   db,
	}
}

// For singleton Bank export module
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

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbBanks.GetRequest],
) (*connect.Response[pbBanks.GetResponse], error) {
	// Validate input
	validateErr := s.ValidateGet(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbBanks.GetResponse{
			Response: &pbBanks.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	// Query into banks
	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbBanks.GetResponse{
			Response: &pbBanks.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), _err
	}
	defer rows.Close()

	if !rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbBanks.GetResponse{
			Response: &pbBanks.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// result but without bank.address.country.fiatList and CryptoList
	bank, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbBanks.GetResponse{
			Response: &pbBanks.GetResponse_Error{
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
		return connect.NewResponse(&pbBanks.GetResponse{
			Response: &pbBanks.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	rows.Close()

	// Start building the response from here
	return connect.NewResponse(&pbBanks.GetResponse{
		Response: &pbBanks.GetResponse_Bank{
			Bank: bank,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbBanks.GetListRequest],
	res *connect.ServerStream[pbBanks.GetListResponse],
) error {
	qb := s.Repo.QbGetList(req.Msg)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)
	}

	defer rows.Close()

	// Start building the response from here
	hasRows := false
	for rows.Next() {
		hasRows = true
		banks, err := s.Repo.ScanRow(rows)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}

		errSend := res.Send(&pbBanks.GetListResponse{
			Response: &pbBanks.GetListResponse_Bank{
				Bank: banks,
			},
		})
		if errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityNamePlural, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
			return _errSend
		}
	}

	if !hasRows { // If there are no match
		if !rows.Next() {
			_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, args)
			log.Error().Err(err).Msg(_err.Error())
			errSend := res.Send(&pbBanks.GetListResponse{
				Response: &pbBanks.GetListResponse_Error{
					Error: &pbCommon.Error{
						Code:    _errno,
						Package: _package,
						Text:    _err.Error(),
					},
				},
			})
			if errSend != nil {
				_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
				_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityNamePlural, "<Selection>")
				log.Error().Err(errSend).Msg(_errSend.Error())
				return _errSend
			}
			return _err
		}
	}

	return rows.Err()
}

func (s *ServiceServer) Create(ctx context.Context, req *connect.Request[pbBanks.CreateRequest],
) (*connect.Response[pbBanks.CreateResponse], error) {
	// Verify that Name, Bic, BankCode is specified
	parentID, validateErr := s.validateCreate(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbBanks.CreateResponse{
			Response: &pbBanks.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	qb, err := s.Repo.QbInsert(req.Msg, parentID)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], err.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbBanks.CreateResponse{
			Response: &pbBanks.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	qb.SetReturnFields("*")
	sqlStr, args, _ := qb.GenerateSQL()

	// Query and building response
	var newBank *pbBanks.Bank
	var errScan error
	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	err = crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		row, err := tx.Query(ctx, sqlStr, args...)
		if err != nil {
			return err
		}
		defer row.Close()

		if row.Next() {
			newBank, errScan = s.Repo.ScanRow(row)
			if errScan != nil {
				log.Error().Err(err).Msgf("unable to create %s with Name = '%s'", _entityName, req.Msg.GetName())
				return errScan
			}
			log.Info().Msgf("%s with name = '%s' created successfully with id = %s",
				_entityName, req.Msg.GetName(), newBank.GetId())
			row.Close()
			return nil
		} else {
			return err
		}
	})
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating", _entityName, "name = '"+req.Msg.GetName()+"'")
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbBanks.CreateResponse{
			Response: &pbBanks.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	// Start building the response from here
	return connect.NewResponse(&pbBanks.CreateResponse{
		Response: &pbBanks.CreateResponse_Bank{
			Bank: newBank,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context, req *connect.Request[pbBanks.UpdateRequest],
) (*connect.Response[pbBanks.UpdateResponse], error) {
	updateBankID, pkResNew, validateErr := s.validateUpdate(req.Msg)
	if validateErr.Err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", validateErr.Err.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbBanks.UpdateResponse{
			Response: &pbBanks.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg, updateBankID, pkResNew)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbBanks.UpdateResponse{
			Response: &pbBanks.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlStr, args, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL " + sqlStr + "")
	updateBank, updateErr := common.ExecuteTxWrite[pbBanks.Bank](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanRow,
	)
	if updateErr != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "updating", _entityName, sel)
		log.Error().Err(updateErr).Msg(_err.Error())
		return connect.NewResponse(&pbBanks.UpdateResponse{
			Response: &pbBanks.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + updateErr.Error() + ")",
				},
			},
		}), updateErr
	}

	// Start building the response from here
	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbBanks.UpdateResponse{
		Response: &pbBanks.UpdateResponse_Bank{
			Bank: updateBank,
		},
	}), nil
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbBanks.DeleteRequest],
) (*connect.Response[pbBanks.DeleteResponse], error) {
	updateRequest := &pbBanks.UpdateRequest{
		Select: req.Msg.GetSelect(),
	}

	updateRes, err := s.Update(ctx, connect.NewRequest(updateRequest))
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		log.Error().Err(err).Msg(err.Error())
		return connect.NewResponse(&pbBanks.DeleteResponse{
			Response: &pbBanks.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, "id = "+updateRes.Msg.GetBank().Id)
	return connect.NewResponse(&pbBanks.DeleteResponse{
		Response: &pbBanks.DeleteResponse_Bank{
			Bank: updateRes.Msg.GetBank(),
		},
	}), nil
}

func streamingErr(res *connect.ServerStream[pbBanks.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbBanks.GetListResponse{
		Response: &pbBanks.GetListResponse_Error{
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
