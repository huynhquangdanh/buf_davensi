package livelinesses

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/util"

	"github.com/rs/zerolog/log"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"

	pbCommon "davensi.com/core/gen/common"
	pbKyc "davensi.com/core/gen/kyc"
	pbLivelinesses "davensi.com/core/gen/liveliness"
	livelinessConnect "davensi.com/core/gen/liveliness/livelinessconnect"
	"davensi.com/core/internal/common"
)

// ServiceServer implements the AddressesService API
type ServiceServer struct {
	Repo LivelinessRepository
	livelinessConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

// For singleton Liveliness export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewLivelinessRepository(db),
		db:   db,
	}
}

// Create function ...
func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbLivelinesses.CreateRequest],
) (*connect.Response[pbLivelinesses.CreateResponse], error) {
	if err := s.validateQueryInsert(req.Msg); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			"creating '"+_entityName+"'", err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbLivelinesses.CreateResponse{
			Response: &pbLivelinesses.CreateResponse_Error{
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
		return connect.NewResponse(&pbLivelinesses.CreateResponse{
			Response: &pbLivelinesses.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlStr, args, _ := qb.GenerateSQL()
	var newLiveliness *pbLivelinesses.Liveliness
	log.Info().Str("sqlStr", sqlStr).Msg("the query")

	newLiveliness, err = common.ExecuteTxWrite[pbLivelinesses.Liveliness](
		ctx, s.db, sqlStr, args, s.Repo.ScanRow,
	)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbLivelinesses.CreateResponse{
			Response: &pbLivelinesses.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}
	log.Info().Msgf("%s created successfully with id = %s", _entityName, newLiveliness.GetId())

	return connect.NewResponse(&pbLivelinesses.CreateResponse{
		Response: &pbLivelinesses.CreateResponse_Liveliness{
			Liveliness: newLiveliness,
		},
	}), nil
}

func (s *ServiceServer) getOldLivelinessToUpdate(msg *pbLivelinesses.UpdateRequest) (*connect.Response[pbLivelinesses.GetResponse], error) {
	getLivelinessRequest := &pbLivelinesses.GetRequest{Id: msg.GetId()}

	getLivelinessRes, err := s.Get(context.Background(), &connect.Request[pbLivelinesses.GetRequest]{
		Msg: getLivelinessRequest,
	})
	if err != nil {
		return nil, err
	}

	return getLivelinessRes, nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbLivelinesses.UpdateRequest],
) (*connect.Response[pbLivelinesses.UpdateResponse], error) {
	if errQueryUpdate := s.validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errQueryUpdate.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbLivelinesses.UpdateResponse{
			Response: &pbLivelinesses.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	getResponse, err := s.getOldLivelinessToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		return connect.NewResponse(&pbLivelinesses.UpdateResponse{
			Response: &pbLivelinesses.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	livelinessBeforeUpdate := getResponse.Msg.GetLiveliness()
	_errno, errUpdateValue := s.validateUpdateValue(livelinessBeforeUpdate, req.Msg)
	if errUpdateValue != nil {
		log.Error().Err(errUpdateValue)
		return connect.NewResponse(&pbLivelinesses.UpdateResponse{
			Response: &pbLivelinesses.UpdateResponse_Error{
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
		return connect.NewResponse(&pbLivelinesses.UpdateResponse{
			Response: &pbLivelinesses.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	qb.SetReturnFields("*")

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	var updatedLiveliness *pbLivelinesses.Liveliness

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	updatedLiveliness, err = common.ExecuteTxWrite[pbLivelinesses.Liveliness](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.Repo.ScanRow,
	)

	// Start building the response from here
	if err != nil {
		errUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errUpdate.Err)
		return connect.NewResponse(&pbLivelinesses.UpdateResponse{
			Response: &pbLivelinesses.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbLivelinesses.UpdateResponse{
		Response: &pbLivelinesses.UpdateResponse_Liveliness{
			Liveliness: updatedLiveliness,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbLivelinesses.GetRequest],
) (*connect.Response[pbLivelinesses.GetResponse], error) {
	if err := s.validateMsgGetOne(req.Msg); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", err.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbLivelinesses.GetResponse{
			Response: &pbLivelinesses.GetResponse_Error{
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
		return connect.NewResponse(&pbLivelinesses.GetResponse{
			Response: &pbLivelinesses.GetResponse_Error{
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
		return connect.NewResponse(&pbLivelinesses.GetResponse{
			Response: &pbLivelinesses.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	liveliness, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbLivelinesses.GetResponse{
			Response: &pbLivelinesses.GetResponse_Error{
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
		return connect.NewResponse(&pbLivelinesses.GetResponse{
			Response: &pbLivelinesses.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbLivelinesses.GetResponse{
		Response: &pbLivelinesses.GetResponse_Liveliness{
			Liveliness: liveliness,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbLivelinesses.GetListRequest],
	res *connect.ServerStream[pbLivelinesses.GetListResponse],
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
		liveliness, err := s.Repo.ScanRow(rows)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}

		if errSend := res.Send(&pbLivelinesses.GetListResponse{
			Response: &pbLivelinesses.GetListResponse_Liveliness{
				Liveliness: liveliness,
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func streamingErr(res *connect.ServerStream[pbLivelinesses.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbLivelinesses.GetListResponse{
		Response: &pbLivelinesses.GetListResponse_Error{
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
	req *connect.Request[pbLivelinesses.DeleteRequest],
) (*connect.Response[pbLivelinesses.DeleteResponse], error) {
	if errQueryDelete := s.validateQueryDelete(req.Msg); errQueryDelete != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"deleting '"+_entityName+"'", errQueryDelete.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbLivelinesses.DeleteResponse{
			Response: &pbLivelinesses.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	cancelledStatus := pbKyc.Status_STATUS_CANCELED
	msgUpdate := &pbLivelinesses.UpdateRequest{
		Id:     req.Msg.GetId(),
		Status: &cancelledStatus,
	}
	updatedLiveliness, err := s.Update(ctx, &connect.Request[pbLivelinesses.UpdateRequest]{
		Msg: msgUpdate,
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbLivelinesses.DeleteResponse{
			Response: &pbLivelinesses.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    updatedLiveliness.Msg.GetError().Code,
					Package: _package,
					Text:    updatedLiveliness.Msg.GetError().Text,
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbLivelinesses.DeleteResponse{
		Response: &pbLivelinesses.DeleteResponse_Liveliness{
			Liveliness: updatedLiveliness.Msg.GetLiveliness(),
		},
	}), nil
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) MakeCreationQB(msg *pbLivelinesses.CreateRequest) (*util.QueryBuilder, *common.ErrWithCode) {
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

func (s *ServiceServer) MakeUpdateQB(msg *pbLivelinesses.UpdateRequest, upsert bool) (*util.QueryBuilder, *common.ErrWithCode) {
	oldLiveliness, getOldLivelinessErr := s.getOldLivelinessToUpdate(msg)
	if getOldLivelinessErr != nil && !upsert {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating",
			_entityName,
			getOldLivelinessErr.Error(),
		)
	}
	var (
		qb    *util.QueryBuilder
		qbErr error
	)
	if upsert && oldLiveliness == nil {
		qb, qbErr = s.Repo.QbInsert(&pbLivelinesses.CreateRequest{
			LivelinessVideoFile:      msg.LivelinessVideoFile,
			LivelinessVideoFileType:  msg.LivelinessVideoFileType,
			TimestampVideoFile:       msg.TimestampVideoFile,
			TimestampVideoFileType:   msg.TimestampVideoFileType,
			IdOwnershipPhotoFile:     msg.IdOwnershipPhotoFile,
			IdOwnershipPhotoFileType: msg.IdOwnershipPhotoFileType,
			Status:                   msg.GetStatus().Enum(),
		})
	} else {
		qb, qbErr = s.Repo.QbUpdate(msg)
	}
	if qbErr != nil {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating",
			_entityName,
			qbErr.Error(),
		)
	}
	return qb, nil
}
