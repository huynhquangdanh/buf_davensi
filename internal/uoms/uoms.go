package uoms

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/util"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbUoMs "davensi.com/core/gen/uoms"
	pbUoMsConnect "davensi.com/core/gen/uoms/uomsconnect"
)

const (
	_package          = "uoms"
	_entityName       = "Unit of Measurement"
	_entityNamePlural = "Units of Measurement"
)

// For singleton UoMs export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

// ServiceServer implements the UoMsService API
type ServiceServer struct {
	Repo UoMRepository
	pbUoMsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewUoMRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) MakeCreationQB(msg *pbUoMs.CreateRequest) (*util.QueryBuilder, *common.ErrWithCode) {
	if validateErr := s.validateMsgCreate(msg); validateErr != nil {
		return nil, validateErr
	}

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

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbUoMs.CreateRequest],
) (*connect.Response[pbUoMs.CreateResponse], error) {
	errCreate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"creating",
		_entityName,
		"",
	)

	qb, createQBErr := s.MakeCreationQB(req.Msg)
	if createQBErr != nil {
		log.Error().Err(createQBErr.Err)
		return connect.NewResponse(&pbUoMs.CreateResponse{
			Response: &pbUoMs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    createQBErr.Code,
					Package: _package,
					Text:    createQBErr.Err.Error(),
				},
			},
		}), createQBErr.Err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	newUoM, errExcute := common.ExecuteTxWrite(
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanRow,
	)
	if errExcute != nil {
		errCreate.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errExcute.Error())

		log.Error().Err(errCreate.Err)

		return connect.NewResponse(&pbUoMs.CreateResponse{
			Response: &pbUoMs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreate.Code,
					Package: _package,
					Text:    errCreate.Err.Error(),
				},
			},
		}), errCreate.Err
	}

	log.Info().Msgf("%s with type/symbol = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetType().Enum().String()+"/"+req.Msg.GetSymbol(), newUoM.GetId())
	return connect.NewResponse(&pbUoMs.CreateResponse{
		Response: &pbUoMs.CreateResponse_Uom{
			Uom: newUoM,
		},
	}), nil
}

func (s *ServiceServer) MakeUpdateQB(msg *pbUoMs.UpdateRequest) (*util.QueryBuilder, *pbUoMs.UoM, *common.ErrWithCode) {
	if errQueryUpdate := validateQueryUpdate(msg); errQueryUpdate != nil {
		return nil, nil, errQueryUpdate
	}

	getUomResponse, err := s.getOldUomsToUpdate(msg)
	if err != nil {
		log.Error().Err(err)
		return nil, nil, common.CreateErrWithCode(
			getUomResponse.Msg.GetError().Code,
			"updating",
			_entityName,
			err.Error(),
		)
	}

	uomBeforeUpdate := getUomResponse.Msg.GetUom()

	if errUpdateValue := s.validateMsgUpdate(uomBeforeUpdate, msg); errUpdateValue != nil {
		return nil, nil, errUpdateValue
	}

	qb, qbUpdateErr := s.Repo.QbUpdate(msg)
	if qbUpdateErr != nil {
		return nil, nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating",
			_entityName,
			qbUpdateErr.Error(),
		)
	}
	return qb, uomBeforeUpdate, nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbUoMs.UpdateRequest],
) (*connect.Response[pbUoMs.UpdateResponse], error) {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"updating",
		_entityName,
		"",
	)

	qb, _, updateQBErr := s.MakeUpdateQB(req.Msg)
	if updateQBErr != nil {
		log.Error().Err(updateQBErr.Err)
		return connect.NewResponse(&pbUoMs.UpdateResponse{
			Response: &pbUoMs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    updateQBErr.Code,
					Package: _package,
					Text:    updateQBErr.Err.Error(),
				},
			},
		}), updateQBErr.Err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	updatedUoM, errExcute := common.ExecuteTxWrite(
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.Repo.ScanRow,
	)
	if errExcute != nil {
		errUpdate.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(fmt.Sprintf("%s with error: %s", sel, errExcute.Error()))
		log.Error().Err(errUpdate.Err)

		return connect.NewResponse(&pbUoMs.UpdateResponse{
			Response: &pbUoMs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), errUpdate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbUoMs.UpdateResponse{
		Response: &pbUoMs.UpdateResponse_Uom{
			Uom: updatedUoM,
		},
	}), nil
}

