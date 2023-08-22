package fiats

import (
	"context"
	"fmt"
	"sync"

	pbCommon "davensi.com/core/gen/common"
	pbFiats "davensi.com/core/gen/fiats"
	pbFiatConnect "davensi.com/core/gen/fiats/fiatsconnect"
	pbUoMs "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/uoms"
	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	_package          = "fiats"
	_entityFiat       = "Fiat"
	_entityFiatPlural = "Fiats"
)

// ServiceServer implements the UoMsService API
type ServiceServer struct {
	pbFiatConnect.ServiceHandler
	db   *pgxpool.Pool
	Repo *FiatRepository
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		db: db,
		Repo: &FiatRepository{
			db: db,
		},
	}
}

// For singleton UoMs export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbUoMs.GetRequest],
) (*connect.Response[pbFiats.GetResponse], error) {
	if errValidateGet := uoms.ValidateQueryGet(req.Msg); errValidateGet != nil {
		log.Error().Err(errValidateGet.Err)
		return connect.NewResponse(&pbFiats.GetResponse{
			Response: &pbFiats.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errValidateGet.Code,
					Package: _package,
					Text:    errValidateGet.Err.Error(),
				},
			},
		}), errValidateGet.Err
	}

	qb := s.Repo.QbGetOne(
		req.Msg,
		uoms.GetSingletonServiceServer(s.db).Repo.QbGetOne(req.Msg),
	)

	sqlStr, args, sel := qb.GenerateSQL()

	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityFiat, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbFiats.GetResponse{
			Response: &pbFiats.GetResponse_Error{
				Error: &pbCommon.Error{
					Code: _errno,
					Text: _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), _err
	}

	defer rows.Close()

	if rows.Next() {
		fiat, err := ScanFiatUom(rows)
		if err != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityFiat, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbFiats.GetResponse{
				Response: &pbFiats.GetResponse_Error{
					Error: &pbCommon.Error{
						Code: _errno,
						Text: _err.Error() + " (" + err.Error() + ")",
					},
				},
			}), _err
		}
		if rows.Next() {
			_errno := pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityFiatPlural, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbFiats.GetResponse{
				Response: &pbFiats.GetResponse_Error{
					Error: &pbCommon.Error{
						Code: _errno,
						Text: _err.Error(),
					},
				},
			}), _err
		}
		return connect.NewResponse(&pbFiats.GetResponse{
			Response: &pbFiats.GetResponse_Fiat{
				Fiat: fiat,
			},
		}), err
	} else {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityFiat, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbFiats.GetResponse{
			Response: &pbFiats.GetResponse_Error{
				Error: &pbCommon.Error{
					Code: _errno,
					Text: _err.Error(),
				},
			},
		}), _err
	}
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbFiats.GetListRequest],
	res *connect.ServerStream[pbFiats.GetListResponse],
) error {
	var getListUoMsRequest *pbUoMs.GetListRequest
	if req.Msg.Uom != nil {
		getListUoMsRequest = req.Msg.GetUom()
	} else {
		getListUoMsRequest = &pbUoMs.GetListRequest{}
	}
	qb := s.Repo.QbGetList(
		req.Msg,
		uoms.GetSingletonServiceServer(s.db).Repo.QbGetList(getListUoMsRequest),
	)

	sqlStr, args, _ := qb.GenerateSQL()

	fmt.Println(sqlStr)

	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(_package, pbCommon.ErrorCode_ERROR_CODE_DB_ERROR, err, func(errStream *pbCommon.Error) error {
			return res.Send(&pbFiats.GetListResponse{
				Response: &pbFiats.GetListResponse_Error{
					Error: errStream,
				},
			})
		})
	}
	defer rows.Close()

	for rows.Next() {
		fiat, err := ScanFiatUom(rows)
		if err != nil {
			return common.StreamError(_package, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR, err, func(errStream *pbCommon.Error) error {
				return res.Send(&pbFiats.GetListResponse{
					Response: &pbFiats.GetListResponse_Error{
						Error: errStream,
					},
				})
			})
		}

		if errSend := res.Send(&pbFiats.GetListResponse{
			Response: &pbFiats.GetListResponse_Fiat{
				Fiat: fiat,
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _package, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbFiats.UpdateRequest],
) (*connect.Response[pbFiats.UpdateResponse], error) {
	uomsServiceServer := uoms.GetSingletonServiceServer(s.db)
	updateUoMReq := req.Msg.GetUom()
	updateUoMReq.Type = pbUoMs.Type_TYPE_FIAT.Enum()

	uomQB, _, uomQBErr := uomsServiceServer.MakeUpdateQB(updateUoMReq)

	if uomQBErr != nil {
		log.Error().Err(uomQBErr.Err)
		return connect.NewResponse(&pbFiats.UpdateResponse{
			Response: &pbFiats.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    uomQBErr.Code,
					Package: _package,
					Text:    uomQBErr.Err.Error(),
				},
			},
		}), uomQBErr.Err
	}

	fiatQB, fiatQBErr := s.Repo.QbUpdate(req.Msg)
	if fiatQBErr != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_package+"'", fiatQBErr.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbFiats.UpdateResponse{
			Response: &pbFiats.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	var updatedFiat *pbFiats.Fiat

	uomSQLStr, uomArgs, uomSel := uomQB.GenerateSQL()
	fiatSQLStr, fiatArgs, fiatSel := fiatQB.GenerateSQL()

	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		uom, updateUomsErr := common.TxWrite[pbUoMs.UoM](
			ctx,
			tx,
			uomSQLStr,
			uomArgs,
			uomsServiceServer.Repo.ScanRow,
		)
		if updateUomsErr != nil {
			return updateUomsErr
		}

		fiat, updateFiatErr := common.TxWrite[pbFiats.Fiat](
			ctx,
			tx,
			fiatSQLStr,
			fiatArgs,
			ScanFiat,
		)

		if updateFiatErr != nil {
			return updateFiatErr
		}

		updatedFiat = fiat
		updatedFiat.Uom = uom

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityFiat,
			uomSel,
			fiatSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbFiats.UpdateResponse{
			Response: &pbFiats.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityFiat, uomSel+fiatSel)
	return connect.NewResponse(&pbFiats.UpdateResponse{
		Response: &pbFiats.UpdateResponse_Fiat{
			Fiat: updatedFiat,
		},
	}), nil
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbFiats.CreateRequest],
) (*connect.Response[pbFiats.CreateResponse], error) {
	uomServiceServer := uoms.GetSingletonServiceServer(s.db)
	uomQB, uomQBErr := uomServiceServer.MakeCreationQB(req.Msg.GetUom())

	if uomQBErr != nil {
		log.Error().Err(uomQBErr.Err)
		return connect.NewResponse(&pbFiats.CreateResponse{
			Response: &pbFiats.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    uomQBErr.Code,
					Package: _package,
					Text:    uomQBErr.Err.Error(),
				},
			},
		}), uomQBErr.Err
	}

	var newFiat *pbFiats.Fiat
	uomSQLStr, uomArgs, uomSel := uomQB.GenerateSQL()

	log.Info().Msg("Executing UOM SQL \"" + uomSQLStr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		uom, insertUomErr := common.TxWrite[pbUoMs.UoM](
			ctx,
			tx,
			uomSQLStr,
			uomArgs,
			uomServiceServer.Repo.ScanRow,
		)

		if insertUomErr != nil {
			return insertUomErr
		}

		fiatQB, fiatQBErr := s.Repo.QbInsert(req.Msg, &pbUoMs.UoM{
			Id: uom.GetId(),
		})

		if fiatQBErr != nil {
			return fiatQBErr
		}

		fiatSQLStr, fiatArgs, _ := fiatQB.GenerateSQL()

		fiat, insertFiatErr := common.TxWrite[pbFiats.Fiat](
			ctx,
			tx,
			fiatSQLStr,
			fiatArgs,
			ScanFiat,
		)

		if insertFiatErr != nil {
			return insertFiatErr
		}

		newFiat = fiat
		newFiat.Uom = uom

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityFiat,
			uomSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbFiats.CreateResponse{
			Response: &pbFiats.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityFiat, uomSel)
	return connect.NewResponse(&pbFiats.CreateResponse{
		Response: &pbFiats.CreateResponse_Fiat{
			Fiat: newFiat,
		},
	}), nil
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbUoMs.DeleteRequest],
) (*connect.Response[pbFiats.DeleteResponse], error) {
	msgUpdate := &pbFiats.UpdateRequest{
		Uom: &pbUoMs.UpdateRequest{
			Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
		},
	}

	switch req.Msg.GetSelect().Select.(type) {
	case *pbUoMs.Select_ById:
		msgUpdate.Uom.Select = &pbUoMs.Select{
			Select: &pbUoMs.Select_ById{
				ById: req.Msg.Select.GetById(),
			},
		}
	case *pbUoMs.Select_ByTypeSymbol:
		msgUpdate.Uom.Select = &pbUoMs.Select{
			Select: &pbUoMs.Select_ByTypeSymbol{
				ByTypeSymbol: req.Msg.Select.GetByTypeSymbol(),
			},
		}
	}

	softDeletedFiat, err := s.Update(ctx, &connect.Request[pbFiats.UpdateRequest]{
		Msg: msgUpdate,
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbFiats.DeleteResponse{
			Response: &pbFiats.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    softDeletedFiat.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + softDeletedFiat.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbFiats.DeleteResponse{
		Response: &pbFiats.DeleteResponse_Fiat{
			Fiat: softDeletedFiat.Msg.GetFiat(),
		},
	}), nil
}
