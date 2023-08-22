package credentials

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/util"

	"github.com/rs/zerolog/log"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"

	pbCommon "davensi.com/core/gen/common"
	pbCredential "davensi.com/core/gen/credentials"
	credentialConnect "davensi.com/core/gen/credentials/credentialsconnect"
	pbKyc "davensi.com/core/gen/kyc"
	"davensi.com/core/internal/common"
)

// ServiceServer implements the AddressesService API
type ServiceServer struct {
	Repo CredentialRepository
	credentialConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

// For singleton Credentials export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewCredentialRepository(db),
		db:   db,
	}
}

// Create function ...
func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbCredential.CreateRequest],
) (*connect.Response[pbCredential.CreateResponse], error) {
	if err := s.validateQueryInsert(req.Msg); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating '"+_entityName+"'", err.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbCredential.CreateResponse{
			Response: &pbCredential.CreateResponse_Error{
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
		return connect.NewResponse(&pbCredential.CreateResponse{
			Response: &pbCredential.CreateResponse_Error{
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

	var newCredential *pbCredential.Credentials

	newCredential, err = common.ExecuteTxWrite[pbCredential.Credentials](
		ctx, s.db, sqlStr, args, s.Repo.ScanRow,
	)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCredential.CreateResponse{
			Response: &pbCredential.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}
	log.Info().Msgf("%s created successfully with id = %s", _entityName, newCredential.GetId())

	return connect.NewResponse(&pbCredential.CreateResponse{
		Response: &pbCredential.CreateResponse_Credentials{
			Credentials: newCredential,
		},
	}), nil
}

func (s *ServiceServer) getOldCredentialToUpdate(msg *pbCredential.UpdateRequest) (*connect.Response[pbCredential.GetResponse], error) {
	getCredentialRequest := &pbCredential.GetRequest{Id: msg.GetId()}

	getCredentialResp, err := s.Get(context.Background(), &connect.Request[pbCredential.GetRequest]{
		Msg: getCredentialRequest,
	})
	if err != nil {
		return nil, err
	}

	return getCredentialResp, nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbCredential.UpdateRequest],
) (*connect.Response[pbCredential.UpdateResponse], error) {
	if errQueryUpdate := s.validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errQueryUpdate.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbCredential.UpdateResponse{
			Response: &pbCredential.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	getResponse, err := s.getOldCredentialToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		return connect.NewResponse(&pbCredential.UpdateResponse{
			Response: &pbCredential.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	credentialBeforeUpdate := getResponse.Msg.GetCredentials()
	_errno, errUpdateValue := s.validateUpdateValue(credentialBeforeUpdate, req.Msg)
	if errUpdateValue != nil {
		log.Error().Err(errUpdateValue)
		return connect.NewResponse(&pbCredential.UpdateResponse{
			Response: &pbCredential.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    errUpdateValue.Error(),
				},
			},
		}), errUpdateValue
	}
	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbCredential.UpdateResponse{
			Response: &pbCredential.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	qb.SetReturnFields("*")

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	var updatedCredential *pbCredential.Credentials

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	updatedCredential, err = common.ExecuteTxWrite[pbCredential.Credentials](
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
		return connect.NewResponse(&pbCredential.UpdateResponse{
			Response: &pbCredential.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbCredential.UpdateResponse{
		Response: &pbCredential.UpdateResponse_Credentials{
			Credentials: updatedCredential,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbCredential.GetRequest],
) (*connect.Response[pbCredential.GetResponse], error) {
	if err := s.validateMsgGetOne(req.Msg); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", err.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbCredential.GetResponse{
			Response: &pbCredential.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCredential.GetResponse{
			Response: &pbCredential.GetResponse_Error{
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
		return connect.NewResponse(&pbCredential.GetResponse{
			Response: &pbCredential.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	credential, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCredential.GetResponse{
			Response: &pbCredential.GetResponse_Error{
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
		return connect.NewResponse(&pbCredential.GetResponse{
			Response: &pbCredential.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbCredential.GetResponse{
		Response: &pbCredential.GetResponse_Credentials{
			Credentials: credential,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbCredential.GetListRequest],
	res *connect.ServerStream[pbCredential.GetListResponse],
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
	for rows.Next() {
		credential, err := s.Repo.ScanRow(rows)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}

		if errSend := res.Send(&pbCredential.GetListResponse{
			Response: &pbCredential.GetListResponse_Credentials{
				Credentials: credential,
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func streamingErr(res *connect.ServerStream[pbCredential.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbCredential.GetListResponse{
		Response: &pbCredential.GetListResponse_Error{
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
	req *connect.Request[pbCredential.DeleteRequest],
) (*connect.Response[pbCredential.DeleteResponse], error) {
	cancelledStatus := pbKyc.Status_STATUS_CANCELED
	msgUpdate := &pbCredential.UpdateRequest{
		Id:     req.Msg.GetId(),
		Status: &cancelledStatus,
	}
	updatedLiveliness, err := s.Update(ctx, &connect.Request[pbCredential.UpdateRequest]{
		Msg: msgUpdate,
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbCredential.DeleteResponse{
			Response: &pbCredential.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    updatedLiveliness.Msg.GetError().Code,
					Package: _package,
					Text:    updatedLiveliness.Msg.GetError().Text,
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbCredential.DeleteResponse{
		Response: &pbCredential.DeleteResponse_Credentials{
			Credentials: updatedLiveliness.Msg.GetCredentials(),
		},
	}), nil
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) MakeCreationQB(msg *pbCredential.CreateRequest) (*util.QueryBuilder, *common.ErrWithCode) {
	qb, err := s.Repo.QbInsert(msg)
	if err != nil {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"creating",
			_entityName,
			err.Error(),
		)
	}
	return qb, nil
}

func (s *ServiceServer) MakeUpdateQB(msg *pbCredential.UpdateRequest, upsert bool) (*util.QueryBuilder, *common.ErrWithCode) {
	oldCredential, getOldCredentialErr := s.getOldCredentialToUpdate(msg)
	if getOldCredentialErr != nil && !upsert {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating",
			_entityName,
			getOldCredentialErr.Error(),
		)
	}
	var (
		qb    *util.QueryBuilder
		qbErr error
	)
	if upsert && oldCredential == nil {
		qb, qbErr = s.Repo.QbInsert(&pbCredential.CreateRequest{
			Photo:                msg.Photo,
			Gender:               msg.Gender,
			Title:                msg.Title,
			FirstName:            msg.FirstName,
			MiddleNames:          msg.MiddleNames,
			LastName:             msg.LastName,
			Birthday:             msg.Birthday,
			CountryOfBirth:       msg.CountryOfBirth,
			CountryOfNationality: msg.CountryOfNationality,
			Status:               msg.GetStatus().Enum(),
		})
	} else {
		qb, qbErr = s.Repo.QbUpdate(msg)
	}

	if qbErr != nil {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating/upserting",
			_entityName,
			qbErr.Error(),
		)
	}
	return qb, nil
}
