package cryptocategories

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbCryptocategories "davensi.com/core/gen/cryptocategories"
	pbCryptocategoriesconnect "davensi.com/core/gen/cryptocategories/cryptocategoriesconnect"
	"davensi.com/core/internal/common"
)

const (
	_package          = "cryptocategories"
	_entityName       = "CryptoCategory"
	_entityNamePlural = "Cryptocategories"
)

// ServiceServer implements the CryptocategoriesService API
type ServiceServer struct {
	Repo CryptoCategoryRepository
	pbCryptocategoriesconnect.ServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewCryptoCategoryRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbCryptocategories.CreateRequest],
) (*connect.Response[pbCryptocategories.CreateResponse], error) {
	if _errno, validateErr := s.ValidateCreate(req.Msg); validateErr != nil {
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			validateErr.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbCryptocategories.CreateResponse{
			Response: &pbCryptocategories.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	qb, err := s.Repo.QbInsert(req.Msg)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbCryptocategories.CreateResponse{
			Response: &pbCryptocategories.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sqlStr, args...)
		return err
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCryptocategories.CreateResponse{
			Response: &pbCryptocategories.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	// Start building the response from here
	newCryptoCategory, err := s.Get(ctx, &connect.Request[pbCryptocategories.GetRequest]{
		Msg: &pbCryptocategories.GetRequest{
			Select: &pbCryptocategories.GetRequest_ByName{
				ByName: req.Msg.GetName(),
			},
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("unable to create %s",
			_entityName,
		)
		return connect.NewResponse(&pbCryptocategories.CreateResponse{
			Response: &pbCryptocategories.CreateResponse_Error{
				Error: newCryptoCategory.Msg.GetError(),
			},
		}), err
	}
	log.Info().Msgf("%s created successfully with id = %s",
		_entityName, newCryptoCategory.Msg.GetCryptocategory().GetId())
	return connect.NewResponse(&pbCryptocategories.CreateResponse{
		Response: &pbCryptocategories.CreateResponse_Cryptocategory{
			Cryptocategory: newCryptoCategory.Msg.GetCryptocategory(),
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbCryptocategories.UpdateRequest],
) (*connect.Response[pbCryptocategories.UpdateResponse], error) {
	if errQueryUpdate := validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errQueryUpdate.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbCryptocategories.UpdateResponse{
			Response: &pbCryptocategories.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	getCryptoCategoryResponse, err := s.getOldCryptoCategoryToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbCryptocategories.UpdateResponse{
			Response: &pbCryptocategories.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getCryptoCategoryResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getCryptoCategoryResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	countrieBeforeUpdate := getCryptoCategoryResponse.Msg.GetCryptocategory()

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbCryptocategories.UpdateResponse{
			Response: &pbCryptocategories.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sqlstr, sqlArgs...)
		return err
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			sel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCryptocategories.UpdateResponse{
			Response: &pbCryptocategories.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	// Start building the response from here
	updatedCryptoCategory, err := s.Get(ctx, &connect.Request[pbCryptocategories.GetRequest]{
		Msg: &pbCryptocategories.GetRequest{
			Select: &pbCryptocategories.GetRequest_ById{
				ById: countrieBeforeUpdate.Id,
			},
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("unable to update %s with %s", _entityName, sel)
		return connect.NewResponse(&pbCryptocategories.UpdateResponse{
			Response: &pbCryptocategories.UpdateResponse_Error{
				Error: updatedCryptoCategory.Msg.GetError(),
			},
		}), err
	}
	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbCryptocategories.UpdateResponse{
		Response: &pbCryptocategories.UpdateResponse_Cryptocategory{
			Cryptocategory: updatedCryptoCategory.Msg.GetCryptocategory(),
		},
	}), nil
}

func (s *ServiceServer) getOldCryptoCategoryToUpdate(
	msg *pbCryptocategories.UpdateRequest,
) (*connect.Response[pbCryptocategories.GetResponse], error) {
	getCryptoCategoryRequest := &pbCryptocategories.GetRequest{}

	switch msg.Select.(type) {
	case *pbCryptocategories.UpdateRequest_ById:
		getCryptoCategoryRequest.Select = &pbCryptocategories.GetRequest_ById{
			ById: msg.GetById(),
		}
	case *pbCryptocategories.UpdateRequest_ByName:
		getCryptoCategoryRequest.Select = &pbCryptocategories.GetRequest_ByName{
			ByName: msg.GetByName(),
		}
	}

	getCryptoCategoryRes, err := s.Get(context.Background(), &connect.Request[pbCryptocategories.GetRequest]{
		Msg: getCryptoCategoryRequest,
	})
	if err != nil {
		return nil, err
	}

	return getCryptoCategoryRes, nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbCryptocategories.GetRequest],
) (*connect.Response[pbCryptocategories.GetResponse], error) {
	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCryptocategories.GetResponse{
			Response: &pbCryptocategories.GetResponse_Error{
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
		return connect.NewResponse(&pbCryptocategories.GetResponse{
			Response: &pbCryptocategories.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	countrie, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCryptocategories.GetResponse{
			Response: &pbCryptocategories.GetResponse_Error{
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
		return connect.NewResponse(&pbCryptocategories.GetResponse{
			Response: &pbCryptocategories.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbCryptocategories.GetResponse{
		Response: &pbCryptocategories.GetResponse_Cryptocategory{
			Cryptocategory: countrie,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbCryptocategories.GetListRequest],
	res *connect.ServerStream[pbCryptocategories.GetListResponse],
) error {
	qb := s.Repo.QbGetList(req.Msg)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbCryptocategories.GetListResponse{
					Response: &pbCryptocategories.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		countrie, err := s.Repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbCryptocategories.GetListResponse{
						Response: &pbCryptocategories.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbCryptocategories.GetListResponse{
			Response: &pbCryptocategories.GetListResponse_Cryptocategory{
				Cryptocategory: countrie,
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
	req *connect.Request[pbCryptocategories.DeleteRequest],
) (*connect.Response[pbCryptocategories.DeleteResponse], error) {
	updateCryptoCategoryRequest := &pbCryptocategories.UpdateRequest{
		Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
	}

	switch req.Msg.Select.(type) {
	case *pbCryptocategories.DeleteRequest_ById:
		updateCryptoCategoryRequest.Select = &pbCryptocategories.UpdateRequest_ById{
			ById: req.Msg.GetById(),
		}
	case *pbCryptocategories.DeleteRequest_ByName:
		updateCryptoCategoryRequest.Select = &pbCryptocategories.UpdateRequest_ByName{
			ByName: req.Msg.GetByName(),
		}
	}
	terminatedStatus := pbCommon.Status_STATUS_TERMINATED
	countrie := &pbCryptocategories.UpdateRequest{
		Select: &pbCryptocategories.UpdateRequest_ById{
			ById: req.Msg.GetById(),
		},
		Status: &terminatedStatus,
	}
	deleteReq := connect.NewRequest(countrie)
	deletedCryptoCategory, err := s.Update(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbCryptocategories.DeleteResponse{
			Response: &pbCryptocategories.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedCryptoCategory.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedCryptoCategory.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbCryptocategories.DeleteResponse{
		Response: &pbCryptocategories.DeleteResponse_Cryptocategory{
			Cryptocategory: deletedCryptoCategory.Msg.GetCryptocategory(),
		},
	}), nil
}