func (s *ServiceServer) getOldUomsToUpdate(msg *pbUoMs.UpdateRequest) (*connect.Response[pbUoMs.GetResponse], error) {
	getUomRequest := &pbUoMs.GetRequest{
		Select: msg.GetSelect(),
	}

	return s.Get(context.Background(), &connect.Request[pbUoMs.GetRequest]{
		Msg: getUomRequest,
	})
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbUoMs.GetRequest],
) (*connect.Response[pbUoMs.GetResponse], error) {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"fetching",
		_entityName,
		"",
	)
	if errValidateGet := ValidateQueryGet(req.Msg); errValidateGet != nil {
		log.Error().Err(errValidateGet.Err)
		return connect.NewResponse(&pbUoMs.GetResponse{
			Response: &pbUoMs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errValidateGet.Code,
					Package: _package,
					Text:    errValidateGet.Err.Error(),
				},
			},
		}), errValidateGet.Err
	}

	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		errGet.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(fmt.Sprintf("%s with error: %s", sel, err.Error()))
		log.Error().Err(errGet.Err)

		return connect.NewResponse(&pbUoMs.GetResponse{
			Response: &pbUoMs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errGet.Code,
					Package: _package,
					Text:    errGet.Err.Error(),
				},
			},
		}), errGet.Err
	}

	defer rows.Close()

	if !rows.Next() {
		errGet.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage("not found")
		log.Error().Err(errGet.Err)
		return connect.NewResponse(&pbUoMs.GetResponse{
			Response: &pbUoMs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errGet.Code,
					Package: _package,
					Text:    errGet.Err.Error(),
				},
			},
		}), errGet.Err
	}
	uom, err := s.Repo.ScanRow(rows)
	if err != nil {
		errGet.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR).
			UpdateMessage(err.Error())
		log.Error().Err(errGet.Err)
		return connect.NewResponse(&pbUoMs.GetResponse{
			Response: &pbUoMs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errGet.Code,
					Package: _package,
					Text:    errGet.Err.Error(),
				},
			},
		}), errGet.Err
	}
	if rows.Next() {
		errGet.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND).
			UpdateMessage("multi value found")
		log.Error().Err(errGet.Err)
		return connect.NewResponse(&pbUoMs.GetResponse{
			Response: &pbUoMs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errGet.Code,
					Package: _package,
					Text:    errGet.Err.Error(),
				},
			},
		}), errGet.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbUoMs.GetResponse{
		Response: &pbUoMs.GetResponse_Uom{
			Uom: uom,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbUoMs.GetListRequest],
	res *connect.ServerStream[pbUoMs.GetListResponse],
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
				return res.Send(&pbUoMs.GetListResponse{
					Response: &pbUoMs.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		uom, err := s.Repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbUoMs.GetListResponse{
						Response: &pbUoMs.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbUoMs.GetListResponse{
			Response: &pbUoMs.GetListResponse_Uom{
				Uom: uom,
			},
		}); errSend != nil {
			log.Error().Err(common.CreateErrWithCode(
				pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR,
				"fetching",
				_entityNamePlural,
				errSend.Error(),
			).Err)
		}
	}

	return rows.Err()
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbUoMs.DeleteRequest],
) (*connect.Response[pbUoMs.DeleteResponse], error) {
	msgUpdate := &pbUoMs.UpdateRequest{
		Select: req.Msg.GetSelect(),
		Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
	}

	updatedUom, err := s.Update(ctx, &connect.Request[pbUoMs.UpdateRequest]{
		Msg: msgUpdate,
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUoMs.DeleteResponse{
			Response: &pbUoMs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    updatedUom.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + updatedUom.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbUoMs.DeleteResponse{
		Response: &pbUoMs.DeleteResponse_Uom{
			Uom: updatedUom.Msg.GetUom(),
		},
	}), nil
}
