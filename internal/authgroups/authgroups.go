package authgroups

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbAuthGroups "davensi.com/core/gen/authgroups"
	pbAuthGroupsConnect "davensi.com/core/gen/authgroups/authgroupsconnect"
	pbCommon "davensi.com/core/gen/common"
	"davensi.com/core/internal/common"
)

const (
	_package          = "authgroups"
	_entityName       = "Auth Group"
	_entityNamePlural = "Auth Groups"
)

// ServiceServer implements the AuthGroupsService API
type ServiceServer struct {
	repo AuthGroupRepository
	pbAuthGroupsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo: *NewAuthGroupRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbAuthGroups.CreateRequest],
) (*connect.Response[pbAuthGroups.CreateResponse], error) {
	if _errno, validateErr := s.validateCreate(req.Msg); validateErr != nil {
		log.Error().Err(validateErr)
		return connect.NewResponse(&pbAuthGroups.CreateResponse{
			Response: &pbAuthGroups.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    validateErr.Error(),
				},
			},
		}), validateErr
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
		return connect.NewResponse(&pbAuthGroups.CreateResponse{
			Response: &pbAuthGroups.CreateResponse_Error{
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
	newAuthGroup, err := common.ExecuteTxWrite[pbAuthGroups.AuthGroup](
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
		return connect.NewResponse(&pbAuthGroups.CreateResponse{
			Response: &pbAuthGroups.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	log.Info().Msgf("%s with name = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetName(), newAuthGroup.Id)
	return connect.NewResponse(&pbAuthGroups.CreateResponse{
		Response: &pbAuthGroups.CreateResponse_Authgroup{
			Authgroup: newAuthGroup,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbAuthGroups.UpdateRequest],
) (*connect.Response[pbAuthGroups.UpdateResponse], error) {
	if errQueryUpdate := s.validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbAuthGroups.UpdateResponse{
			Response: &pbAuthGroups.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	authGroupBeforeUpdate, err := s.getOldAuthGroupToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbAuthGroups.UpdateResponse{
			Response: &pbAuthGroups.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
					Package: _package,
					Text:    "update failed: " + err.Error(),
				},
			},
		}), err
	}

	if errUpdateValue := s.validateUpdateValue(authGroupBeforeUpdate, req.Msg); errUpdateValue != nil {
		log.Error().Err(errUpdateValue.Err)
		return connect.NewResponse(&pbAuthGroups.UpdateResponse{
			Response: &pbAuthGroups.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdateValue.Code,
					Package: _package,
					Text:    errUpdateValue.Err.Error(),
				},
			},
		}), errUpdateValue.Err
	}

	qb, genSQLError := s.repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		errGenSQL := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			err.Error(),
		)
		log.Error().Err(errGenSQL.Err)
		return connect.NewResponse(&pbAuthGroups.UpdateResponse{
			Response: &pbAuthGroups.UpdateResponse_Error{
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
	updatedAuthGroup, err := common.ExecuteTxWrite[pbAuthGroups.AuthGroup](
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
		return connect.NewResponse(&pbAuthGroups.UpdateResponse{
			Response: &pbAuthGroups.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), errUpdate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbAuthGroups.UpdateResponse{
		Response: &pbAuthGroups.UpdateResponse_Authgroup{
			Authgroup: updatedAuthGroup,
		},
	}), nil
}

func (s *ServiceServer) getOldAuthGroupToUpdate(msg *pbAuthGroups.UpdateRequest) (*pbAuthGroups.AuthGroup, error) {
	var getUomRequest *pbAuthGroups.GetRequest
	switch msg.GetSelect().Select.(type) {
	case *pbAuthGroups.Select_ById:
		getUomRequest = &pbAuthGroups.GetRequest{
			Select: &pbAuthGroups.Select{
				Select: &pbAuthGroups.Select_ById{
					ById: msg.GetSelect().GetById(),
				},
			},
		}
	case *pbAuthGroups.Select_ByName:
		getUomRequest = &pbAuthGroups.GetRequest{
			Select: &pbAuthGroups.Select{
				Select: &pbAuthGroups.Select_ByName{
					ByName: msg.GetSelect().GetByName(),
				},
			},
		}
	}

	qb := s.repo.QbGetOne(getUomRequest, nil)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(context.Background(), sqlstr, sqlArgs...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return nil, fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
	}
	getAuthGroupRes, err := s.repo.ScanRow(rows)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, sel)
		log.Error().Err(err).Msg(_err.Error())
		return nil, _err
	}

	return getAuthGroupRes, nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbAuthGroups.GetRequest],
) (*connect.Response[pbAuthGroups.GetResponse], error) {
	if errQueryGet := validateQueryGet(req.Msg); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbAuthGroups.GetResponse{
			Response: &pbAuthGroups.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}
	status := pbCommon.Status_STATUS_ACTIVE
	qb := s.repo.QbGetOne(req.Msg, &status)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbAuthGroups.GetResponse{
			Response: &pbAuthGroups.GetResponse_Error{
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
		return connect.NewResponse(&pbAuthGroups.GetResponse{
			Response: &pbAuthGroups.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	authGroup, err := s.repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbAuthGroups.GetResponse{
			Response: &pbAuthGroups.GetResponse_Error{
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
		return connect.NewResponse(&pbAuthGroups.GetResponse{
			Response: &pbAuthGroups.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbAuthGroups.GetResponse{
		Response: &pbAuthGroups.GetResponse_Authgroup{
			Authgroup: authGroup,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbAuthGroups.GetListRequest],
	res *connect.ServerStream[pbAuthGroups.GetListResponse],
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
				return res.Send(&pbAuthGroups.GetListResponse{
					Response: &pbAuthGroups.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		authGroup, err := s.repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbAuthGroups.GetListResponse{
						Response: &pbAuthGroups.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbAuthGroups.GetListResponse{
			Response: &pbAuthGroups.GetListResponse_Authgroup{
				Authgroup: authGroup,
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
	req *connect.Request[pbAuthGroups.DeleteRequest],
) (*connect.Response[pbAuthGroups.DeleteResponse], error) {
	var updateRequest pbAuthGroups.UpdateRequest
	switch req.Msg.GetSelect().Select.(type) {
	case *pbAuthGroups.Select_ById:
		updateRequest.Select = &pbAuthGroups.Select{
			Select: &pbAuthGroups.Select_ById{
				ById: req.Msg.GetSelect().GetById(),
			},
		}
		updateRequest.Status = pbCommon.Status_STATUS_TERMINATED.Enum()
	case *pbAuthGroups.Select_ByName:
		updateRequest.Select = &pbAuthGroups.Select{
			Select: &pbAuthGroups.Select_ByName{
				ByName: req.Msg.GetSelect().GetByName(),
			},
		}
		updateRequest.Status = pbCommon.Status_STATUS_TERMINATED.Enum()
	}

	updateRes, err := s.Update(ctx, connect.NewRequest(&updateRequest))
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		log.Error().Err(err).Msg(err.Error())
		return connect.NewResponse(&pbAuthGroups.DeleteResponse{
			Response: &pbAuthGroups.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, "id = "+updateRes.Msg.GetAuthgroup().Id)
	return connect.NewResponse(&pbAuthGroups.DeleteResponse{
		Response: &pbAuthGroups.DeleteResponse_Authgroup{
			Authgroup: updateRes.Msg.GetAuthgroup(),
		},
	}), nil
}
