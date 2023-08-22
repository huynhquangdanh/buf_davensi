package blockchains

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbBlockchains "davensi.com/core/gen/blockchains"
	pbBlockchainsConnect "davensi.com/core/gen/blockchains/blockchainsconnect"
	pbCommon "davensi.com/core/gen/common"
	pbCryptos "davensi.com/core/gen/cryptos"
	pbUoMs "davensi.com/core/gen/uoms"
)

const (
	_package          = "blockchains"
	_entityName       = "blockchain"
	_entityNamePlural = "blockchains"
)

// ServiceServer implements the UoMsService API
type ServiceServer struct {
	pbBlockchainsConnect.UnimplementedServiceHandler
	repo BlockchainRepository
	db   *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo: *NewBlockchainRepository(db),
		db:   db,
	}
}

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

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbBlockchains.CreateRequest],
) (*connect.Response[pbBlockchains.CreateResponse], error) {
	if errCreation := s.validateCreate(req.Msg); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbBlockchains.CreateResponse{
			Response: &pbBlockchains.CreateResponse_Error{
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
		return connect.NewResponse(&pbBlockchains.CreateResponse{
			Response: &pbBlockchains.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	sqlStr, args, _ := qb.SetReturnFields("*").GenerateSQL()
	newBlockchain := &pbBlockchains.Blockchain{}

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		insertBlockchainErr := func() error {
			rows, err := tx.Query(ctx, sqlStr, args...)
			if err != nil {
				return err
			}
			defer rows.Close()

			if !rows.Next() {
				return rows.Err()
			}

			newBlockchain, err = s.repo.ScanRow(rows)
			if err != nil {
				log.Error().Err(err).Msgf("unable to create %s with type/name = '%s'",
					_entityName,
					req.Msg.GetType().Enum().String()+"/"+req.Msg.GetName(),
				)
			}
			return nil
		}()
		if insertBlockchainErr != nil {
			return insertBlockchainErr
		}
		return err
	}); err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_package,
			fmt.Sprintf(
				"type/name = '%s/%s'",
				req.Msg.GetType().Enum().String(),
				req.Msg.GetName(),
			),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbBlockchains.CreateResponse{
			Response: &pbBlockchains.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	log.Info().Msgf("%s with type/name = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetType().Enum().String()+"/"+req.Msg.GetName(), newBlockchain.Id)
	return connect.NewResponse(&pbBlockchains.CreateResponse{
		Response: &pbBlockchains.CreateResponse_Blockchain{
			Blockchain: newBlockchain,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbBlockchains.UpdateRequest],
) (*connect.Response[pbBlockchains.UpdateResponse], error) {
	if errQueryUpdate := validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbBlockchains.UpdateResponse{
			Response: &pbBlockchains.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getBlockchainResponse, err := s.getOldBlockchainToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbBlockchains.UpdateResponse{
			Response: &pbBlockchains.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getBlockchainResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getBlockchainResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	dataSourceBeforeUpdate := getBlockchainResponse.Msg.GetBlockchain()

	if errUpdateValue := s.validateUpdateValue(dataSourceBeforeUpdate, req.Msg); errUpdateValue != nil {
		log.Error().Err(errUpdateValue.Err)
		return connect.NewResponse(&pbBlockchains.UpdateResponse{
			Response: &pbBlockchains.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdateValue.Code,
					Package: _package,
					Text:    errUpdateValue.Err.Error(),
				},
			},
		}), errUpdateValue.Err
	}

	qb, err := s.repo.QbUpdate(req.Msg)
	if err != nil {
		errGenSQL := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			err.Error(),
		)
		log.Error().Err(errGenSQL.Err)
		return connect.NewResponse(&pbBlockchains.UpdateResponse{
			Response: &pbBlockchains.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errGenSQL.Code,
					Package: _package,
					Text:    errGenSQL.Err.Error(),
				},
			},
		}), errGenSQL.Err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	updatedBlockchain, err := common.ExecuteTxWrite[pbBlockchains.Blockchain](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.repo.ScanRow,
	)
	if err != nil {
		errUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errUpdate.Err)
		return connect.NewResponse(&pbBlockchains.UpdateResponse{
			Response: &pbBlockchains.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), errUpdate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbBlockchains.UpdateResponse{
		Response: &pbBlockchains.UpdateResponse_Blockchain{
			Blockchain: updatedBlockchain,
		},
	}), nil
}

func (s *ServiceServer) getOldBlockchainToUpdate(msg *pbBlockchains.UpdateRequest) (*connect.Response[pbBlockchains.GetResponse], error) {
	var getFsproviderRequest *pbBlockchains.GetRequest
	switch msg.GetSelect().Select.(type) {
	case *pbBlockchains.Select_ById:
		getFsproviderRequest = &pbBlockchains.GetRequest{
			Select: &pbBlockchains.Select{
				Select: &pbBlockchains.Select_ById{
					ById: msg.GetSelect().GetById(),
				},
			},
		}
	case *pbBlockchains.Select_ByName:
		getFsproviderRequest = &pbBlockchains.GetRequest{
			Select: &pbBlockchains.Select{
				Select: &pbBlockchains.Select_ByName{
					ByName: msg.GetSelect().GetByName(),
				},
			},
		}
	}

	return s.Get(context.Background(), &connect.Request[pbBlockchains.GetRequest]{
		Msg: getFsproviderRequest,
	})
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbBlockchains.GetRequest],
) (*connect.Response[pbBlockchains.GetResponse], error) {
	if errQueryGet := validateQueryGet(req.Msg); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbBlockchains.GetResponse{
			Response: &pbBlockchains.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qb := s.repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbBlockchains.GetResponse{
			Response: &pbBlockchains.GetResponse_Error{
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
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
			"fetching",
			_package,
			sel,
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbBlockchains.GetResponse{
			Response: &pbBlockchains.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}
	dataSource, err := s.repo.ScanRow(rows)
	if err != nil {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
			"fetching",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbBlockchains.GetResponse{
			Response: &pbBlockchains.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}
	if rows.Next() {
		errScan := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND,
			"fetching",
			_entityNamePlural,
			sel,
		)
		log.Error().Err(errScan.Err)
		return connect.NewResponse(&pbBlockchains.GetResponse{
			Response: &pbBlockchains.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errScan.Code,
					Package: _package,
					Text:    errScan.Err.Error(),
				},
			},
		}), errScan.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbBlockchains.GetResponse{
		Response: &pbBlockchains.GetResponse_Blockchain{
			Blockchain: dataSource,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbBlockchains.GetListRequest],
	res *connect.ServerStream[pbBlockchains.GetListResponse],
) error {
	qb := s.repo.QbGetList(req.Msg)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbBlockchains.GetListResponse{
					Response: &pbBlockchains.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		dataSource, err := s.repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbBlockchains.GetListResponse{
						Response: &pbBlockchains.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbBlockchains.GetListResponse{
			Response: &pbBlockchains.GetListResponse_Blockchain{
				Blockchain: dataSource,
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
	req *connect.Request[pbBlockchains.DeleteRequest],
) (*connect.Response[pbBlockchains.DeleteResponse], error) {
	terminatedStatus := pbCommon.Status_STATUS_TERMINATED
	var dataSource *pbBlockchains.UpdateRequest
	switch req.Msg.GetSelect().Select.(type) {
	case *pbBlockchains.Select_ById:
		dataSource = &pbBlockchains.UpdateRequest{
			Select: &pbBlockchains.Select{
				Select: &pbBlockchains.Select_ById{
					ById: req.Msg.GetSelect().GetById(),
				},
			},
		}

	case *pbBlockchains.Select_ByName:
		dataSource = &pbBlockchains.UpdateRequest{
			Select: &pbBlockchains.Select{
				Select: &pbBlockchains.Select_ByName{
					ByName: req.Msg.GetSelect().GetByName(),
				},
			},
		}
	}
	dataSource.Status = &terminatedStatus
	deletedBlockchain, err := s.Update(ctx, connect.NewRequest(dataSource))
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbBlockchains.DeleteResponse{
			Response: &pbBlockchains.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedBlockchain.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedBlockchain.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbBlockchains.DeleteResponse{
		Response: &pbBlockchains.DeleteResponse_Blockchain{
			Blockchain: deletedBlockchain.Msg.GetBlockchain(),
		},
	}), nil
}

func (s *ServiceServer) SetCryptos(
	ctx context.Context,
	req *connect.Request[pbBlockchains.SetCryptosRequest],
) (*connect.Response[pbBlockchains.SetCryptosResponse], error) {
	errSet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
		"set cryptos",
		_package,
		"",
	)
	if errCreation := s.validateHandleCryptos(req.Msg.Select, req.Msg.Cryptos); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbBlockchains.SetCryptosResponse{
			Response: &pbBlockchains.SetCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	cryptos := s.GetCryptosSelectList(req.Msg.Cryptos)
	blockchainRes, err := s.Get(context.Background(), connect.NewRequest[pbBlockchains.GetRequest](&pbBlockchains.GetRequest{
		Select: req.Msg.Select,
	}))

	if err != nil {
		errSet.UpdateMessage("blockchain is not found")
		log.Error().Err(errSet.Err)
		return connect.NewResponse(&pbBlockchains.SetCryptosResponse{
			Response: &pbBlockchains.SetCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errSet.Code,
					Package: _package,
					Text:    errSet.Err.Error(),
				},
			},
		}), errSet.Err
	}

	if len(cryptos) == 0 {
		errSet.UpdateMessage("cryptos list is not found")
		log.Error().Err(errSet.Err)
		return connect.NewResponse(&pbBlockchains.SetCryptosResponse{
			Response: &pbBlockchains.SetCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errSet.Code,
					Package: _package,
					Text:    errSet.Err.Error(),
				},
			},
		}), errSet.Err
	}

	req.Msg.Select = &pbBlockchains.Select{
		Select: &pbBlockchains.Select_ById{
			ById: blockchainRes.Msg.GetBlockchain().Id,
		},
	}
	req.Msg.Cryptos = &pbUoMs.SelectList{
		List: []*pbUoMs.Select{},
	}

	for _, crypto := range cryptos {
		req.Msg.Cryptos.List = append(req.Msg.Cryptos.List, &pbUoMs.Select{
			Select: &pbUoMs.Select_ById{
				ById: crypto.Uom.Id,
			},
		})
	}

	qb := QbUpsertBlockchainCrypto(
		&pbBlockchains.Blockchain{
			Id: req.Msg.Select.GetById(),
		},
		req.Msg.Cryptos,
	)

	sqlStr, args, _ := qb.SetReturnFields("*").GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sqlStr, args...)
		return err
	}); err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_package,
			err.Error(),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbBlockchains.SetCryptosResponse{
			Response: &pbBlockchains.SetCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	log.Info().Msgf("%s with created successfully",
		_entityName)
	return connect.NewResponse(&pbBlockchains.SetCryptosResponse{
		Response: &pbBlockchains.SetCryptosResponse_Cryptos{
			Cryptos: &pbBlockchains.CryptoList{
				Cryptos: &pbCryptos.List{
					List: cryptos,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) AddCryptos(
	ctx context.Context,
	req *connect.Request[pbBlockchains.AddCryptosRequest],
) (*connect.Response[pbBlockchains.AddCryptosResponse], error) {
	errSet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
		"set cryptos",
		_package,
		"",
	)
	if errCreation := s.validateHandleCryptos(req.Msg.Select, req.Msg.Cryptos); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbBlockchains.AddCryptosResponse{
			Response: &pbBlockchains.AddCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	blockchainRes, err := s.Get(context.Background(), connect.NewRequest[pbBlockchains.GetRequest](&pbBlockchains.GetRequest{
		Select: req.Msg.Select,
	}))

	if err != nil {
		errSet.UpdateMessage("blockchain is not found")
		log.Error().Err(errSet.Err)
		return connect.NewResponse(&pbBlockchains.AddCryptosResponse{
			Response: &pbBlockchains.AddCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errSet.Code,
					Package: _package,
					Text:    errSet.Err.Error(),
				},
			},
		}), errSet.Err
	}

	cryptos := s.GetCryptosSelectList(req.Msg.Cryptos)

	if len(cryptos) == 0 {
		errSet.UpdateMessage("cryptos list is not found")
		log.Error().Err(errSet.Err)
		return connect.NewResponse(&pbBlockchains.AddCryptosResponse{
			Response: &pbBlockchains.AddCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errSet.Code,
					Package: _package,
					Text:    errSet.Err.Error(),
				},
			},
		}), errSet.Err
	}

	req.Msg.Select = &pbBlockchains.Select{
		Select: &pbBlockchains.Select_ById{
			ById: blockchainRes.Msg.GetBlockchain().Id,
		},
	}
	req.Msg.Cryptos = &pbUoMs.SelectList{
		List: []*pbUoMs.Select{},
	}

	for _, crypto := range cryptos {
		req.Msg.Cryptos.List = append(req.Msg.Cryptos.List, &pbUoMs.Select{
			Select: &pbUoMs.Select_ById{
				ById: crypto.Uom.Id,
			},
		})
	}

	qb := QbUpsertBlockchainCrypto(
		&pbBlockchains.Blockchain{
			Id: req.Msg.Select.GetById(),
		},
		req.Msg.Cryptos,
	)

	sqlStr, args, _ := qb.SetReturnFields("*").GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sqlStr, args...)
		return err
	}); err != nil {
		log.Error().Err(errSet.UpdateMessage(err.Error()).Err)
		return connect.NewResponse(&pbBlockchains.AddCryptosResponse{
			Response: &pbBlockchains.AddCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errSet.Code,
					Package: _package,
					Text:    errSet.Err.Error(),
				},
			},
		}), errSet.Err
	}

	log.Info().Msgf("%s with created successfully",
		_entityName)
	return connect.NewResponse(&pbBlockchains.AddCryptosResponse{
		Response: &pbBlockchains.AddCryptosResponse_Cryptos{
			Cryptos: &pbBlockchains.CryptoList{
				Cryptos: &pbCryptos.List{
					List: cryptos,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) RemoveCryptos(
	ctx context.Context,
	req *connect.Request[pbBlockchains.RemoveCryptosRequest],
) (*connect.Response[pbBlockchains.RemoveCryptosResponse], error) {
	errRemove := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
		"remove cryptos",
		_package,
		"",
	)
	if errCreation := s.validateHandleCryptos(req.Msg.Select, req.Msg.Cryptos); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbBlockchains.RemoveCryptosResponse{
			Response: &pbBlockchains.RemoveCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	blockchainRes, err := s.Get(context.Background(), connect.NewRequest[pbBlockchains.GetRequest](&pbBlockchains.GetRequest{
		Select: req.Msg.Select,
	}))

	cryptos := s.GetCryptosSelectList(req.Msg.Cryptos)

	if len(cryptos) == 0 {
		errRemove.UpdateMessage("cryptos list is not found")
		log.Error().Err(errRemove.Err)
		return connect.NewResponse(&pbBlockchains.RemoveCryptosResponse{
			Response: &pbBlockchains.RemoveCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errRemove.Code,
					Package: _package,
					Text:    errRemove.Err.Error(),
				},
			},
		}), errRemove.Err
	}

	if err != nil {
		errRemove.UpdateMessage("blockchain is not found")
		log.Error().Err(errRemove.Err)
		return connect.NewResponse(&pbBlockchains.RemoveCryptosResponse{
			Response: &pbBlockchains.RemoveCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    errRemove.Code,
					Package: _package,
					Text:    errRemove.Err.Error(),
				},
			},
		}), errRemove.Err
	}

	req.Msg.Cryptos = &pbUoMs.SelectList{
		List: []*pbUoMs.Select{},
	}

	req.Msg.Select = &pbBlockchains.Select{
		Select: &pbBlockchains.Select_ById{
			ById: blockchainRes.Msg.GetBlockchain().Id,
		},
	}

	for _, crypto := range cryptos {
		req.Msg.Cryptos.List = append(req.Msg.Cryptos.List, &pbUoMs.Select{
			Select: &pbUoMs.Select_ById{
				ById: crypto.Uom.Id,
			},
		})
	}

	qb := QbSoftRemoveBlockchainCrypto(
		&pbBlockchains.Blockchain{
			Id: req.Msg.Select.GetById(),
		},
		req.Msg.Cryptos,
	)

	sqlStr, args, _ := qb.SetReturnFields("*").GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sqlStr, args...)
		return err
	}); err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbBlockchains.RemoveCryptosResponse{
			Response: &pbBlockchains.RemoveCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), errRemove.Err
	}

	log.Info().Msgf("%s with removed successfully",
		_entityName)
	return connect.NewResponse(&pbBlockchains.RemoveCryptosResponse{
		Response: &pbBlockchains.RemoveCryptosResponse_Cryptos{
			Cryptos: &pbBlockchains.CryptoList{
				Cryptos: &pbCryptos.List{
					List: cryptos,
				},
			},
		},
	}), nil
}
