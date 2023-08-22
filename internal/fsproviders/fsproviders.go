package fsproviders

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbFSProviders "davensi.com/core/gen/fsproviders"
	pbFSProvidersConnect "davensi.com/core/gen/fsproviders/fsprovidersconnect"
)

const (
	_package          = "fsproviders"
	_entityName       = "FS Provider"
	_entityNamePlural = "FS Providers"
)

// ServiceServer implements the UoMsService API
type ServiceServer struct {
	Repo FSProviderRepository
	pbFSProvidersConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewFSProviderRepository(db),
		db:   db,
	}
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbFSProviders.CreateRequest],
) (*connect.Response[pbFSProviders.CreateResponse], error) {
	if validateErr := s.validateCreate(req.Msg); validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbFSProviders.CreateResponse{
			Response: &pbFSProviders.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	qb, err := s.Repo.QbInsert(req.Msg)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbFSProviders.CreateResponse{
			Response: &pbFSProviders.CreateResponse_Error{
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
	newFsprovider, err := common.ExecuteTxWrite[pbFSProviders.FSProvider](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanRow,
	)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
			"type/name = '"+req.Msg.GetType().Enum().String()+"/"+req.Msg.GetName()+"'",
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbFSProviders.CreateResponse{
			Response: &pbFSProviders.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with type/name = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetType().Enum().String()+"/"+req.Msg.GetName(), newFsprovider.Id)
	return connect.NewResponse(&pbFSProviders.CreateResponse{
		Response: &pbFSProviders.CreateResponse_Fsprovider{
			Fsprovider: newFsprovider,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbFSProviders.UpdateRequest],
) (*connect.Response[pbFSProviders.UpdateResponse], error) {
	if errQueryUpdate := ValidateSelect(req.Msg.Select, "updating"); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbFSProviders.UpdateResponse{
			Response: &pbFSProviders.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getFsproviderResponse, errGetOldProvider := s.Get(context.Background(), &connect.Request[pbFSProviders.GetRequest]{
		Msg: &pbFSProviders.GetRequest{
			Select: req.Msg.Select,
		},
	})

	if errGetOldProvider != nil {
		log.Error().Err(errGetOldProvider)
		return connect.NewResponse(&pbFSProviders.UpdateResponse{
			Response: &pbFSProviders.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getFsproviderResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getFsproviderResponse.Msg.GetError().Text,
				},
			},
		}), errGetOldProvider
	}

	fsproviderBeforeUpdate := getFsproviderResponse.Msg.GetFsprovider()

	if errValidateValue := s.validateUpdateValue(fsproviderBeforeUpdate, req.Msg); errValidateValue != nil {
		log.Error().Err(errValidateValue.Err)
		return connect.NewResponse(&pbFSProviders.UpdateResponse{
			Response: &pbFSProviders.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errValidateValue.Code,
					Package: _package,
					Text:    errValidateValue.Err.Error(),
				},
			},
		}), errValidateValue.Err
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		errWithCode := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating",
			_entityName,
			genSQLError.Error(),
		)
		log.Error().Err(errWithCode.Err)
		return connect.NewResponse(&pbFSProviders.UpdateResponse{
			Response: &pbFSProviders.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errWithCode.Code,
					Package: _package,
					Text:    errWithCode.Err.Error(),
				},
			},
		}), errWithCode.Err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	updatedFsprovider, errGetOldProvider := common.ExecuteTxWrite[pbFSProviders.FSProvider](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.Repo.ScanRow,
	)
	if errGetOldProvider != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			sel,
		)
		log.Error().Err(errGetOldProvider).Msg(_err.Error())
		return connect.NewResponse(&pbFSProviders.UpdateResponse{
			Response: &pbFSProviders.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + errGetOldProvider.Error() + ")",
				},
			},
		}), errGetOldProvider
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbFSProviders.UpdateResponse{
		Response: &pbFSProviders.UpdateResponse_Fsprovider{
			Fsprovider: updatedFsprovider,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbFSProviders.GetRequest],
) (*connect.Response[pbFSProviders.GetResponse], error) {
	if validateErr := ValidateSelect(req.Msg.Select, "fetching"); validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbFSProviders.GetResponse{
			Response: &pbFSProviders.GetResponse_Error{
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

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbFSProviders.GetResponse{
			Response: &pbFSProviders.GetResponse_Error{
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
		return connect.NewResponse(&pbFSProviders.GetResponse{
			Response: &pbFSProviders.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	fsProvider, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbFSProviders.GetResponse{
			Response: &pbFSProviders.GetResponse_Error{
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
		return connect.NewResponse(&pbFSProviders.GetResponse{
			Response: &pbFSProviders.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbFSProviders.GetResponse{
		Response: &pbFSProviders.GetResponse_Fsprovider{
			Fsprovider: fsProvider,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbFSProviders.GetListRequest],
	res *connect.ServerStream[pbFSProviders.GetListResponse],
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
				return res.Send(&pbFSProviders.GetListResponse{
					Response: &pbFSProviders.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		fsProvider, err := s.Repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbFSProviders.GetListResponse{
						Response: &pbFSProviders.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbFSProviders.GetListResponse{
			Response: &pbFSProviders.GetListResponse_Fsprovider{
				Fsprovider: fsProvider,
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
	req *connect.Request[pbFSProviders.DeleteRequest],
) (*connect.Response[pbFSProviders.DeleteResponse], error) {
	deletedContact, err := s.Update(ctx, connect.NewRequest(&pbFSProviders.UpdateRequest{
		Select: req.Msg.Select,
	}))
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbFSProviders.DeleteResponse{
			Response: &pbFSProviders.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedContact.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedContact.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbFSProviders.DeleteResponse{
		Response: &pbFSProviders.DeleteResponse_Fsprovider{
			Fsprovider: deletedContact.Msg.GetFsprovider(),
		},
	}), nil
}
