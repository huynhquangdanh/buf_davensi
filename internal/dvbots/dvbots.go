package dvbots

import (
	"context"
	"fmt"
	"sync"

	pbCommon "davensi.com/core/gen/common"
	pbDvbots "davensi.com/core/gen/dvbots"
	pbDvbotConnect "davensi.com/core/gen/dvbots/dvbotsconnect"
	pbRecipients "davensi.com/core/gen/recipients"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/recipients"
	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// For singleton Dvbot export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

// ServiceServer implements the AddressesService API
type ServiceServer struct {
	Repo DvbotRepository
	pbDvbotConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewDvbotRepository(db),
		db:   db,
	}
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbRecipients.GetRequest],
) (*connect.Response[pbDvbots.GetResponse], error) {
	if errQueryGet := s.validateSelect(req.Msg.Select, "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbDvbots.GetResponse{
			Response: &pbDvbots.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

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
		return connect.NewResponse(&pbDvbots.GetResponse{
			Response: &pbDvbots.GetResponse_Error{
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
			return connect.NewResponse(&pbDvbots.GetResponse{
				Response: &pbDvbots.GetResponse_Error{
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
			return connect.NewResponse(&pbDvbots.GetResponse{
				Response: &pbDvbots.GetResponse_Error{
					Error: &pbCommon.Error{
						Code: _errno,
						Text: _err.Error(),
					},
				},
			}), _err
		}
		return connect.NewResponse(&pbDvbots.GetResponse{
			Response: &pbDvbots.GetResponse_Dvbot{
				Dvbot: dvbots,
			},
		}), err
	} else {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDvbots.GetResponse{
			Response: &pbDvbots.GetResponse_Error{
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
	req *connect.Request[pbDvbots.GetListRequest],
	res *connect.ServerStream[pbDvbots.GetListResponse],
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
				return res.Send(&pbDvbots.GetListResponse{
					Response: &pbDvbots.GetListResponse_Error{
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
					return res.Send(&pbDvbots.GetListResponse{
						Response: &pbDvbots.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbDvbots.GetListResponse{
			Response: &pbDvbots.GetListResponse_Dvbot{
				Dvbot: dvbot,
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
	req *connect.Request[pbDvbots.CreateRequest],
) (*connect.Response[pbDvbots.CreateResponse], error) {
	if err := s.validateCreate(req.Msg); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			"creating '"+_entityName+"'", err.Err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbDvbots.CreateResponse{
			Response: &pbDvbots.CreateResponse_Error{
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
		return connect.NewResponse(&pbDvbots.CreateResponse{
			Response: &pbDvbots.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    recipientQBErr.Code,
					Package: _package,
					Text:    recipientQBErr.Err.Error(),
				},
			},
		}), recipientQBErr.Err
	}
	var newDvbot *pbDvbots.DVBot
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

		dvbotQB, dvbotQBErr := s.Repo.QbInsert(req.Msg, &pbRecipients.Recipient{
			Id: recipient.GetId(),
		})

		if dvbotQBErr != nil {
			return dvbotQBErr
		}

		dvbotSQLStr, dvbotArgs, _ := dvbotQB.GenerateSQL()

		dvbot, insertDvbotErr := common.TxWrite[pbDvbots.DVBot](
			ctx,
			tx,
			dvbotSQLStr,
			dvbotArgs,
			s.Repo.ScanRow,
		)

		if insertDvbotErr != nil {
			return insertDvbotErr
		}

		newDvbot = dvbot
		newDvbot.Recipient = recipient

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
			recipientSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDvbots.CreateResponse{
			Response: &pbDvbots.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}
	log.Info().Msgf("%s with %s created successfully", _entityName, recipientSel)
	return connect.NewResponse(&pbDvbots.CreateResponse{
		Response: &pbDvbots.CreateResponse_Dvbot{
			Dvbot: newDvbot,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbDvbots.UpdateRequest],
) (*connect.Response[pbDvbots.UpdateResponse], error) {
	if err := s.validateQueryUpdate(req.Msg); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			"creating '"+_entityName+"'", err.Err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbDvbots.UpdateResponse{
			Response: &pbDvbots.UpdateResponse_Error{
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
		return connect.NewResponse(&pbDvbots.UpdateResponse{
			Response: &pbDvbots.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    recipientQBErr.Code,
					Package: _package,
					Text:    recipientQBErr.Err.Error(),
				},
			},
		}), recipientQBErr.Err
	}

	dvbotQB, dvbotQBErr := s.Repo.QbUpdate(req.Msg)
	if dvbotQBErr != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_package+"'", dvbotQBErr.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbDvbots.UpdateResponse{
			Response: &pbDvbots.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	var updatedDvbot *pbDvbots.DVBot

	recipientSQLStr, recipientArgs, recipientSel := recipientQB.GenerateSQL()
	dvbotQBStr, dvbotArgs, dvbotSel := dvbotQB.GenerateSQL()

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

		dvbot, updateDvbotErr := common.TxWrite[pbDvbots.DVBot](
			ctx,
			tx,
			dvbotQBStr,
			dvbotArgs,
			s.Repo.ScanUpdateRow,
		)

		if updateDvbotErr != nil {
			return updateDvbotErr
		}

		updatedDvbot = dvbot
		updatedDvbot.Recipient = recipient

		return nil
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			recipientSel,
			dvbotSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbDvbots.UpdateResponse{
			Response: &pbDvbots.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, recipientSel+dvbotSel)
	return connect.NewResponse(&pbDvbots.UpdateResponse{
		Response: &pbDvbots.UpdateResponse_Dvbot{
			Dvbot: updatedDvbot,
		},
	}), nil
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbRecipients.DeleteRequest],
) (*connect.Response[pbDvbots.DeleteResponse], error) {
	msgUpdate := &pbDvbots.UpdateRequest{
		Recipient: &pbRecipients.UpdateRequest{
			Select: req.Msg.GetSelect(),
			Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
		},
		BotStatus: pbDvbots.BotState_BOT_STATE_STOPPED.Enum(),
	}

	softDeletedDvbot, err := s.Update(ctx, &connect.Request[pbDvbots.UpdateRequest]{
		Msg: msgUpdate,
	})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbDvbots.DeleteResponse{
			Response: &pbDvbots.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    softDeletedDvbot.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + softDeletedDvbot.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbDvbots.DeleteResponse{
		Response: &pbDvbots.DeleteResponse_Dvbot{
			Dvbot: softDeletedDvbot.Msg.GetDvbot(),
		},
	}), nil
}
