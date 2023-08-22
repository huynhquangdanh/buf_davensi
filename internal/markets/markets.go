package markets

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbMarkets "davensi.com/core/gen/markets"
	pbMarketsConnect "davensi.com/core/gen/markets/marketsconnect"
	pbTradingPairs "davensi.com/core/gen/tradingpairs"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/tradingpairs"
)

const (
	_package          = "markets"
	_entityName       = "Market"
	_entityNamePlural = "Markets"
)

// ServiceServer implements the MarketsService API
type ServiceServer struct {
	Repo MarketRepository
	pbMarketsConnect.UnimplementedServiceHandler
	db            *pgxpool.Pool
	tradingpairSS *tradingpairs.ServiceServer
}

// For singleton UoMs export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:          *NewMarketRepository(db),
		db:            db,
		tradingpairSS: tradingpairs.GetSingletonServiceServer(db),
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
	req *connect.Request[pbMarkets.CreateRequest],
) (*connect.Response[pbMarkets.CreateResponse], error) {
	if validateErr := s.validateCreate(req.Msg); validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbMarkets.CreateResponse{
			Response: &pbMarkets.CreateResponse_Error{
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
		return connect.NewResponse(&pbMarkets.CreateResponse{
			Response: &pbMarkets.CreateResponse_Error{
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
	newMarket, err := common.ExecuteTxWrite[pbMarkets.Market](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanMainEntity,
	)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
			"symbol = '"+req.Msg.GetSymbol()+"'",
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbMarkets.CreateResponse{
			Response: &pbMarkets.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with symbol = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetSymbol(), newMarket.GetId())
	return connect.NewResponse(&pbMarkets.CreateResponse{
		Response: &pbMarkets.CreateResponse_Market{
			Market: newMarket,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbMarkets.UpdateRequest],
) (*connect.Response[pbMarkets.UpdateResponse], error) {
	if errQueryUpdate := ValidateSelect(req.Msg.Select, "updating"); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbMarkets.UpdateResponse{
			Response: &pbMarkets.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getMarketResponse, err := s.Get(context.Background(), &connect.Request[pbMarkets.GetRequest]{
		Msg: &pbMarkets.GetRequest{
			Select: req.Msg.Select,
		},
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbMarkets.UpdateResponse{
			Response: &pbMarkets.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getMarketResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getMarketResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	if errUpdateValue := s.validateUpdateValue(getMarketResponse.Msg.GetMarket(), req.Msg); errUpdateValue != nil {
		log.Error().Err(errUpdateValue.Err)
		return connect.NewResponse(&pbMarkets.UpdateResponse{
			Response: &pbMarkets.UpdateResponse_Error{
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
		return connect.NewResponse(&pbMarkets.UpdateResponse{
			Response: &pbMarkets.UpdateResponse_Error{
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
	updatedMarket, updateErr := common.ExecuteTxWrite[pbMarkets.Market](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.Repo.ScanMainEntity,
	)
	if updateErr != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			sel,
		)
		log.Error().Err(updateErr).Msg(_err.Error())
		return connect.NewResponse(&pbMarkets.UpdateResponse{
			Response: &pbMarkets.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + updateErr.Error() + ")",
				},
			},
		}), updateErr
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbMarkets.UpdateResponse{
		Response: &pbMarkets.UpdateResponse_Market{
			Market: updatedMarket,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbMarkets.GetRequest],
) (*connect.Response[pbMarkets.GetResponse], error) {
	if validateErr := ValidateSelect(req.Msg.Select, "fetching"); validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbMarkets.GetResponse{
			Response: &pbMarkets.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	qbTradingPairs := s.tradingpairSS.Repo.QbGetList(&pbTradingPairs.GetListRequest{})
	tradingPairFB, tradingPairArgs := qbTradingPairs.Filters.GenerateSQL()

	qb := s.Repo.QbGetOne(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON %s.id = markets.tradingpair_id", qbTradingPairs.TableName, qbTradingPairs.TableName)).
		Select(strings.Join(qbTradingPairs.SelectFields, ", ")).
		Where(tradingPairFB, tradingPairArgs...)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbMarkets.GetResponse{
			Response: &pbMarkets.GetResponse_Error{
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
		return connect.NewResponse(&pbMarkets.GetResponse{
			Response: &pbMarkets.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	market, err := s.Repo.ScanWithRelationship(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbMarkets.GetResponse{
			Response: &pbMarkets.GetResponse_Error{
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
		return connect.NewResponse(&pbMarkets.GetResponse{
			Response: &pbMarkets.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbMarkets.GetResponse{
		Response: &pbMarkets.GetResponse_Market{
			Market: market,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbMarkets.GetListRequest],
	res *connect.ServerStream[pbMarkets.GetListResponse],
) error {
	if req.Msg.Tradingpair == nil {
		req.Msg.Tradingpair = &pbTradingPairs.SelectList{}
	}

	qbTradingPairs := s.tradingpairSS.Repo.QbGetBySelect(req.Msg.Tradingpair)
	tradingPairFB, tradingPairArgs := qbTradingPairs.Filters.GenerateSQL()

	qb := s.Repo.QbGetList(req.Msg).
		Join(fmt.Sprintf(
			"LEFT JOIN %s ON %s.id = %s.tradingpair_id",
			qbTradingPairs.TableName,
			qbTradingPairs.TableName,
			_tableName,
		)).
		Select(strings.Join(qbTradingPairs.SelectFields, ", ")).
		Where(tradingPairFB, tradingPairArgs...)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbMarkets.GetListResponse{
					Response: &pbMarkets.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		market, err := s.Repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbMarkets.GetListResponse{
						Response: &pbMarkets.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbMarkets.GetListResponse{
			Response: &pbMarkets.GetListResponse_Market{
				Market: market,
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
	req *connect.Request[pbMarkets.DeleteRequest],
) (*connect.Response[pbMarkets.DeleteResponse], error) {
	deletedMarket, err := s.Update(ctx, connect.NewRequest[pbMarkets.UpdateRequest](&pbMarkets.UpdateRequest{
		Select: req.Msg.Select,
		Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
	}))
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbMarkets.DeleteResponse{
			Response: &pbMarkets.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedMarket.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedMarket.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbMarkets.DeleteResponse{
		Response: &pbMarkets.DeleteResponse_Market{
			Market: deletedMarket.Msg.GetMarket(),
		},
	}), nil
}
