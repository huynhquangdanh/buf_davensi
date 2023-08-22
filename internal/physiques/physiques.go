package physiques

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/util"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbKyc "davensi.com/core/gen/kyc"
	pbPhysiques "davensi.com/core/gen/physiques"
	pbPhysiquesConnect "davensi.com/core/gen/physiques/physiquesconnect"
	"davensi.com/core/internal/common"
)

const (
	_package          = "physiques"
	_tableName        = "core.kyc_physiques"
	_entityName       = "KYC Physique"
	_entityNamePlural = "KYC Physiques"
	_fields           = "id, race, ethnicity, body_shape, height, weight, status, eyes_color, hair_color"
)

// ServiceServer implements the PhysiquesService API
type ServiceServer struct {
	Repo PhysiqueRepository
	pbPhysiquesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

// For singleton Physiques export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewPhysiqueRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbPhysiques.CreateRequest],
) (*connect.Response[pbPhysiques.CreateResponse], error) {
	qb, err := s.Repo.QbInsert(req.Msg)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbPhysiques.CreateResponse{
			Response: &pbPhysiques.CreateResponse_Error{
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
	newPhysique, err := common.ExecuteTxWrite[pbPhysiques.Physique](
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
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbPhysiques.CreateResponse{
			Response: &pbPhysiques.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s created successfully with id = %s",
		_entityName, newPhysique.Id)
	return connect.NewResponse(&pbPhysiques.CreateResponse{
		Response: &pbPhysiques.CreateResponse_Physique{
			Physique: newPhysique,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbPhysiques.UpdateRequest],
) (*connect.Response[pbPhysiques.UpdateResponse], error) {
	if errQueryUpdate := validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errQueryUpdate.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbPhysiques.UpdateResponse{
			Response: &pbPhysiques.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	getPhysiqueResponse, err := s.getOldPhysiqueToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbPhysiques.UpdateResponse{
			Response: &pbPhysiques.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getPhysiqueResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getPhysiqueResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbPhysiques.UpdateResponse{
			Response: &pbPhysiques.UpdateResponse_Error{
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
	updatedPhysique, err := common.ExecuteTxWrite[pbPhysiques.Physique](
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
		return connect.NewResponse(&pbPhysiques.UpdateResponse{
			Response: &pbPhysiques.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), errUpdate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbPhysiques.UpdateResponse{
		Response: &pbPhysiques.UpdateResponse_Physique{
			Physique: updatedPhysique,
		},
	}), nil
}

func (s *ServiceServer) getOldPhysiqueToUpdate(msg *pbPhysiques.UpdateRequest) (*connect.Response[pbPhysiques.GetResponse], error) {
	getPhysiqueRequest := &pbPhysiques.GetRequest{Id: msg.GetId()}

	getPhysiqueRes, err := s.Get(context.Background(), &connect.Request[pbPhysiques.GetRequest]{
		Msg: getPhysiqueRequest,
	})
	if err != nil {
		return nil, err
	}

	return getPhysiqueRes, nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbPhysiques.GetRequest],
) (*connect.Response[pbPhysiques.GetResponse], error) {
	if errQueryGet := validateQueryGet(req.Msg); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbPhysiques.GetResponse{
			Response: &pbPhysiques.GetResponse_Error{
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
		return connect.NewResponse(&pbPhysiques.GetResponse{
			Response: &pbPhysiques.GetResponse_Error{
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
		return connect.NewResponse(&pbPhysiques.GetResponse{
			Response: &pbPhysiques.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	physique, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbPhysiques.GetResponse{
			Response: &pbPhysiques.GetResponse_Error{
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
		return connect.NewResponse(&pbPhysiques.GetResponse{
			Response: &pbPhysiques.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbPhysiques.GetResponse{
		Response: &pbPhysiques.GetResponse_Physique{
			Physique: physique,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbPhysiques.GetListRequest],
	res *connect.ServerStream[pbPhysiques.GetListResponse],
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
				return res.Send(&pbPhysiques.GetListResponse{
					Response: &pbPhysiques.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		physique, err := s.Repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbPhysiques.GetListResponse{
						Response: &pbPhysiques.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbPhysiques.GetListResponse{
			Response: &pbPhysiques.GetListResponse_Physique{
				Physique: physique,
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
	req *connect.Request[pbPhysiques.DeleteRequest],
) (*connect.Response[pbPhysiques.DeleteResponse], error) {
	canceledStatus := pbKyc.Status_STATUS_CANCELED
	physique := &pbPhysiques.UpdateRequest{
		Id:     req.Msg.GetId(),
		Status: &canceledStatus,
	}
	deleteReq := connect.NewRequest(physique)
	deletedPhysique, err := s.Update(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbPhysiques.DeleteResponse{
			Response: &pbPhysiques.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedPhysique.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedPhysique.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbPhysiques.DeleteResponse{
		Response: &pbPhysiques.DeleteResponse_Physique{
			Physique: deletedPhysique.Msg.GetPhysique(),
		},
	}), nil
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) MakeCreationQB(msg *pbPhysiques.CreateRequest) (*util.QueryBuilder, *common.ErrWithCode) {
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

func (s *ServiceServer) MakeUpdateQB(msg *pbPhysiques.UpdateRequest, upsert bool) (*util.QueryBuilder, *common.ErrWithCode) {
	oldPhysique, getOldPhysiqueErr := s.getOldPhysiqueToUpdate(msg)
	if getOldPhysiqueErr != nil && !upsert {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating",
			_entityName,
			getOldPhysiqueErr.Error(),
		)
	}
	var (
		qb    *util.QueryBuilder
		qbErr error
	)
	if upsert && oldPhysique == nil {
		qb, qbErr = s.Repo.QbInsert(&pbPhysiques.CreateRequest{
			Race:      msg.Race,
			Ethnicity: msg.Ethnicity,
			EyesColor: msg.EyesColor,
			HairColor: msg.HairColor,
			BodyShape: msg.BodyShape,
			Height:    msg.Height,
			Weight:    msg.Weight,
			Status:    msg.GetStatus().Enum(),
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
