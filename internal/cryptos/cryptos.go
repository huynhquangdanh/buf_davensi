package cryptos

import (
	"context"
	"fmt"
	"sync"

	pbCommon "davensi.com/core/gen/common"
	pbCryptos "davensi.com/core/gen/cryptos"
	pbCryptoConnect "davensi.com/core/gen/cryptos/cryptosconnect"
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
	_package            = "cryptos"
	_entityCrypto       = "Crypto"
	_entityCryptoPlural = "Cryptos"
)

// ServiceServer implements the UoMsService API
type ServiceServer struct {
	pbCryptoConnect.ServiceHandler
	db   *pgxpool.Pool
	Repo *CryptoRepository
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		db: db,
		Repo: &CryptoRepository{
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
) (*connect.Response[pbCryptos.GetResponse], error) {
	if errValidateGet := uoms.ValidateQueryGet(req.Msg); errValidateGet != nil {
		log.Error().Err(errValidateGet.Err)
		return connect.NewResponse(&pbCryptos.GetResponse{
			Response: &pbCryptos.GetResponse_Error{
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
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityCrypto, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCryptos.GetResponse{
			Response: &pbCryptos.GetResponse_Error{
				Error: &pbCommon.Error{
					Code: _errno,
					Text: _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), _err
	}

	defer rows.Close()

	if rows.Next() {
		crypto, err := ScanCryptoUom(rows)
		if err != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityCrypto, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbCryptos.GetResponse{
				Response: &pbCryptos.GetResponse_Error{
					Error: &pbCommon.Error{
						Code: _errno,
						Text: _err.Error() + " (" + err.Error() + ")",
					},
				},
			}), _err
		}
		if rows.Next() {
			_errno := pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityCryptoPlural, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbCryptos.GetResponse{
				Response: &pbCryptos.GetResponse_Error{
					Error: &pbCommon.Error{
						Code: _errno,
						Text: _err.Error(),
					},
				},
			}), _err
		}
		return connect.NewResponse(&pbCryptos.GetResponse{
			Response: &pbCryptos.GetResponse_Crypto{
				Crypto: crypto,
			},
		}), err
	} else {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityCrypto, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCryptos.GetResponse{
			Response: &pbCryptos.GetResponse_Error{
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
	req *connect.Request[pbCryptos.GetListRequest],
	res *connect.ServerStream[pbCryptos.GetListResponse],
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

	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(_package, pbCommon.ErrorCode_ERROR_CODE_DB_ERROR, err, func(errStream *pbCommon.Error) error {
			return res.Send(&pbCryptos.GetListResponse{
				Response: &pbCryptos.GetListResponse_Error{
					Error: errStream,
				},
			})
		})
	}
	defer rows.Close()

	for rows.Next() {
		crypto, err := ScanCryptoUom(rows)
		if err != nil {
			return common.StreamError(_package, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR, err, func(errStream *pbCommon.Error) error {
				return res.Send(&pbCryptos.GetListResponse{
					Response: &pbCryptos.GetListResponse_Error{
						Error: errStream,
					},
				})
			})
		}

		if errSend := res.Send(&pbCryptos.GetListResponse{
			Response: &pbCryptos.GetListResponse_Crypto{
				Crypto: crypto,
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
	req *connect.Request[pbCryptos.UpdateRequest],
) (*connect.Response[pbCryptos.UpdateResponse], error) {
	uomsServiceServer := uoms.GetSingletonServiceServer(s.db)
	updateUoMReq := req.Msg.GetUom()
	updateUoMReq.Type = pbUoMs.Type_TYPE_CRYPTO.Enum()

	uomQB, _, uomQBErr := uomsServiceServer.MakeUpdateQB(updateUoMReq)

	if uomQBErr != nil {
		log.Error().Err(uomQBErr.Err)
		return connect.NewResponse(&pbCryptos.UpdateResponse{
			Response: &pbCryptos.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    uomQBErr.Code,
					Package: _package,
					Text:    uomQBErr.Err.Error(),
				},
			},
		}), uomQBErr.Err
	}

	cryptoQB, cryptoQBErr := s.Repo.QbUpdate(req.Msg)
	if cryptoQBErr != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_package+"'", cryptoQBErr.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbCryptos.UpdateResponse{
			Response: &pbCryptos.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	var updatedCrypto *pbCryptos.Crypto

	uomSQLStr, uomArgs, uomSel := uomQB.GenerateSQL()
	cryptoSQLStr, cryptoArgs, cryptoSel := cryptoQB.GenerateSQL()

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

		crypto, updateCryptoErr := common.TxWrite[pbCryptos.Crypto](
			ctx,
			tx,
			cryptoSQLStr,
			cryptoArgs,
			ScanCrypto,
		)

		if updateCryptoErr != nil {
			return updateCryptoErr
		}

		updatedCrypto = crypto
		updatedCrypto.Uom = uom

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityCrypto,
			uomSel,
			cryptoSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCryptos.UpdateResponse{
			Response: &pbCryptos.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityCrypto, uomSel+cryptoSel)
	return connect.NewResponse(&pbCryptos.UpdateResponse{
		Response: &pbCryptos.UpdateResponse_Crypto{
			Crypto: updatedCrypto,
		},
	}), nil
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbCryptos.CreateRequest],
) (*connect.Response[pbCryptos.CreateResponse], error) {
	uomServiceServer := uoms.GetSingletonServiceServer(s.db)
	uomQB, uomQBErr := uomServiceServer.MakeCreationQB(req.Msg.GetUom())

	if uomQBErr != nil {
		log.Error().Err(uomQBErr.Err)
		return connect.NewResponse(&pbCryptos.CreateResponse{
			Response: &pbCryptos.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    uomQBErr.Code,
					Package: _package,
					Text:    uomQBErr.Err.Error(),
				},
			},
		}), uomQBErr.Err
	}

	var newCrypto *pbCryptos.Crypto
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

		cryptoQB, cryptoQBErr := s.Repo.QbInsert(req.Msg, &pbUoMs.UoM{
			Id: uom.GetId(),
		})

		if cryptoQBErr != nil {
			return cryptoQBErr
		}

		cryptoSQLStr, cryptoArgs, _ := cryptoQB.GenerateSQL()

		crypto, insertCryptoErr := common.TxWrite[pbCryptos.Crypto](
			ctx,
			tx,
			cryptoSQLStr,
			cryptoArgs,
			ScanCrypto,
		)

		if insertCryptoErr != nil {
			return insertCryptoErr
		}

		newCrypto = crypto
		newCrypto.Uom = uom

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityCrypto,
			uomSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCryptos.CreateResponse{
			Response: &pbCryptos.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityCrypto, uomSel)
	return connect.NewResponse(&pbCryptos.CreateResponse{
		Response: &pbCryptos.CreateResponse_Crypto{
			Crypto: newCrypto,
		},
	}), nil
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbUoMs.DeleteRequest],
) (*connect.Response[pbCryptos.DeleteResponse], error) {
	msgUpdate := &pbCryptos.UpdateRequest{
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

	softDeletedCrypto, err := s.Update(ctx, &connect.Request[pbCryptos.UpdateRequest]{
		Msg: msgUpdate,
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbCryptos.DeleteResponse{
			Response: &pbCryptos.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    softDeletedCrypto.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + softDeletedCrypto.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbCryptos.DeleteResponse{
		Response: &pbCryptos.DeleteResponse_Crypto{
			Crypto: softDeletedCrypto.Msg.GetCrypto(),
		},
	}), nil
}
