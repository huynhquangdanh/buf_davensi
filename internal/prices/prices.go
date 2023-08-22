package prices

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
	pbPrices "davensi.com/core/gen/prices"
	pbPricesConnect "davensi.com/core/gen/prices/pricesconnect"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/datasources"
	"davensi.com/core/internal/markets"
)

const (
	_package          = "prices"
	_tableName        = "core.prices"
	_entityName       = "Price"
	_entityNamePlural = "Prices"
	PriceFields       = "id, source_id, market_id, type, timestamp, price, status"
)

// ServiceServer implements the PricesService API
type ServiceServer struct {
	Repo PriceRepository
	pbPricesConnect.UnimplementedServiceHandler
	db            *pgxpool.Pool
	dataSourcesSS *datasources.ServiceServer
	marketsSS     *markets.ServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:          *NewPriceRepository(db),
		db:            db,
		dataSourcesSS: datasources.GetSingletonServiceServer(db),
		marketsSS:     markets.GetSingletonServiceServer(db),
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbPrices.CreateRequest],
) (*connect.Response[pbPrices.CreateResponse], error) {
	commonErr := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"creating",
		_entityName,
		"",
	)
	if validateErr := s.validateCreate(req.Msg); validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbPrices.CreateResponse{
			Response: &pbPrices.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	qb, errGenInsert := s.Repo.QbInsert(req.Msg)
	if errGenInsert != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errGenInsert.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.CreateResponse{
			Response: &pbPrices.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	sqlStr, args, _ := qb.SetReturnFields("*").GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	newPrice, errExcute := common.ExecuteTxWrite(ctx, s.db, sqlStr, args, s.Repo.ScanMainEntity)
	if errExcute != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errExcute.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.CreateResponse{
			Response: &pbPrices.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	log.Info().Msgf("%s created successfully with id = %s",
		_entityName, newPrice.Id)
	return connect.NewResponse(&pbPrices.CreateResponse{
		Response: &pbPrices.CreateResponse_Price{
			Price: newPrice,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbPrices.UpdateRequest],
) (*connect.Response[pbPrices.UpdateResponse], error) {
	commonErr := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"updating",
		_entityName,
		"",
	)

	if errQueryUpdate := s.validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbPrices.UpdateResponse{
			Response: &pbPrices.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getPriceResponse, errGetOldPrice := s.GetOneMainEntity(context.Background(), &connect.Request[pbPrices.GetRequest]{
		Msg: &pbPrices.GetRequest{
			Select: req.Msg.Select,
		},
	})

	if errGetOldPrice != nil {
		log.Error().Err(errGetOldPrice)
		return connect.NewResponse(&pbPrices.UpdateResponse{
			Response: &pbPrices.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getPriceResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getPriceResponse.Msg.GetError().Text,
				},
			},
		}), errGetOldPrice
	}

	priceBeforeUpdate := getPriceResponse.Msg.GetPrice()

	if errValidateMsgUpdate := s.validateUpdateValue(req.Msg, priceBeforeUpdate); errValidateMsgUpdate != nil {
		log.Error().Err(errValidateMsgUpdate.Err)
		return connect.NewResponse(&pbPrices.UpdateResponse{
			Response: &pbPrices.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errValidateMsgUpdate.Code,
					Package: _package,
					Text:    errValidateMsgUpdate.Err.Error(),
				},
			},
		}), errValidateMsgUpdate.Err
	}

	qb, errGenUpdate := s.Repo.QbUpdate(req.Msg)
	if errGenUpdate != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errGenUpdate.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.UpdateResponse{
			Response: &pbPrices.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	updatedPrice, errExcute := common.ExecuteTxWrite(
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.Repo.ScanMainEntity,
	)
	if errExcute != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errExcute.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.UpdateResponse{
			Response: &pbPrices.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbPrices.UpdateResponse{
		Response: &pbPrices.UpdateResponse_Price{
			Price: updatedPrice,
		},
	}), nil
}

func (s *ServiceServer) GetOneMainEntity(
	ctx context.Context,
	req *connect.Request[pbPrices.GetRequest],
) (*connect.Response[pbPrices.GetResponse], error) {
	commonErr := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"fetching",
		_entityName,
		"",
	)

	if errQueryGet := ValidateSelect(req.Msg.GetSelect(), "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
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

	rows, errQueryRows := s.db.Query(ctx, sqlstr, sqlArgs...)
	if errQueryRows != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errQueryRows.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	defer rows.Close()

	if !rows.Next() {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND).
			UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}
	price, errScanRow := s.Repo.ScanMainEntity(rows)
	if errScanRow != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR).
			UpdateMessage(errScanRow.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}
	if rows.Next() {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND).
			UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbPrices.GetResponse{
		Response: &pbPrices.GetResponse_Price{
			Price: price,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbPrices.GetRequest],
) (*connect.Response[pbPrices.GetResponse], error) {
	commonErr := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"fetching",
		_entityName,
		"",
	)

	if errQueryGet := ValidateSelect(req.Msg.GetSelect(), "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qbDataSource := s.dataSourcesSS.Repo.QbGetList(&pbDataSources.GetListRequest{})
	qbMarkets := s.marketsSS.Repo.QbGetList(&pbMarkets.GetListRequest{})

	filterDSstr, filterDsArgs := qbDataSource.Filters.GenerateSQL()
	filterMarketstr, filterMarketArgs := qbMarkets.Filters.GenerateSQL()

	qb := s.Repo.QbGetOne(req.Msg).
		Join(
			fmt.Sprintf("LEFT JOIN %s ON prices.market_id = markets.id", qbMarkets.TableName),
		).
		Join(
			fmt.Sprintf("LEFT JOIN %s ON prices.source_id = datasources.id", qbDataSource.TableName),
		).
		Select(strings.Join(qbMarkets.SelectFields, ", ")).
		Select(strings.Join(qbDataSource.SelectFields, ", ")).
		Where(filterDSstr, filterDsArgs...).
		Where(filterMarketstr, filterMarketArgs...)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, errQueryRows := s.db.Query(ctx, sqlstr, sqlArgs...)
	if errQueryRows != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(errQueryRows.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	defer rows.Close()

	if !rows.Next() {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND).
			UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}
	price, errScanRow := s.Repo.ScanWithRelationship(rows)
	if errScanRow != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR).
			UpdateMessage(errScanRow.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}
	if rows.Next() {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND).
			UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbPrices.GetResponse{
			Response: &pbPrices.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbPrices.GetResponse{
		Response: &pbPrices.GetResponse_Price{
			Price: price,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbPrices.GetListRequest],
	res *connect.ServerStream[pbPrices.GetListResponse],
) error {
	if req.Msg.Source == nil {
		req.Msg.Source = &pbDataSources.GetListRequest{}
	}
	if req.Msg.Market == nil {
		req.Msg.Market = &pbMarkets.GetListRequest{}
	}

	qbDataSource := s.dataSourcesSS.Repo.QbGetList(req.Msg.Source)
	qbMarkets := s.marketsSS.Repo.QbGetList(req.Msg.Market)
	qbPrices := s.Repo.QbGetList(req.Msg)

	filterDSstr, filterDsArgs := qbDataSource.Filters.GenerateSQL()
	filterMarketstr, filterMarketArgs := qbMarkets.Filters.GenerateSQL()

	qbPrices.
		Join(
			fmt.Sprintf("LEFT JOIN %s ON prices.market_id = markets.id", qbMarkets.TableName),
		).
		Join(
			fmt.Sprintf("LEFT JOIN %s ON prices.source_id = datasources.id", qbDataSource.TableName),
		).
		Select(strings.Join(qbMarkets.SelectFields, ", ")).
		Select(strings.Join(qbDataSource.SelectFields, ", ")).
		Where(filterDSstr, filterDsArgs...).
		Where(filterMarketstr, filterMarketArgs...)

	sqlStr, args, _ := qbPrices.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbPrices.GetListResponse{
					Response: &pbPrices.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		price, err := s.Repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbPrices.GetListResponse{
						Response: &pbPrices.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbPrices.GetListResponse{
			Response: &pbPrices.GetListResponse_Price{
				Price: price,
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
	req *connect.Request[pbPrices.DeleteRequest],
) (*connect.Response[pbPrices.DeleteResponse], error) {
	deletedPrice, err := s.Update(ctx, connect.NewRequest[pbPrices.UpdateRequest](&pbPrices.UpdateRequest{
		Select: req.Msg.Select,
		Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
	}))
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbPrices.DeleteResponse{
			Response: &pbPrices.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedPrice.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedPrice.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbPrices.DeleteResponse{
		Response: &pbPrices.DeleteResponse_Price{
			Price: deletedPrice.Msg.GetPrice(),
		},
	}), nil
}
