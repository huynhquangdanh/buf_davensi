package cexaccounts

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	pbCexAccount "davensi.com/core/gen/cexaccounts"
	cexAccountConnect "davensi.com/core/gen/cexaccounts/cexaccountsconnect"
	pbCommon "davensi.com/core/gen/common"
	pbFSProvider "davensi.com/core/gen/fsproviders"
	pbRecipients "davensi.com/core/gen/recipients"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/fsproviders"
	"davensi.com/core/internal/recipients"
	"davensi.com/core/internal/util"
	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type ServiceServer struct {
	repo CexAccountRepository
	cexAccountConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo: *NewCexAccountRepository(db),
		db:   db,
	}
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) GetFSProviderFromSelect(
	ctx context.Context,
	providerSelect *pbFSProvider.Select,
) (*pbFSProvider.FSProvider, error) {
	fsProviderServiceServer := fsproviders.GetSingletonServiceServer(s.db)

	var foundFsProvider *pbFSProvider.FSProvider
	switch providerSelect.Select.(type) {
	case *pbFSProvider.Select_ById:
		fsProviderGetResponse, foundProviderErr := fsProviderServiceServer.Get(ctx, connect.NewRequest[pbFSProvider.GetRequest](
			&pbFSProvider.GetRequest{
				Select: &pbFSProvider.Select{
					Select: &pbFSProvider.Select_ById{
						ById: providerSelect.GetById(),
					},
				},
			},
		))
		if foundProviderErr != nil {
			return nil, foundProviderErr
		}
		foundFsProvider = fsProviderGetResponse.Msg.GetFsprovider()
		if foundFsProvider == nil {
			return nil, errors.New("provider is not found")
		}

	case *pbFSProvider.Select_ByTypeName:
		fsProviderGetResponse, foundProviderErr := fsProviderServiceServer.Get(ctx, connect.NewRequest[pbFSProvider.GetRequest](
			&pbFSProvider.GetRequest{
				Select: &pbFSProvider.Select{
					Select: &pbFSProvider.Select_ByTypeName{
						ByTypeName: providerSelect.GetByTypeName(),
					},
				},
			},
		))
		if foundProviderErr != nil {
			return nil, foundProviderErr
		}
		foundFsProvider = fsProviderGetResponse.Msg.GetFsprovider()
		if foundFsProvider == nil {
			return nil, errors.New("provider is not found")
		}
	}
	return foundFsProvider, nil
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbCexAccount.CreateRequest],
) (*connect.Response[pbCexAccount.CreateResponse], error) {
	// get qb insert for recipients
	recipientServiceServer := recipients.GetSingletonServiceServer(s.db)
	PKRes, validateErr := recipientServiceServer.ValidateCreate(req.Msg.Recipient)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbCexAccount.CreateResponse{
			Response: &pbCexAccount.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}
	if err := s.validateCreateRequest(req.Msg); err != nil {
		return connect.NewResponse(&pbCexAccount.CreateResponse{
			Response: &pbCexAccount.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	qbInsertRecipient, err := recipientServiceServer.Repo.QbInsert(req.Msg.GetRecipient(), PKRes)
	if err != nil {
		log.Error().Err(err)
		errorCode := pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED
		return connect.NewResponse(&pbCexAccount.CreateResponse{
			Response: &pbCexAccount.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errorCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	recipientSQLStr, recipientArgs, recipientSel := qbInsertRecipient.GenerateSQL()
	var newCexAccount *pbCexAccount.CExAccount

	err = crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		recipient, insertRecipientErr := common.TxWrite[pbRecipients.Recipient](
			ctx, tx, recipientSQLStr, recipientArgs, recipientServiceServer.Repo.ScanRow,
		)

		if insertRecipientErr != nil {
			return insertRecipientErr
		}

		foundFsProvider, foundFSProviderErr := s.GetFSProviderFromSelect(ctx, req.Msg.Provider)
		if foundFSProviderErr != nil {
			return foundFSProviderErr
		}
		// check fsprovider first
		cexAccountDB, cexAccountQBErr := s.repo.QbInsert(recipient.Id, foundFsProvider.Id)

		if cexAccountQBErr != nil {
			return cexAccountQBErr
		}
		cexAccountSQLStr, cexAccountArgs, _ := cexAccountDB.GenerateSQL()
		cexAccount, insertCexAccountErr := common.TxWrite[pbCexAccount.CExAccount](
			ctx, tx, cexAccountSQLStr, cexAccountArgs, s.repo.ScanRow,
		)
		if insertCexAccountErr != nil {
			return insertCexAccountErr
		}
		newCexAccount = cexAccount
		newCexAccount.Recipient = recipient
		newCexAccount.Provider = foundFsProvider
		return nil
	})
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
			recipientSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCexAccount.CreateResponse{
			Response: &pbCexAccount.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}
	return connect.NewResponse(&pbCexAccount.CreateResponse{
		Response: &pbCexAccount.CreateResponse_Cexaccount{
			Cexaccount: newCexAccount,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbCexAccount.UpdateRequest],
) (*connect.Response[pbCexAccount.UpdateResponse], error) {
	if err := s.ValidateUpdateRequest(req.Msg); err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbCexAccount.UpdateResponse{
			Response: &pbCexAccount.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	recipientServiceServer := recipients.GetSingletonServiceServer(s.db)
	if recipientServiceServer == nil {
		return nil, errors.New("recipient service server not found")
	}
	getRecipientReq := &pbRecipients.GetRequest{}
	if req.Msg.GetRecipient().GetSelect().GetById() != "" {
		getRecipientReq.Select = &pbRecipients.Select{
			Select: &pbRecipients.Select_ById{
				ById: req.Msg.GetRecipient().GetSelect().GetById(),
			},
		}
	} else {
		getRecipientReq.Select = &pbRecipients.Select{
			Select: &pbRecipients.Select_ByLegalEntityUserLabel{
				ByLegalEntityUserLabel: req.Msg.GetRecipient().GetSelect().GetByLegalEntityUserLabel(),
			},
		}
	}

	getRecipientRes, err := recipientServiceServer.Get(ctx, connect.NewRequest[pbRecipients.GetRequest](getRecipientReq))
	if err != nil {
		return connect.NewResponse(
			&pbCexAccount.UpdateResponse{
				Response: &pbCexAccount.UpdateResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
						Package: _package,
						Text:    err.Error(),
					},
				},
			},
		), err
	}

	foundRecipient := getRecipientRes.Msg.GetRecipient()
	foundFSProvider, err := s.GetFSProviderFromSelect(ctx, req.Msg.GetProvider())
	if err != nil {
		return connect.NewResponse(
			&pbCexAccount.UpdateResponse{
				Response: &pbCexAccount.UpdateResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
						Package: _package,
						Text:    err.Error(),
					},
				},
			},
		), err
	}

	qbCexAccountUpdate, err := s.repo.QbUpdate(foundRecipient.Id, foundFSProvider.Id)
	if err != nil {
		return connect.NewResponse(&pbCexAccount.UpdateResponse{
			Response: &pbCexAccount.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	qbCexAccountUpdate.SetReturnFields("*")
	sqlStr, args, sel := qbCexAccountUpdate.GenerateSQL()

	updatedCexAccount, err := common.ExecuteTxWrite[pbCexAccount.CExAccount](
		ctx, s.db, sqlStr, args, s.repo.ScanRow,
	)
	if err != nil {
		errUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)

		log.Error().Err(errUpdate.Err)
		return connect.NewResponse(&pbCexAccount.UpdateResponse{
			Response: &pbCexAccount.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	updatedCexAccount.Recipient = getRecipientRes.Msg.GetRecipient()
	updatedCexAccount.Provider = foundFSProvider
	return connect.NewResponse(&pbCexAccount.UpdateResponse{
		Response: &pbCexAccount.UpdateResponse_Cexaccount{
			Cexaccount: updatedCexAccount,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbRecipients.GetRequest],
) (*connect.Response[pbCexAccount.GetResponse], error) {
	recipientServiceServer := recipients.GetSingletonServiceServer(s.db)
	if _, validateErr := recipientServiceServer.ValidateGet(req.Msg); validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbCexAccount.GetResponse{
			Response: &pbCexAccount.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	getRecipientReq := &pbRecipients.GetRequest{}
	if req.Msg.GetSelect().GetById() != "" {
		getRecipientReq.Select = &pbRecipients.Select{
			Select: &pbRecipients.Select_ById{
				ById: req.Msg.GetSelect().GetById(),
			},
		}
	} else {
		getRecipientReq.Select = &pbRecipients.Select{
			Select: &pbRecipients.Select_ByLegalEntityUserLabel{
				ByLegalEntityUserLabel: req.Msg.GetSelect().GetByLegalEntityUserLabel(),
			},
		}
	}

	getRecipientRes, err := recipientServiceServer.Get(ctx, connect.NewRequest[pbRecipients.GetRequest](getRecipientReq))
	if err != nil {
		return connect.NewResponse(&pbCexAccount.GetResponse{
			Response: &pbCexAccount.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	qb := s.repo.QbGetOne(getRecipientRes.Msg.GetRecipient().Id)
	sqlStr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")

	rows, err := s.db.Query(ctx, sqlStr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCexAccount.GetResponse{
			Response: &pbCexAccount.GetResponse_Error{
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
		return connect.NewResponse(&pbCexAccount.GetResponse{
			Response: &pbCexAccount.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	cexaccount, err := s.repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCexAccount.GetResponse{
			Response: &pbCexAccount.GetResponse_Error{
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
		return connect.NewResponse(&pbCexAccount.GetResponse{
			Response: &pbCexAccount.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	getProvider, getProviderErr := s.GetFSProviderFromSelect(ctx, &pbFSProvider.Select{
		Select: &pbFSProvider.Select_ById{
			ById: cexaccount.GetProvider().GetId(),
		},
	})
	if err != nil {
		return connect.NewResponse(&pbCexAccount.GetResponse{
			Response: &pbCexAccount.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
					Package: _package,
					Text:    getProviderErr.Error(),
				},
			},
		}), err
	}
	cexaccount.Recipient = getRecipientRes.Msg.GetRecipient()
	cexaccount.Provider = getProvider
	// Start building the response from here
	return connect.NewResponse(&pbCexAccount.GetResponse{
		Response: &pbCexAccount.GetResponse_Cexaccount{
			Cexaccount: cexaccount,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbCexAccount.GetListRequest],
	res *connect.ServerStream[pbCexAccount.GetListResponse],
) error {
	qb := s.repo.QbGetList(req.Msg)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		cexaccount, err := s.repo.ScanRow(rows)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}

		if errSend := res.Send(&pbCexAccount.GetListResponse{
			Response: &pbCexAccount.GetListResponse_Cexaccount{
				Cexaccount: cexaccount,
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func streamingErr(res *connect.ServerStream[pbCexAccount.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbCexAccount.GetListResponse{
		Response: &pbCexAccount.GetListResponse_Error{
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

func (s *CexAccountRepository) QbGetRecipientStatement(msg *pbCexAccount.GetListRequest) *util.QueryBuilder {
	qbRecipients := util.CreateQueryBuilder(util.Select, "recipients")
	qbRecipients.Select("id")

	req := msg.GetRecipient()
	if req.Label != nil {
		qbRecipients.Where("label LIKE '%' || ? || '%'", req.GetLabel())
	}

	if req.LegalEntity != nil {
		qbRecipients.Where("legalentity_id = '%' || ? || '%'", req.GetLegalEntity())
	}

	if req.User != nil {
		qbRecipients.Where("user_id = '%' || ? || '%'", req.GetUser())
	}

	if req.Org != nil {
		qbRecipients.Where("org_id = '%' || ? || '%'", req.GetOrg())
	}

	if req.Type != nil {
		recipientTypes := req.GetType().GetList()

		if len(recipientTypes) > 0 {
			var args []any
			for _, v := range recipientTypes {
				args = append(args, v)
			}

			qbRecipients.Where(
				fmt.Sprintf(
					"type IN (%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(recipientTypes)), ""), ", "),
				),
				args...,
			)
		}
	}

	if req.Status != nil {
		recipientStatuses := req.GetStatus().GetList()

		if len(recipientStatuses) == 1 {
			qbRecipients.Where(fmt.Sprintf("status = %d", recipientStatuses[0]))
		}
		if len(recipientStatuses) > 1 {
			var args []any
			for _, v := range recipientStatuses {
				args = append(args, v)
			}

			qbRecipients.Where(
				fmt.Sprintf(
					"status = IN (%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(recipientStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}
	return qbRecipients
}

func (s *CexAccountRepository) QbGetProvider(msg *pbCexAccount.GetListRequest) *util.QueryBuilder {
	req := msg.GetProvider()
	qbProvider := util.CreateQueryBuilder(util.Select, "fsproviders")
	qbProvider.Select("id")
	if req.Name != nil {
		qbProvider.Where("label LIKE '%' || ? || '%'", req.GetName())
	}
	if req.Icon != nil {
		qbProvider.Where("label LIKE '%' || ? || '%'", req.GetIcon())
	}

	if req.Status != nil {
		fsProviderStatuses := req.GetStatus().GetList()

		var args []any
		for _, v := range fsProviderStatuses {
			args = append(args, v)
		}

		qbProvider.Where(
			fmt.Sprintf(
				"status IN (%s)",
				strings.Join(strings.Split(strings.Repeat("?", len(fsProviderStatuses)), ""), ", "),
			),
			args...,
		)
	}

	if req.Type != nil {
		fsProviderTypes := req.GetType().GetList()

		var args []any
		for _, v := range fsProviderTypes {
			args = append(args, v)
		}

		qbProvider.Where(
			fmt.Sprintf(
				"status IN (%s)",
				strings.Join(strings.Split(strings.Repeat("?", len(fsProviderTypes)), ""), ", "),
			),
			args...,
		)
	}
	return qbProvider
}

func (s *CexAccountRepository) QbGetList(msg *pbCexAccount.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _table)
	qb.Select(_fields)

	if msg.GetRecipient() != nil {
		qbRecipients := s.QbGetRecipientStatement(msg)
		sqlRecipientStr, sqlRecipientArgs, _ := qbRecipients.GenerateSQL()
		qb.Where(fmt.Sprintf("id IN (%s)", sqlRecipientStr), sqlRecipientArgs...)
	}

	if msg.GetProvider() != nil {
		qbProvider := s.QbGetProvider(msg)
		sqlProviderStr, sqlProviderArgs, _ := qbProvider.GenerateSQL()
		qb.Where(fmt.Sprintf("fsprovider_id IN (%s)", sqlProviderStr), sqlProviderArgs...)
	}

	return qb
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbRecipients.DeleteRequest],
) (*connect.Response[pbCexAccount.DeleteResponse], error) {
	msgUpdate := &pbCexAccount.UpdateRequest{
		Recipient: &pbRecipients.UpdateRequest{
			Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
		},
	}
	msgGet := &pbRecipients.GetRequest{}

	switch req.Msg.GetSelect().GetSelect().(type) {
	case *pbRecipients.Select_ById:
		msgUpdate.Recipient.Select = &pbRecipients.Select{
			Select: &pbRecipients.Select_ById{
				ById: req.Msg.GetSelect().GetById(),
			},
		}
		msgGet.Select = &pbRecipients.Select{
			Select: &pbRecipients.Select_ById{
				ById: req.Msg.GetSelect().GetById(),
			},
		}
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		msgUpdate.Recipient.Select = &pbRecipients.Select{
			Select: &pbRecipients.Select_ByLegalEntityUserLabel{
				ByLegalEntityUserLabel: req.Msg.GetSelect().GetByLegalEntityUserLabel(),
			},
		}
		msgGet.Select = &pbRecipients.Select{
			Select: &pbRecipients.Select_ByLegalEntityUserLabel{
				ByLegalEntityUserLabel: req.Msg.GetSelect().GetByLegalEntityUserLabel(),
			},
		}
	}

	recipientServiceServer := recipients.GetSingletonServiceServer(s.db)
	if recipientServiceServer == nil {
		return nil, errors.New("recipient service server not found")
	}
	getCexaccountRes, getErr := s.Get(ctx, &connect.Request[pbRecipients.GetRequest]{
		Msg: msgGet,
	})
	if getErr != nil {
		return connect.NewResponse(
			&pbCexAccount.DeleteResponse{
				Response: &pbCexAccount.DeleteResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
						Package: _package,
						Text:    getErr.Error(),
					},
				},
			},
		), getErr
	}
	cexaccount := getCexaccountRes.Msg.GetCexaccount()
	pkResOld, pkResNew, validateErr := recipientServiceServer.ValidateUpdate(msgUpdate.GetRecipient())
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbCexAccount.DeleteResponse{
			Response: &pbCexAccount.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}
	qbUpdateRecipient, updateErr := recipientServiceServer.Repo.QbUpdate(msgUpdate.GetRecipient(), pkResOld, pkResNew)
	if updateErr != nil {
		log.Error().Err(updateErr)
		errorCode := pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED
		return connect.NewResponse(&pbCexAccount.DeleteResponse{
			Response: &pbCexAccount.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    errorCode,
					Package: _package,
					Text:    updateErr.Error(),
				},
			},
		}), updateErr
	}
	recipientSQLStr, recipientArgs, recipientSel := qbUpdateRecipient.GenerateSQL()
	var deletedCexaccount *pbCexAccount.CExAccount

	err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		recipient, deleteRecipientErr := common.TxWrite[pbRecipients.Recipient](
			ctx, tx, recipientSQLStr, recipientArgs, recipientServiceServer.Repo.ScanRow,
		)
		if deleteRecipientErr != nil {
			return deleteRecipientErr
		}

		cexAccountDB := s.repo.QbDelete(recipient.Id)
		cexAccountSQLStr, cexAccountArgs, _ := cexAccountDB.GenerateSQL()
		cexAccount, deleteCexAccountErr := tx.Query(ctx, cexAccountSQLStr, cexAccountArgs...)
		if deleteCexAccountErr != nil {
			return deleteCexAccountErr
		}
		defer cexAccount.Close()
		if deleteCexAccountErr != nil {
			return deleteCexAccountErr
		}
		deletedCexaccount = cexaccount
		deletedCexaccount.Recipient = recipient
		deletedCexaccount.Provider = cexaccount.GetProvider()
		return nil
	})
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"delete",
			_entityName,
			recipientSel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCexAccount.DeleteResponse{
			Response: &pbCexAccount.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	return connect.NewResponse(&pbCexAccount.DeleteResponse{
		Response: &pbCexAccount.DeleteResponse_Cexaccount{
			Cexaccount: deletedCexaccount,
		},
	}), nil
}
