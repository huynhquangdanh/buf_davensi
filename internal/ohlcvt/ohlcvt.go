package ohlcvt

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbDataSources "davensi.com/core/gen/datasources"
	pbMarkets "davensi.com/core/gen/markets"
	pbOhlcvt "davensi.com/core/gen/ohlcvt"
	pbOhlcvtConnect "davensi.com/core/gen/ohlcvt/ohlcvtconnect"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/datasources"
	"davensi.com/core/internal/markets"
)

const (
	_package          = "ohlcvt"
	_tableName        = "core.ohlcvt"
	_entityName       = "OHLCVT"
	_entityNamePlural = "OHLVCT"
	_fields           = "ohlcvt.id, source_id, market_id, ohlcvt.price_type, timestamp, open, high," +
		"low, close, volume_in_quantity_uom, volume_in_price_uom, trades, ohlcvt.status"
)

// ServiceServer implements the OhlcvtService API
type ServiceServer struct {
	repo OhlcvtRepository
	pbOhlcvtConnect.UnimplementedServiceHandler
	db            *pgxpool.Pool
	srcService    *datasources.ServiceServer
	marketService *markets.ServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo:          *NewOhlcvtRepository(db),
		db:            db,
		srcService:    datasources.GetSingletonServiceServer(db),
		marketService: markets.GetSingletonServiceServer(db),
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbOhlcvt.CreateRequest],
) (*connect.Response[pbOhlcvt.CreateResponse], error) {
	if errCreation := s.validateCreate(req.Msg); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbOhlcvt.CreateResponse{
			Response: &pbOhlcvt.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
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
		return connect.NewResponse(&pbOhlcvt.CreateResponse{
			Response: &pbOhlcvt.CreateResponse_Error{
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
	newOhlcvt, err := common.ExecuteTxWrite[pbOhlcvt.OHLCVT](
		ctx,
		s.db,
		sqlStr,
		args,
		s.repo.ScanMainEntity,
	)
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_package,
			_entityNamePlural,
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbOhlcvt.CreateResponse{
			Response: &pbOhlcvt.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	log.Info().Msgf("%s created successfully with id = %s",
		_entityName, newOhlcvt.Id)
	return connect.NewResponse(&pbOhlcvt.CreateResponse{
		Response: &pbOhlcvt.CreateResponse_Ohlcvt{
			Ohlcvt: newOhlcvt,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbOhlcvt.UpdateRequest],
) (*connect.Response[pbOhlcvt.UpdateResponse], error) {
	if err := s.ValidateSelect(req.Msg.Select, "updating"); err != nil {
		return connect.NewResponse(&pbOhlcvt.UpdateResponse{
			Response: &pbOhlcvt.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    err.Code,
					Package: _package,
					Text:    err.Err.Error(),
				},
			},
		}), err.Err
	}

	if errQueryUpdate := s.validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbOhlcvt.UpdateResponse{
			Response: &pbOhlcvt.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getOhlcvtResponse, err := s.getOldOhlcvtToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbOhlcvt.UpdateResponse{
			Response: &pbOhlcvt.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getOhlcvtResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getOhlcvtResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	qb, genSQLError := s.repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbOhlcvt.UpdateResponse{
			Response: &pbOhlcvt.UpdateResponse_Error{
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
	updatedOhlcvt, err := common.ExecuteTxWrite[pbOhlcvt.OHLCVT](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.repo.ScanMainEntity,
	)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			sel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbOhlcvt.UpdateResponse{
			Response: &pbOhlcvt.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbOhlcvt.UpdateResponse{
		Response: &pbOhlcvt.UpdateResponse_Ohlcvt{
			Ohlcvt: updatedOhlcvt,
		},
	}), nil
}

func (s *ServiceServer) getOldOhlcvtToUpdate(msg *pbOhlcvt.UpdateRequest) (*connect.Response[pbOhlcvt.GetResponse], error) {
	getOhlcvtRequest := &pbOhlcvt.GetRequest{
		Select: &pbOhlcvt.Select{
			Select: &pbOhlcvt.Select_ByOhlcvtKey{
				ByOhlcvtKey: &pbOhlcvt.OHLCVTKey{
					Source:    &pbDataSources.DataSource{Id: msg.GetSource().GetById()},
					Market:    &pbMarkets.Market{Id: msg.GetMarket().GetById()},
					PriceType: msg.GetPriceType(),
					Timestamp: msg.GetTimestamp(),
				},
			},
		},
	}

	getOhlcvtRes, err := s.Get(context.Background(), &connect.Request[pbOhlcvt.GetRequest]{
		Msg: getOhlcvtRequest,
	})
	if err != nil {
		return nil, err
	}

	return getOhlcvtRes, nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbOhlcvt.GetRequest],
) (*connect.Response[pbOhlcvt.GetResponse], error) {
	if errQueryGet := validateSelect(req.Msg); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbOhlcvt.GetResponse{
			Response: &pbOhlcvt.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qbDataSources := s.srcService.Repo.QbGetList(&pbDataSources.GetListRequest{})
	sourceFB, sourceArgs := qbDataSources.Filters.GenerateSQL()
	qbMarkets := s.marketService.Repo.QbGetList(&pbMarkets.GetListRequest{})
	marketFB, marketArgs := qbMarkets.Filters.GenerateSQL()

	qb := s.repo.QbGetOne(req.Msg).
		Join(fmt.Sprintf("JOIN %s ON datasources.id = ohlcvt.source_id JOIN %s ON markets.id = ohlcvt.market_id",
			qbDataSources.TableName,
			qbMarkets.TableName)).
		Select(strings.Join(qbDataSources.SelectFields, ", ")).
		Select(strings.Join(qbMarkets.SelectFields, ", ")).
		Where(sourceFB, sourceArgs...).
		Where(marketFB, marketArgs...)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbOhlcvt.GetResponse{
			Response: &pbOhlcvt.GetResponse_Error{
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
		return connect.NewResponse(&pbOhlcvt.GetResponse{
			Response: &pbOhlcvt.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	ohlcvt, err := s.repo.ScanWithRelationship(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbOhlcvt.GetResponse{
			Response: &pbOhlcvt.GetResponse_Error{
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
		return connect.NewResponse(&pbOhlcvt.GetResponse{
			Response: &pbOhlcvt.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbOhlcvt.GetResponse{
		Response: &pbOhlcvt.GetResponse_Ohlcvt{
			Ohlcvt: ohlcvt,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbOhlcvt.GetListRequest],
	res *connect.ServerStream[pbOhlcvt.GetListResponse],
) error {
	if req.Msg.Source == nil {
		req.Msg.Source = &pbDataSources.GetListRequest{}
	}
	if req.Msg.Market == nil {
		req.Msg.Market = &pbMarkets.GetListRequest{}
	}
	qbDataSources := s.srcService.Repo.QbGetList(req.Msg.Source)
	sourceFB, sourceArgs := qbDataSources.Filters.GenerateSQL()

	qbMarkets := s.marketService.Repo.QbGetList(req.Msg.Market)
	marketFB, marketArgs := qbDataSources.Filters.GenerateSQL()

	qb := s.repo.QbGetList(req.Msg).
		Join(fmt.Sprintf("JOIN %s ON datasources.id = ohlcvt.source_id JOIN %s ON markets.id = ohlcvt.market_id",
			qbDataSources.TableName,
			qbMarkets.TableName)).
		Select(strings.Join(qbDataSources.SelectFields, ", ")).
		Select(strings.Join(qbMarkets.SelectFields, ", ")).
		Where(sourceFB, sourceArgs...).
		Where(marketFB, marketArgs...)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbOhlcvt.GetListResponse{
					Response: &pbOhlcvt.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		ohlcvt, err := s.repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbOhlcvt.GetListResponse{
						Response: &pbOhlcvt.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbOhlcvt.GetListResponse{
			Response: &pbOhlcvt.GetListResponse_Ohlcvt{
				Ohlcvt: ohlcvt,
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
	req *connect.Request[pbOhlcvt.DeleteRequest],
) (*connect.Response[pbOhlcvt.DeleteResponse], error) {
	terminatedStatus := pbCommon.Status_STATUS_TERMINATED
	var ohlcvt pbOhlcvt.UpdateRequest
	switch req.Msg.GetSelect().Select.(type) {
	case *pbOhlcvt.Select_ById:
		ohlcvt.Select = &pbOhlcvt.Select{
			Select: &pbOhlcvt.Select_ById{
				ById: req.Msg.GetSelect().GetById(),
			},
		}
	case *pbOhlcvt.Select_ByOhlcvtKey:
		ohlcvt.Select = &pbOhlcvt.Select{
			Select: &pbOhlcvt.Select_ByOhlcvtKey{
				ByOhlcvtKey: &pbOhlcvt.OHLCVTKey{
					Source:    req.Msg.GetSelect().GetByOhlcvtKey().Source,
					Market:    req.Msg.GetSelect().GetByOhlcvtKey().Market,
					PriceType: req.Msg.GetSelect().GetByOhlcvtKey().PriceType,
					Timestamp: req.Msg.GetSelect().GetByOhlcvtKey().Timestamp,
				},
			},
		}
	}
	ohlcvt.Status = &terminatedStatus
	deletedOhlcvt, err := s.Update(ctx, connect.NewRequest(&ohlcvt))
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbOhlcvt.DeleteResponse{
			Response: &pbOhlcvt.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedOhlcvt.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedOhlcvt.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbOhlcvt.DeleteResponse{
		Response: &pbOhlcvt.DeleteResponse_Ohlcvt{
			Ohlcvt: deletedOhlcvt.Msg.GetOhlcvt(),
		},
	}), nil
}
