package defiwallets

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/blockchains"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/recipients"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbDefiwallets "davensi.com/core/gen/defiwallets"
	pbDefiwalletsConnect "davensi.com/core/gen/defiwallets/defiwalletsconnect"
	pbRecipients "davensi.com/core/gen/recipients"
)

const (
	_package          = "defiwallets"
	_entityName       = "Defi Wallet"
	_entityNamePlural = "Defi Wallets"
)

// ServiceServer implements the DefiwalletsService API
type ServiceServer struct {
	Repo DefiWalletRepository
	pbDefiwalletsConnect.UnimplementedServiceHandler
	db           *pgxpool.Pool
	blockchainSS *blockchains.ServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:         *NewDefiWalletRepository(db),
		blockchainSS: blockchains.GetSingletonServiceServer(db),
		db:           db,
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
	req *connect.Request[pbDefiwallets.CreateRequest],
) (*connect.Response[pbDefiwallets.CreateResponse], error) {
	if errCreation := s.validateCreate(req.Msg); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbDefiwallets.CreateResponse{
			Response: &pbDefiwallets.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	if err := s.validateQueryInsert(req.Msg); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			"creating '"+_entityName+"'", err.Err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbDefiwallets.CreateResponse{
			Response: &pbDefiwallets.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	recipientServiceServer := recipients.GetSingletonServiceServer(s.db)
	recipientQB, recipientQBErr := recipientServiceServer.MakeCreationQB(req.Msg.GetRecipient())
	if recipientQBErr != nil {
		log.Error().Err(recipientQBErr.Err)
		return connect.NewResponse(&pbDefiwallets.CreateResponse{
			Response: &pbDefiwallets.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    recipientQBErr.Code,
					Package: _package,
					Text:    recipientQBErr.Err.Error(),
				},
			},
		}), recipientQBErr.Err
	}
	var newDefiwallet *pbDefiwallets.DeFiWallet
	recipientSQLStr, recipientArgs, recipientSel := recipientQB.GenerateSQL()
	log.Info().Msg("Executing Recipient SQL \"" + recipientSQLStr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		recipient, insertRecipientErr := common.TxWrite[pbRecipients.Recipient](
			ctx,
			tx,
			recipientSQLStr,
			recipientArgs,
			recipientServiceServer.Repo.ScanRow,
		)

		if insertRecipientErr != nil {
			return insertRecipientErr
		}

		defiwalletQB, defiwalletQBErr := s.Repo.QbInsert(req.Msg, &pbRecipients.Recipient{
			Id: recipient.GetId(),
		})

		if defiwalletQBErr != nil {
			return defiwalletQBErr
		}

		defiwalletSQLStr, defiwalletArgs, _ := defiwalletQB.GenerateSQL()

		defiwallet, insertDefiwalletErr := common.TxWrite[pbDefiwallets.DeFiWallet](
			ctx,
			tx,
			defiwalletSQLStr,
			defiwalletArgs,
			s.Repo.ScanGetRow,
		)
		if insertDefiwalletErr != nil {
			return insertDefiwalletErr
		}

		newDefiwallet = defiwallet
		newDefiwallet.Recipient = recipient

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
			recipientSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDefiwallets.CreateResponse{
			Response: &pbDefiwallets.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}
	log.Info().Msgf("%s with %s created successfully", _entityName, recipientSel)
	return connect.NewResponse(&pbDefiwallets.CreateResponse{
		Response: &pbDefiwallets.CreateResponse_Defiwallet{
			Defiwallet: newDefiwallet,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbDefiwallets.UpdateRequest],
) (*connect.Response[pbDefiwallets.UpdateResponse], error) {
	if errQueryUpdate := s.validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbDefiwallets.UpdateResponse{
			Response: &pbDefiwallets.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	recipientServiceServer := recipients.GetSingletonServiceServer(s.db)
	updateRecipientReq := req.Msg.GetRecipient()
	updateRecipientReq.Type = pbRecipients.Type_TYPE_DV_BOT.Enum()

	recipientQB, recipientQBErr := recipientServiceServer.MakeUpdateQB(updateRecipientReq)

	if recipientQBErr != nil {
		log.Error().Err(recipientQBErr.Err)
		return connect.NewResponse(&pbDefiwallets.UpdateResponse{
			Response: &pbDefiwallets.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    recipientQBErr.Code,
					Package: _package,
					Text:    recipientQBErr.Err.Error(),
				},
			},
		}), recipientQBErr.Err
	}

	defiwalletQB, defiwalletQBErr := s.Repo.QbUpdate(req.Msg)
	if defiwalletQBErr != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_package+"'", defiwalletQBErr.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbDefiwallets.UpdateResponse{
			Response: &pbDefiwallets.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	var updatedDefiwallet *pbDefiwallets.DeFiWallet

	recipientSQLStr, recipientArgs, recipientSel := recipientQB.GenerateSQL()
	defiwalletQBStr, defiwalletArgs, defiwalletSel := defiwalletQB.GenerateSQL()

	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		recipient, updateRecipientErr := common.TxWrite[pbRecipients.Recipient](
			ctx,
			tx,
			recipientSQLStr,
			recipientArgs,
			recipientServiceServer.Repo.ScanRow,
		)
		if updateRecipientErr != nil {
			return updateRecipientErr
		}

		defiwallet, updateDefiwalletErr := common.TxWrite[pbDefiwallets.DeFiWallet](
			ctx,
			tx,
			defiwalletQBStr,
			defiwalletArgs,
			s.Repo.ScanUpdateRow,
		)

		if updateDefiwalletErr != nil {
			return updateDefiwalletErr
		}

		updatedDefiwallet = defiwallet
		updatedDefiwallet.Recipient = recipient

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			recipientSel,
			defiwalletSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDefiwallets.UpdateResponse{
			Response: &pbDefiwallets.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, recipientSel+defiwalletSel)
	return connect.NewResponse(&pbDefiwallets.UpdateResponse{
		Response: &pbDefiwallets.UpdateResponse_Defiwallet{
			Defiwallet: updatedDefiwallet,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbRecipients.GetRequest],
) (*connect.Response[pbDefiwallets.GetResponse], error) {
	recipientServer := recipients.GetSingletonServiceServer(s.db)
	pkRes := recipientServer.GetRecipientRelationshipIds(
		req.Msg.GetSelect().GetByLegalEntityUserLabel().GetUser(),
		req.Msg.GetSelect().GetByLegalEntityUserLabel().GetLegalEntity(),
		nil,
	)

	qb := s.Repo.QbGetOne(
		req.Msg,
		recipientServer.Repo.QbGetOne(req.Msg, pkRes, false, true),
	)

	sqlStr, args, sel := qb.GenerateSQL()

	rows, err := s.db.Query(ctx, sqlStr, args...)
	log.Info().Msg(sqlStr)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDefiwallets.GetResponse{
			Response: &pbDefiwallets.GetResponse_Error{
				Error: &pbCommon.Error{
					Code: _errno,
					Text: _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), _err
	}

	defer rows.Close()

	if rows.Next() {
		dvbots, err := s.Repo.ScanGetRow(rows)
		if err != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbDefiwallets.GetResponse{
				Response: &pbDefiwallets.GetResponse_Error{
					Error: &pbCommon.Error{
						Code: _errno,
						Text: _err.Error() + " (" + err.Error() + ")",
					},
				},
			}), _err
		}
		if rows.Next() {
			_errno := pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbDefiwallets.GetResponse{
				Response: &pbDefiwallets.GetResponse_Error{
					Error: &pbCommon.Error{
						Code: _errno,
						Text: _err.Error(),
					},
				},
			}), _err
		}
		return connect.NewResponse(&pbDefiwallets.GetResponse{
			Response: &pbDefiwallets.GetResponse_Defiwallet{
				Defiwallet: dvbots,
			},
		}), err
	} else {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDefiwallets.GetResponse{
			Response: &pbDefiwallets.GetResponse_Error{
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
	req *connect.Request[pbDefiwallets.GetListRequest],
	res *connect.ServerStream[pbDefiwallets.GetListResponse],
) error {
	var getListRecipientsRequest *pbRecipients.GetListRequest
	if req.Msg.Recipient != nil {
		getListRecipientsRequest = req.Msg.GetRecipient()
	} else {
		getListRecipientsRequest = &pbRecipients.GetListRequest{}
	}
	qb := s.Repo.QbGetList(
		req.Msg,
		recipients.GetSingletonServiceServer(s.db).Repo.QbGetList(getListRecipientsRequest),
	)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr + " with args: " + fmt.Sprint(args))
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbDefiwallets.GetListResponse{
					Response: &pbDefiwallets.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		defiwallet, err := s.Repo.ScanListRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbDefiwallets.GetListResponse{
						Response: &pbDefiwallets.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbDefiwallets.GetListResponse{
			Response: &pbDefiwallets.GetListResponse_Defiwallet{
				Defiwallet: defiwallet,
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
	req *connect.Request[pbRecipients.DeleteRequest],
) (*connect.Response[pbDefiwallets.DeleteResponse], error) {
	msgUpdate := &pbDefiwallets.UpdateRequest{
		Recipient: &pbRecipients.UpdateRequest{
			Select: req.Msg.GetSelect(),
			Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
		},
	}

	softDeletedDefiwallet, err := s.Update(ctx, &connect.Request[pbDefiwallets.UpdateRequest]{
		Msg: msgUpdate,
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbDefiwallets.DeleteResponse{
			Response: &pbDefiwallets.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    softDeletedDefiwallet.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + softDeletedDefiwallet.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbDefiwallets.DeleteResponse{
		Response: &pbDefiwallets.DeleteResponse_Defiwallet{
			Defiwallet: softDeletedDefiwallet.Msg.GetDefiwallet(),
		},
	}), nil
}
