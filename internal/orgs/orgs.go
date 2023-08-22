package orgs

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbOrgs "davensi.com/core/gen/orgs"
	pbOrgsConnect "davensi.com/core/gen/orgs/orgsconnect"
)

const (
	_package          = "orgs"
	_entityName       = "Org"
	_entityNamePlural = "Orgs"
)

// For singleton Users export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

// ServiceServer implements the OrgsService API
type ServiceServer struct {
	Repo OrgRepository
	pbOrgsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewOrgRepository(db),
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
	req *connect.Request[pbOrgs.CreateRequest],
) (*connect.Response[pbOrgs.CreateResponse], error) {
	validateErr := s.validateCreate(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbOrgs.CreateResponse{
			Response: &pbOrgs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	qb, errInsert := s.Repo.QbInsert(req.Msg)
	if errInsert != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], errInsert.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbOrgs.CreateResponse{
			Response: &pbOrgs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	newOrg, errInsert := common.ExecuteTxWrite(
		ctx,
		s.db,
		sqlStr,
		args,
		ScanRow,
	)
	if errInsert != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating", _entityName, "name = '"+req.Msg.GetName()+"'")
		log.Error().Err(errInsert).Msg(_err.Error())
		return connect.NewResponse(&pbOrgs.CreateResponse{
			Response: &pbOrgs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + errInsert.Error() + ")",
				},
			},
		}), errInsert
	}

	log.Info().Msgf("%s with name = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetType().Enum().String()+"/"+req.Msg.GetName(), newOrg.GetId())
	return connect.NewResponse(&pbOrgs.CreateResponse{
		Response: &pbOrgs.CreateResponse_Org{
			Org: newOrg,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context, req *connect.Request[pbOrgs.UpdateRequest],
) (*connect.Response[pbOrgs.UpdateResponse], error) {
	errQueryUpdate := validateQueryUpdate(req.Msg)
	if errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbOrgs.UpdateResponse{
			Response: &pbOrgs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getOrgResponse, errUpdate := s.getOldOrgsToUpdate(req.Msg)
	if errUpdate != nil {
		log.Error().Err(errUpdate)
		return connect.NewResponse(&pbOrgs.UpdateResponse{
			Response: &pbOrgs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    "update failed: could not get old Org to update",
				},
			},
		}), errUpdate
	}

	errUpdateValue := s.validateUpdateValue(getOrgResponse.Msg.GetOrg(), req.Msg)
	if errUpdateValue != nil {
		log.Error().Err(errUpdateValue.Err)
		return connect.NewResponse(&pbOrgs.UpdateResponse{
			Response: &pbOrgs.UpdateResponse_Error{
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
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbOrgs.UpdateResponse{
			Response: &pbOrgs.UpdateResponse_Error{
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
	updatedOrg, errUpdate := common.ExecuteTxWrite(
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		ScanRow,
	)
	if errUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "updating", _entityName, sel)
		log.Error().Err(errUpdate).Msg(_err.Error())
		return connect.NewResponse(&pbOrgs.UpdateResponse{
			Response: &pbOrgs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + errUpdate.Error() + ")",
				},
			},
		}), errUpdate
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbOrgs.UpdateResponse{
		Response: &pbOrgs.UpdateResponse_Org{
			Org: updatedOrg,
		},
	}), nil
}

func (s *ServiceServer) getOldOrgsToUpdate(req *pbOrgs.UpdateRequest) (*connect.Response[pbOrgs.GetResponse], error) {
	var getOrgRequest *pbOrgs.GetRequest
	switch req.GetSelect().(type) {
	case *pbOrgs.UpdateRequest_ById:
		getOrgRequest = &pbOrgs.GetRequest{
			Select: &pbOrgs.GetRequest_ById{
				ById: req.GetById(),
			},
		}
	case *pbOrgs.UpdateRequest_ByName:
		getOrgRequest = &pbOrgs.GetRequest{
			Select: &pbOrgs.GetRequest_ByName{
				ByName: req.GetByName(),
			},
		}
	}

	return s.Get(context.Background(), &connect.Request[pbOrgs.GetRequest]{
		Msg: getOrgRequest,
	})
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbOrgs.DeleteRequest],
) (*connect.Response[pbOrgs.DeleteResponse], error) {
	var updateRequest pbOrgs.UpdateRequest
	switch req.Msg.Select.(type) {
	case *pbOrgs.DeleteRequest_ById:
		updateRequest.Select = &pbOrgs.UpdateRequest_ById{
			ById: req.Msg.GetById(),
		}
	case *pbOrgs.DeleteRequest_ByName:
		updateRequest.Select = &pbOrgs.UpdateRequest_ByName{
			ByName: req.Msg.GetByName(),
		}
	}
	updateRequest.Status = pbCommon.Status_STATUS_TERMINATED.Enum()
	updateRes, err := s.Update(ctx, connect.NewRequest(&updateRequest))
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		log.Error().Err(err).Msg(err.Error())
		return connect.NewResponse(&pbOrgs.DeleteResponse{
			Response: &pbOrgs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, "id = "+updateRes.Msg.GetOrg().Id)
	return connect.NewResponse(&pbOrgs.DeleteResponse{
		Response: &pbOrgs.DeleteResponse_Org{
			Org: updateRes.Msg.GetOrg(),
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbOrgs.GetRequest],
) (*connect.Response[pbOrgs.GetResponse], error) {
	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbOrgs.GetResponse{
			Response: &pbOrgs.GetResponse_Error{
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
		return connect.NewResponse(&pbOrgs.GetResponse{
			Response: &pbOrgs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	var (
		id      string
		name    string
		orgType uint32
		status  uint32
	)

	err = rows.Scan(
		&id,
		&name,
		&orgType,
		&status,
	)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbOrgs.GetResponse{
			Response: &pbOrgs.GetResponse_Error{
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
		return connect.NewResponse(&pbOrgs.GetResponse{
			Response: &pbOrgs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbOrgs.GetResponse{
		Response: &pbOrgs.GetResponse_Org{
			Org: &pbOrgs.Org{
				Id:     id,
				Name:   name,
				Type:   pbOrgs.Type(orgType),
				Status: pbCommon.Status(status),
			},
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbOrgs.GetListRequest],
	res *connect.ServerStream[pbOrgs.GetListResponse],
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
		var (
			id      string
			name    string
			orgType uint32
			status  uint32
		)
		err := rows.Scan(
			&id,
			&name,
			&orgType,
			&status,
		)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}

		errSend := res.Send(&pbOrgs.GetListResponse{
			Response: &pbOrgs.GetListResponse_Org{
				Org: &pbOrgs.Org{
					Id:     id,
					Name:   name,
					Type:   pbOrgs.Type(orgType),
					Status: pbCommon.Status(status),
				},
			},
		})
		if errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityNamePlural, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	if !hasRows { // If there are no match
		if !rows.Next() {
			_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, args)
			log.Error().Err(err).Msg(_err.Error())
			errSend := res.Send(&pbOrgs.GetListResponse{
				Response: &pbOrgs.GetListResponse_Error{
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
			return nil
		}
	}

	return rows.Err()
}

func streamingErr(res *connect.ServerStream[pbOrgs.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbOrgs.GetListResponse{
		Response: &pbOrgs.GetListResponse_Error{
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
