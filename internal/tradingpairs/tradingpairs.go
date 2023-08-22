package tradingpairs

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbTradingPairs "davensi.com/core/gen/tradingpairs"
	pbTradingPairsConnect "davensi.com/core/gen/tradingpairs/tradingpairsconnect"
	pbUoms "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/uoms"
	"davensi.com/core/internal/util"
)

const (
	_package          = "tradingpairs"
	_entityName       = "Trading Pair"
	_entityNamePlural = "Trading Pairs"
)

// ServiceServer implements the TradingPairsService API
type ServiceServer struct {
	Repo TradingPairRepository
	pbTradingPairsConnect.UnimplementedServiceHandler
	db    *pgxpool.Pool
	uomSS *uoms.ServiceServer
}

var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:  *NewTradingPairRepository(db),
		db:    db,
		uomSS: uoms.GetSingletonServiceServer(db),
	}
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) Create(ctx context.Context, req *connect.Request[pbTradingPairs.CreateRequest],
) (*connect.Response[pbTradingPairs.CreateResponse], error) {
	// Verify that Name, Bic, BankCode is specified
	validateErr := s.validateCreate(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbTradingPairs.CreateResponse{
			Response: &pbTradingPairs.CreateResponse_Error{
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
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], err.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbTradingPairs.CreateResponse{
			Response: &pbTradingPairs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	// Query and building response
	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	newTradingPair, err := common.ExecuteTxWrite(ctx, s.db, sqlStr, args, ScanMainEntity)
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_package,
			fmt.Sprintf(
				"symbol = '%s'",
				req.Msg.GetSymbol(),
			),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbTradingPairs.CreateResponse{
			Response: &pbTradingPairs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbTradingPairs.CreateResponse{
		Response: &pbTradingPairs.CreateResponse_Tradingpair{
			Tradingpair: newTradingPair,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context, req *connect.Request[pbTradingPairs.UpdateRequest],
) (*connect.Response[pbTradingPairs.UpdateResponse], error) {
	errQueryUpdate := ValidateSelect(req.Msg.Select, "updating")
	if errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbTradingPairs.UpdateResponse{
			Response: &pbTradingPairs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getTradingPairResponse, err := s.Get(context.Background(), &connect.Request[pbTradingPairs.GetRequest]{
		Msg: &pbTradingPairs.GetRequest{
			Select: req.Msg.Select,
		},
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbTradingPairs.UpdateResponse{
			Response: &pbTradingPairs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    "update failed: could not get old trading pair to update",
				},
			},
		}), err
	}

	tradingPairBeforeUpdate := getTradingPairResponse.Msg.GetTradingpair()

	if errUpdateValue := s.validateUpdateValue(tradingPairBeforeUpdate, req.Msg); errUpdateValue != nil {
		log.Error().Err(errUpdateValue.Err)
		return connect.NewResponse(&pbTradingPairs.UpdateResponse{
			Response: &pbTradingPairs.UpdateResponse_Error{
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
		return connect.NewResponse(&pbTradingPairs.UpdateResponse{
			Response: &pbTradingPairs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	// Executing update and saving response
	log.Info().Msg("Executing SQL '" + sqlstr + "'")
	updatedTradingPair, err := common.ExecuteTxWrite[pbTradingPairs.TradingPair](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		ScanMainEntity,
	)
	if err != nil {
		errUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errUpdate.Err)
		return connect.NewResponse(&pbTradingPairs.UpdateResponse{
			Response: &pbTradingPairs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), errUpdate.Err
	}

	return connect.NewResponse(&pbTradingPairs.UpdateResponse{
		Response: &pbTradingPairs.UpdateResponse_Tradingpair{
			Tradingpair: updatedTradingPair,
		},
	}), nil
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbTradingPairs.DeleteRequest],
) (*connect.Response[pbTradingPairs.DeleteResponse], error) {
	updateRes, err := s.Update(ctx, connect.NewRequest(&pbTradingPairs.UpdateRequest{
		Select: req.Msg.Select,
		Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
	}))
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		log.Error().Err(err).Msg(err.Error())
		return connect.NewResponse(&pbTradingPairs.DeleteResponse{
			Response: &pbTradingPairs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, "id = "+updateRes.Msg.GetTradingpair().Id)
	return connect.NewResponse(&pbTradingPairs.DeleteResponse{
		Response: &pbTradingPairs.DeleteResponse_Tradingpair{
			Tradingpair: updateRes.Msg.GetTradingpair(),
		},
	}), nil
}

func generateJoinSQL(qbTradingpairs, qbPriceUoms, qbQuantityUoms *util.QueryBuilder) *util.QueryBuilder {
	return qbTradingpairs.
		Join(fmt.Sprintf("LEFT JOIN %s %s ON %s.id = %s.price_uom_id", qbPriceUoms.TableName, "puoms", "puoms", _tableName)).
		Join(fmt.Sprintf("LEFT JOIN %s %s ON %s.id = %s.quantity_uom_id", qbQuantityUoms.TableName, "quoms", "quoms", _tableName)).
		Select(strings.Join(
			util.Map(qbPriceUoms.SelectFields,
				func(rowValue string, rowIndex int) string {
					return strings.ReplaceAll(rowValue, "uoms", "puoms")
				}),
			", ")).
		Select(strings.Join(
			util.Map(qbQuantityUoms.SelectFields,
				func(rowValue string, rowIndex int) string {
					return strings.ReplaceAll(rowValue, "uoms", "quoms")
				}),
			", "))
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbTradingPairs.GetRequest],
) (*connect.Response[pbTradingPairs.GetResponse], error) {
	if errQueryGet := ValidateSelect(req.Msg.Select, "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbTradingPairs.GetResponse{
			Response: &pbTradingPairs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qbQuantityUoms := s.uomSS.Repo.QbGetList(&pbUoms.GetListRequest{})
	qbPriceUoms := s.uomSS.Repo.QbGetList(&pbUoms.GetListRequest{})

	qb := generateJoinSQL(s.Repo.QbGetOne(req.Msg), qbPriceUoms, qbQuantityUoms)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	// Query into tradingpairs
	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbTradingPairs.GetResponse{
			Response: &pbTradingPairs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), _err
	}
	defer rows.Close()

	if !rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbTradingPairs.GetResponse{
			Response: &pbTradingPairs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	tradingPair, err := ScanWithRelationship(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbTradingPairs.GetResponse{
			Response: &pbTradingPairs.GetResponse_Error{
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
		return connect.NewResponse(&pbTradingPairs.GetResponse{
			Response: &pbTradingPairs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	rows.Close()

	// Start building the response from here
	return connect.NewResponse(&pbTradingPairs.GetResponse{
		Response: &pbTradingPairs.GetResponse_Tradingpair{
			Tradingpair: tradingPair,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbTradingPairs.GetListRequest],
	res *connect.ServerStream[pbTradingPairs.GetListResponse],
) error {
	if req.Msg.QuantityUom == nil {
		req.Msg.QuantityUom = &pbUoms.SelectList{}
	}
	if req.Msg.PriceUom == nil {
		req.Msg.PriceUom = &pbUoms.SelectList{}
	}

	qbQuantityUoms, qbPriceUoms := s.uomSS.Repo.QbGetBySelect(req.Msg.QuantityUom), s.uomSS.Repo.QbGetBySelect(req.Msg.PriceUom)
	quantityUomFB, quantityUomArgs := qbQuantityUoms.Filters.GenerateSQL()
	priceUomFB, priceUomArgs := qbPriceUoms.Filters.GenerateSQL()

	fmt.Println(priceUomFB)

	qb := generateJoinSQL(s.Repo.QbGetList(req.Msg), qbPriceUoms, qbQuantityUoms).
		Where(strings.ReplaceAll(quantityUomFB, "uoms", "quoms"), quantityUomArgs...).
		Where(strings.ReplaceAll(priceUomFB, "uoms", "puoms"), priceUomArgs...)

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
		tradingPair, err := ScanWithRelationship(rows)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}

		errSend := res.Send(&pbTradingPairs.GetListResponse{
			Response: &pbTradingPairs.GetListResponse_Tradingpair{
				Tradingpair: tradingPair,
			},
		})
		if errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityNamePlural, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
			return _errSend
		}
	}

	if !hasRows { // If there are no match
		if !rows.Next() {
			_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, args)
			log.Error().Err(err).Msg(_err.Error())
			errSend := res.Send(&pbTradingPairs.GetListResponse{
				Response: &pbTradingPairs.GetListResponse_Error{
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
			return _err
		}
	}

	return rows.Err()
}

func streamingErr(res *connect.ServerStream[pbTradingPairs.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbTradingPairs.GetListResponse{
		Response: &pbTradingPairs.GetListResponse_Error{
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
