package users

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbUsers "davensi.com/core/gen/users"
	pbUsersConnect "davensi.com/core/gen/users/usersconnect"
)

const (
	_package          = "users"
	_entityName       = "User"
	_entityNamePlural = "Users"
)

// For singleton Users export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

// ServiceServer implements the UsersService API
type ServiceServer struct {
	Repo UserRepository
	pbUsersConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewUserRepository(db),
		db:   db,
	}
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) Create(ctx context.Context, req *connect.Request[pbUsers.CreateRequest],
) (*connect.Response[pbUsers.CreateResponse], error) {
	// Verify that Login is specified
	errCreation := s.validateCreate(req.Msg)
	if errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbUsers.CreateResponse{
			Response: &pbUsers.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	qb, err := s.Repo.QbInsert(req.Msg)
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"create",
			_package,
			err.Error(),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbUsers.CreateResponse{
			Response: &pbUsers.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	sqlStr, args, _ := qb.GenerateSQL()
	newUser := &pbUsers.User{}

	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	err = crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		insertUserErr := func() error {
			rows, err := tx.Query(ctx, sqlStr, args...)
			if err != nil {
				return err
			}
			defer rows.Close()

			if !rows.Next() {
				return rows.Err()
			}

			newUser, err = s.Repo.ScanRow(rows)
			if err != nil {
				log.Error().Err(err).Msgf("unable to create %s with login = '%s'",
					_entityName,
					req.Msg.GetLogin(),
				)
			}
			return nil
		}()
		if insertUserErr != nil {
			return insertUserErr
		}
		return err
	})
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_package,
			fmt.Sprintf(
				"login = '%s'",
				req.Msg.GetLogin(),
			),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbUsers.CreateResponse{
			Response: &pbUsers.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	log.Info().Msgf("%s with login = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetLogin(), newUser.Id)
	return connect.NewResponse(&pbUsers.CreateResponse{
		Response: &pbUsers.CreateResponse_User{
			User: newUser,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context, req *connect.Request[pbUsers.UpdateRequest],
) (*connect.Response[pbUsers.UpdateResponse], error) {
	errQueryUpdate := validateQueryUpdate(req.Msg)
	if errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbUsers.UpdateResponse{
			Response: &pbUsers.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getUserResponse, err := s.getOldUsersToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUsers.UpdateResponse{
			Response: &pbUsers.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    "update failed: could not get old Org to update",
				},
			},
		}), err
	}

	userBeforeUpdate := getUserResponse.Msg.GetUser()

	errUpdateValue := s.validateUpdateValue(userBeforeUpdate, req.Msg)
	if errUpdateValue != nil {
		log.Error().Err(errUpdateValue.Err)
		return connect.NewResponse(&pbUsers.UpdateResponse{
			Response: &pbUsers.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdateValue.Code,
					Package: _package,
					Text:    errUpdateValue.Err.Error(),
				},
			},
		}), errUpdateValue.Err
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		errGenSQL := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			err.Error(),
		)
		log.Error().Err(errGenSQL.Err)
		return connect.NewResponse(&pbUsers.UpdateResponse{
			Response: &pbUsers.UpdateResponse_Error{
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
	updatedUser, err := common.ExecuteTxWrite[pbUsers.User](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.Repo.ScanRow,
	)
	if err != nil {
		errUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errUpdate.Err)
		return connect.NewResponse(&pbUsers.UpdateResponse{
			Response: &pbUsers.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), errUpdate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbUsers.UpdateResponse{
		Response: &pbUsers.UpdateResponse_User{
			User: updatedUser,
		},
	}), nil
}

func (s *ServiceServer) getOldUsersToUpdate(req *pbUsers.UpdateRequest) (*connect.Response[pbUsers.GetResponse], error) {
	var getUserRequest *pbUsers.GetRequest
	switch req.GetSelect().(type) {
	case *pbUsers.UpdateRequest_ById:
		getUserRequest = &pbUsers.GetRequest{
			Select: &pbUsers.GetRequest_ById{
				ById: req.GetById(),
			},
		}
	case *pbUsers.UpdateRequest_ByLogin:
		getUserRequest = &pbUsers.GetRequest{
			Select: &pbUsers.GetRequest_ByLogin{
				ByLogin: req.GetByLogin(),
			},
		}
	}

	return s.Get(context.Background(), &connect.Request[pbUsers.GetRequest]{
		Msg: getUserRequest,
	})
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbUsers.DeleteRequest],
) (*connect.Response[pbUsers.DeleteResponse], error) {
	var updateRequest pbUsers.UpdateRequest
	switch req.Msg.Select.(type) {
	case *pbUsers.DeleteRequest_ById:
		updateRequest.Select = &pbUsers.UpdateRequest_ById{
			ById: req.Msg.GetById(),
		}
		updateRequest.Status = pbCommon.Status_STATUS_TERMINATED.Enum()
	case *pbUsers.DeleteRequest_ByLogin:
		updateRequest.Select = &pbUsers.UpdateRequest_ByLogin{
			ByLogin: req.Msg.GetByLogin(),
		}
		updateRequest.Status = pbCommon.Status_STATUS_TERMINATED.Enum()
	}

	updateRes, err := s.Update(ctx, connect.NewRequest(&updateRequest))
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		log.Error().Err(err).Msg(err.Error())
		return connect.NewResponse(&pbUsers.DeleteResponse{
			Response: &pbUsers.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, "id = "+updateRes.Msg.GetUser().Id)
	return connect.NewResponse(&pbUsers.DeleteResponse{
		Response: &pbUsers.DeleteResponse_User{
			User: updateRes.Msg.GetUser(),
		},
	}), nil
}

func (s *ServiceServer) Get(ctx context.Context, req *connect.Request[pbUsers.GetRequest]) (*connect.Response[pbUsers.GetResponse], error) {
	if errQueryGet := validateQueryGet(req.Msg); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbUsers.GetResponse{
			Response: &pbUsers.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUsers.GetResponse{
			Response: &pbUsers.GetResponse_Error{
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
		return connect.NewResponse(&pbUsers.GetResponse{
			Response: &pbUsers.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}

	user, err := s.Repo.ScanRow(rows)
	if err != nil {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
			"fetching",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbUsers.GetResponse{
			Response: &pbUsers.GetResponse_Error{
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
		return connect.NewResponse(&pbUsers.GetResponse{
			Response: &pbUsers.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}
	// Start building the response from here
	return connect.NewResponse(&pbUsers.GetResponse{
		Response: &pbUsers.GetResponse_User{
			User: user,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbUsers.GetListRequest],
	res *connect.ServerStream[pbUsers.GetListResponse],
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
				return res.Send(&pbUsers.GetListResponse{
					Response: &pbUsers.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		user, err := s.Repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbUsers.GetListResponse{
						Response: &pbUsers.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		errSend := res.Send(&pbUsers.GetListResponse{
			Response: &pbUsers.GetListResponse_User{
				User: user,
			},
		})
		if errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityNamePlural, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}
