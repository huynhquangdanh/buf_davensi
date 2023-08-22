package dvsubaccounts

import (
	"context"
	"fmt"
	"sync"

	pbCommon "davensi.com/core/gen/common"
	pbDvSubAccounts "davensi.com/core/gen/dvsubaccounts"
	pbDvSubAccountsConnect "davensi.com/core/gen/dvsubaccounts/dvsubaccountsconnect"
	pbRecipients "davensi.com/core/gen/recipients"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/recipients"
	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// For singleton DvSubAccounts export module
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

// ServiceServer implements the DvSubAccountsService API
type ServiceServer struct {
	Repo DvSubAccountRepository
	pbDvSubAccountsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewDvSubAccountRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbRecipients.GetRequest],
) (*connect.Response[pbDvSubAccounts.GetResponse], error) {
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
		return connect.NewResponse(&pbDvSubAccounts.GetResponse{
			Response: &pbDvSubAccounts.GetResponse_Error{
				Error: &pbCommon.Error{
					Code: _errno,
					Text: _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), _err
	}

	defer rows.Close()

	if rows.Next() {
		dvSubAccount, err := s.Repo.ScanGetRow(rows)
		if err != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
			log.Error().Err(err).Msg(_err.Error())
			return connect.NewResponse(&pbDvSubAccounts.GetResponse{
				Response: &pbDvSubAccounts.GetResponse_Error{
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
			return connect.NewResponse(&pbDvSubAccounts.GetResponse{
				Response: &pbDvSubAccounts.GetResponse_Error{
					Error: &pbCommon.Error{
						Code: _errno,
						Text: _err.Error(),
					},
				},
			}), _err
		}
		return connect.NewResponse(&pbDvSubAccounts.GetResponse{
			Response: &pbDvSubAccounts.GetResponse_Dvsubaccount{
				Dvsubaccount: dvSubAccount,
			},
		}), err
	} else {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDvSubAccounts.GetResponse{
			Response: &pbDvSubAccounts.GetResponse_Error{
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
	req *connect.Request[pbDvSubAccounts.GetListRequest],
	res *connect.ServerStream[pbDvSubAccounts.GetListResponse],
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
				return res.Send(&pbDvSubAccounts.GetListResponse{
					Response: &pbDvSubAccounts.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		dvbot, err := s.Repo.ScanListRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbDvSubAccounts.GetListResponse{
						Response: &pbDvSubAccounts.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbDvSubAccounts.GetListResponse{
			Response: &pbDvSubAccounts.GetListResponse_Dvsubaccount{
				Dvsubaccount: dvbot,
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

// Create function ...
func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbDvSubAccounts.CreateRequest],
) (*connect.Response[pbDvSubAccounts.CreateResponse], error) {
	if err := s.validateQueryInsert(req.Msg); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			"creating '"+_entityName+"'", err.Err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbDvSubAccounts.CreateResponse{
			Response: &pbDvSubAccounts.CreateResponse_Error{
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
		return connect.NewResponse(&pbDvSubAccounts.CreateResponse{
			Response: &pbDvSubAccounts.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    recipientQBErr.Code,
					Package: _package,
					Text:    recipientQBErr.Err.Error(),
				},
			},
		}), recipientQBErr.Err
	}
	var newDvSubAccount *pbDvSubAccounts.DVSubAccount
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

		dvSubAccountQB, dvSubAccountQBErr := s.Repo.QbInsert(req.Msg, &pbRecipients.Recipient{
			Id: recipient.GetId(),
		})

		if dvSubAccountQBErr != nil {
			return dvSubAccountQBErr
		}

		dvSubAccountSQLStr, dvSubAccountArgs, _ := dvSubAccountQB.GenerateSQL()

		dvSubAccount, insertDvSubAccountErr := common.TxWrite[pbDvSubAccounts.DVSubAccount](
			ctx,
			tx,
			dvSubAccountSQLStr,
			dvSubAccountArgs,
			s.Repo.ScanRow,
		)

		if insertDvSubAccountErr != nil {
			return insertDvSubAccountErr
		}

		newDvSubAccount = dvSubAccount
		newDvSubAccount.Recipient = recipient

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
			recipientSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDvSubAccounts.CreateResponse{
			Response: &pbDvSubAccounts.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}
	log.Info().Msgf("%s with %s created successfully", _entityName, recipientSel)
	return connect.NewResponse(&pbDvSubAccounts.CreateResponse{
		Response: &pbDvSubAccounts.CreateResponse_Dvsubaccount{
			Dvsubaccount: newDvSubAccount,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbDvSubAccounts.UpdateRequest],
) (*connect.Response[pbDvSubAccounts.UpdateResponse], error) {
	if err := s.validateQueryUpdate(req.Msg); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			"creating '"+_entityName+"'", err.Err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbDvSubAccounts.UpdateResponse{
			Response: &pbDvSubAccounts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	recipientServiceServer := recipients.GetSingletonServiceServer(s.db)
	updateRecipientReq := req.Msg.GetRecipient()
	updateRecipientReq.Type = pbRecipients.Type_TYPE_DV_BOT.Enum()

	recipientQB, recipientQBErr := recipientServiceServer.MakeUpdateQB(updateRecipientReq)

	if recipientQBErr != nil {
		log.Error().Err(recipientQBErr.Err)
		return connect.NewResponse(&pbDvSubAccounts.UpdateResponse{
			Response: &pbDvSubAccounts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    recipientQBErr.Code,
					Package: _package,
					Text:    recipientQBErr.Err.Error(),
				},
			},
		}), recipientQBErr.Err
	}

	dvSubAccountQB, dvSubAccountQBErr := s.Repo.QbUpdate(req.Msg)
	if dvSubAccountQBErr != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_package+"'", dvSubAccountQBErr.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbDvSubAccounts.UpdateResponse{
			Response: &pbDvSubAccounts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	var updatedDvSubAccount *pbDvSubAccounts.DVSubAccount

	recipientSQLStr, recipientArgs, recipientSel := recipientQB.GenerateSQL()
	dvSubAccountQBStr, dvSubAccountArgs, dvSubAccountSel := dvSubAccountQB.GenerateSQL()

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

		dvSubAccount, updateDvSubAccountErr := common.TxWrite[pbDvSubAccounts.DVSubAccount](
			ctx,
			tx,
			dvSubAccountQBStr,
			dvSubAccountArgs,
			s.Repo.ScanUpdateRow,
		)

		if updateDvSubAccountErr != nil {
			return updateDvSubAccountErr
		}

		updatedDvSubAccount = dvSubAccount
		updatedDvSubAccount.Recipient = recipient

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			recipientSel,
			dvSubAccountSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDvSubAccounts.UpdateResponse{
			Response: &pbDvSubAccounts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, recipientSel+dvSubAccountSel)
	return connect.NewResponse(&pbDvSubAccounts.UpdateResponse{
		Response: &pbDvSubAccounts.UpdateResponse_Dvsubaccount{
			Dvsubaccount: updatedDvSubAccount,
		},
	}), nil
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbRecipients.DeleteRequest],
) (*connect.Response[pbDvSubAccounts.DeleteResponse], error) {
	msgUpdate := &pbDvSubAccounts.UpdateRequest{
		Recipient: &pbRecipients.UpdateRequest{
			Select: req.Msg.GetSelect(),
			Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
		},
	}

	softDeletedDvSubAccount, err := s.Update(ctx, &connect.Request[pbDvSubAccounts.UpdateRequest]{
		Msg: msgUpdate,
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbDvSubAccounts.DeleteResponse{
			Response: &pbDvSubAccounts.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    softDeletedDvSubAccount.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + softDeletedDvSubAccount.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbDvSubAccounts.DeleteResponse{
		Response: &pbDvSubAccounts.DeleteResponse_Dvsubaccount{
			Dvsubaccount: softDeletedDvSubAccount.Msg.GetDvsubaccount(),
		},
	}), nil
}
