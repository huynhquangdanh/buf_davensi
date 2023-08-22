package bankaccounts

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbBankAccounts "davensi.com/core/gen/bankaccounts"
	pbBankAccountsConnect "davensi.com/core/gen/bankaccounts/bankaccountsconnect"
	pbBankBranches "davensi.com/core/gen/bankbranches"
	pbCommon "davensi.com/core/gen/common"
	pbRecipients "davensi.com/core/gen/recipients"
	pbUoms "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/bankbranches"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/recipients"
	"davensi.com/core/internal/uoms"
)

const (
	_package          = "bankaccounts"
	_entityName       = "Bank Account"
	_entityNamePlural = "Bank Accounts"
)

// ServiceServer implements the BanksService API
type ServiceServer struct {
	repo BankAccountRepository
	pbBankAccountsConnect.UnimplementedServiceHandler
	db             *pgxpool.Pool
	recipientsSS   *recipients.ServiceServer
	bankBranchesSS *bankbranches.ServiceServer
	uomsSS         *uoms.ServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo:           *NewBankRepository(db),
		db:             db,
		recipientsSS:   recipients.GetSingletonServiceServer(db),
		bankBranchesSS: bankbranches.GetSingletonServiceServer(db),
		uomsSS:         uoms.GetSingletonServiceServer(db),
	}
}

func (s *ServiceServer) Create(ctx context.Context, req *connect.Request[pbBankAccounts.CreateRequest],
) (*connect.Response[pbBankAccounts.CreateResponse], error) {
	var (
		recipentUUID        string
		recipientCreationFn func(tx pgx.Tx) (*pbRecipients.Recipient, error)

		genErr *common.ErrWithCode
	)

	// id field, is also recipient's ID
	if req.Msg.Recipient != nil {
		recipentUUID = uuid.NewString()
		recipientCreationFn, genErr = s.recipientsSS.GenCreateFunc(req.Msg.Recipient, recipentUUID)
		if genErr != nil {
			log.Error().Err(genErr.Err)
			return connect.NewResponse(&pbBankAccounts.CreateResponse{
				Response: &pbBankAccounts.CreateResponse_Error{
					Error: &pbCommon.Error{
						Code:    genErr.Code,
						Package: _package,
						Text:    genErr.Err.Error(),
					},
				},
			}), genErr.Err
		}
	}

	bankAccountCreationFunc, genErr := s.GenCreateFunc(req.Msg, recipentUUID)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return connect.NewResponse(&pbBankAccounts.CreateResponse{
			Response: &pbBankAccounts.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    genErr.Code,
					Package: _package,
					Text:    genErr.Err.Error(),
				},
			},
		}), genErr.Err
	}

	var newBankAccount *pbBankAccounts.BankAccount
	if errExcute := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if req.Msg.Recipient != nil {
			_, errWriteRecipient := recipientCreationFn(tx)
			if errWriteRecipient != nil {
				return nil
			}
		}

		excutedBankAccount, errWriteBankAccount := bankAccountCreationFunc(tx)
		if errWriteBankAccount != nil {
			return nil
		}
		newBankAccount = excutedBankAccount

		return errWriteBankAccount
	}); errExcute != nil {
		commonErrCreate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_entityName,
			errExcute.Error(),
		)

		log.Error().Err(commonErrCreate.Err)
		return connect.NewResponse(&pbBankAccounts.CreateResponse{
			Response: &pbBankAccounts.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErrCreate.Code,
					Package: _package,
					Text:    commonErrCreate.Err.Error(),
				},
			},
		}), commonErrCreate.Err
	}

	log.Info().Msgf(
		"%s created successfully with id = %s",
		_entityName, newBankAccount.GetRecipient().GetId(),
	)

	return connect.NewResponse(&pbBankAccounts.CreateResponse{
		Response: &pbBankAccounts.CreateResponse_BankAccount{
			BankAccount: newBankAccount,
		},
	}), nil
}

func (s *ServiceServer) GenCreateFunc(req *pbBankAccounts.CreateRequest, recipientUUID string) (
	func(tx pgx.Tx) (*pbBankAccounts.BankAccount, error), *common.ErrWithCode,
) {
	errGenFn := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR, "creating", _entityName, "")

	if validateErr := s.validateCreate(req); validateErr != nil {
		return nil, validateErr
	}

	qb, errInsert := s.repo.QbInsert(req, recipientUUID)
	if errInsert != nil {
		return nil, errGenFn.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT).
			UpdateMessage(errInsert.Error())
	}

	sqlStr, args, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	fmt.Println(args...)

	return func(tx pgx.Tx) (*pbBankAccounts.BankAccount, error) {
		executedBankBranch, errWriteBankBranch := common.TxWrite[pbBankAccounts.BankAccount](
			context.Background(),
			tx,
			sqlStr,
			args,
			ScanRow,
		)

		if errWriteBankBranch != nil {
			return nil, errWriteBankBranch
		}

		return executedBankBranch, nil
	}, nil
}

func (s *ServiceServer) Get(ctx context.Context, req *connect.Request[pbRecipients.GetRequest],
) (*connect.Response[pbBankAccounts.GetResponse], error) {
	commonErr := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, "fetching", _entityName, "")

	if _, errQueryGet := s.recipientsSS.ValidateGet(req.Msg); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbBankAccounts.GetResponse{
			Response: &pbBankAccounts.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	var (
		qbRecipient  = s.recipientsSS.Repo.QbGetList(&pbRecipients.GetListRequest{})
		qbBankBranch = s.bankBranchesSS.Repo.QbGetList(&pbBankBranches.GetListRequest{})
		qbUom        = s.uomsSS.Repo.QbGetList(&pbUoms.GetListRequest{})

		filterRecipientStr, filterRecipientArgs   = qbRecipient.Filters.GenerateSQL()
		filterBankBranchStr, filterBankBranchArgs = qbBankBranch.Filters.GenerateSQL()
		filterUomStr, filterUomArgs               = qbUom.Filters.GenerateSQL()
	)

	qb := s.repo.QbGetOne(req.Msg).
		Join(fmt.Sprintf("JOIN %s ON bankaccounts.id = recipients.id", qbRecipient.TableName)).
		Join(fmt.Sprintf("JOIN %s ON bankaccounts.bankbranch_id = bankbranches.id", qbBankBranch.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON bankaccounts.currency_id = uoms.id", qbUom.TableName)).
		Select(strings.Join(qbRecipient.SelectFields, ", ")).
		Select(strings.Join(qbBankBranch.SelectFields, ", ")).
		Select(strings.Join(qbUom.SelectFields, ", ")).
		Where(filterRecipientStr, filterRecipientArgs...).
		Where(filterBankBranchStr, filterBankBranchArgs...).
		Where(filterUomStr, filterUomArgs...)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	// Make query
	log.Info().Msg("Executing SQL '" + sqlstr + "'")
	rows, errQueryRows := s.db.Query(ctx, sqlstr, sqlArgs...)
	if errQueryRows != nil {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).UpdateMessage(errQueryRows.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbBankAccounts.GetResponse{
			Response: &pbBankAccounts.GetResponse_Error{
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
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND).UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbBankAccounts.GetResponse{
			Response: &pbBankAccounts.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	bankAccount, errScanRow := s.repo.ScanWithRelationship(rows)
	if errScanRow != nil {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR).UpdateMessage(errScanRow.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbBankAccounts.GetResponse{
			Response: &pbBankAccounts.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	if rows.Next() {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND).UpdateMessage(sel)
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbBankAccounts.GetResponse{
			Response: &pbBankAccounts.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	// Start building the response from here
	return connect.NewResponse(&pbBankAccounts.GetResponse{
		Response: &pbBankAccounts.GetResponse_BankAccount{
			BankAccount: bankAccount,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context, req *connect.Request[pbBankAccounts.UpdateRequest],
) (*connect.Response[pbBankAccounts.UpdateResponse], error) {
	// Check if Bank Branch, Currency exist
	if errQueryUpdate := s.validateUpdateQuery(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbBankAccounts.UpdateResponse{
			Response: &pbBankAccounts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	// Validation for recipient side
	_, _, validateErr := s.recipientsSS.ValidateUpdate(req.Msg.GetRecipient())
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbBankAccounts.UpdateResponse{
			Response: &pbBankAccounts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	recipientUpdateFn, _, genErr := s.recipientsSS.GenUpdateFunc(req.Msg.Recipient)
	if genErr != nil {
		log.Error().Err(genErr.Err)
		return connect.NewResponse(&pbBankAccounts.UpdateResponse{
			Response: &pbBankAccounts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    genErr.Code,
					Package: _package,
					Text:    genErr.Err.Error(),
				},
			},
		}), genErr.Err
	}

	// Executing update and saving response
	var (
		updatedBankAccount *pbBankAccounts.BankAccount
		errScan            error
		sqlstr             string
		sqlArgs            []any
		sel                string
	)

	err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// UPDATE RECIPIENT
		updatedRecipient, err := recipientUpdateFn(tx)
		if err != nil {
			return err
		}

		// UPDATE MAIN ENTITY
		// Generate QB based on updatedRecipient.ID first
		qb, genSQLError := s.repo.QbUpdate(req.Msg, updatedRecipient.Id)
		if genSQLError != nil {
			_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
				"updating '"+_entityName+"'", genSQLError.Error())
			log.Error().Err(_err)
			return genSQLError
		}
		qb.SetReturnFields("*")
		sqlstr, sqlArgs, sel = qb.GenerateSQL()

		// Then execute the generated update
		log.Info().Msg("Executing SQL '" + sqlstr + "'")
		row, err := tx.Query(ctx, sqlstr, sqlArgs...)
		if err != nil {
			return err
		}

		if !row.Next() {
			return fmt.Errorf("no bank account exists for that recipient ID")
		} else {
			updatedBankAccount, errScan = ScanRow(row)
			if errScan != nil {
				log.Error().Err(err).Msgf("unable to update %s with identifier = '%s'", _entityName, sel)
				return errScan
			}
			log.Info().Msgf("%s updated successfully with id = %s",
				_entityName, updatedBankAccount.GetRecipient().GetId())
			row.Close()
		}
		return nil
	})
	if err != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)], "updating", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbBankAccounts.UpdateResponse{
			Response: &pbBankAccounts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	return connect.NewResponse(&pbBankAccounts.UpdateResponse{
		Response: &pbBankAccounts.UpdateResponse_BankAccount{
			BankAccount: updatedBankAccount,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbBankAccounts.GetListRequest],
	res *connect.ServerStream[pbBankAccounts.GetListResponse],
) error {
	if req.Msg.Recipient == nil {
		req.Msg.Recipient = &pbRecipients.GetListRequest{}
	}

	if req.Msg.BankBranch == nil {
		req.Msg.BankBranch = &pbBankBranches.GetListRequest{}
	}

	if req.Msg.Currency == nil {
		req.Msg.Currency = &pbUoms.GetListRequest{}
	}

	var (
		qbRecipient  = s.recipientsSS.Repo.QbGetList(req.Msg.Recipient)
		qbBankBranch = s.bankBranchesSS.Repo.QbGetList(req.Msg.BankBranch)
		qbCurrency   = s.uomsSS.Repo.QbGetList(req.Msg.Currency)

		filterRecipientStr, filterRecipientArgs   = qbRecipient.Filters.GenerateSQL()
		filterBankBranchStr, filterBankBranchArgs = qbBankBranch.Filters.GenerateSQL()
		filterCurrencyStr, filterCurrencyArgs     = qbCurrency.Filters.GenerateSQL()
	)

	qb := s.repo.QbGetList(req.Msg)
	qb.Join(fmt.Sprintf("LEFT JOIN %s ON bankaccounts.id = recipients.id", qbRecipient.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON bankaccounts.bankbranch_id = bankbranches.id", qbBankBranch.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON bankaccounts.currency_id = uoms.id", qbCurrency.TableName)).
		Select(strings.Join(qbRecipient.SelectFields, ", ")).
		Select(strings.Join(qbBankBranch.SelectFields, ", ")).
		Select(strings.Join(qbCurrency.SelectFields, ", ")).
		Where(filterRecipientStr, filterRecipientArgs...).
		Where(filterBankBranchStr, filterBankBranchArgs...).
		Where(filterCurrencyStr, filterCurrencyArgs...)

	sqlStr, sqlArgs, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL: " + sqlStr)
	for i := 0; i < len(sqlArgs); i++ {
		fmt.Println(sqlArgs[i])
	}
	rows, err := s.db.Query(ctx, sqlStr, sqlArgs...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbBankAccounts.GetListResponse{
					Response: &pbBankAccounts.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}
	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		bankAccount, err := s.repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbBankAccounts.GetListResponse{
						Response: &pbBankAccounts.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbBankAccounts.GetListResponse{
			Response: &pbBankAccounts.GetListResponse_BankAccount{
				BankAccount: bankAccount,
			},
		}); errSend != nil {
			_errSend := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR)], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbRecipients.DeleteRequest],
) (*connect.Response[pbBankAccounts.DeleteResponse], error) {
	var (
		deletedBankAccount *connect.Response[pbBankAccounts.GetResponse]
		err                error
	)

	err = crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		recipientInput := &pbRecipients.DeleteRequest{
			Select: req.Msg.GetSelect(),
		}

		updateRes, err := s.recipientsSS.Delete(ctx, connect.NewRequest(recipientInput))
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			return err
		}

		deletedBankAccount, err = s.Get(ctx, connect.NewRequest(&pbRecipients.GetRequest{
			Select: &pbRecipients.Select{
				Select: &pbRecipients.Select_ById{
					ById: updateRes.Msg.GetRecipient().GetId(),
				},
			},
		}))
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			return err
		}

		log.Info().Msgf("%s with %s deleted successfully", _entityName, "id = "+updateRes.Msg.GetRecipient().Id)

		return nil
	})
	if err != nil {
		return connect.NewResponse(&pbBankAccounts.DeleteResponse{
			Response: &pbBankAccounts.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	return connect.NewResponse(&pbBankAccounts.DeleteResponse{
		Response: &pbBankAccounts.DeleteResponse_BankAccount{
			BankAccount: deletedBankAccount.Msg.GetBankAccount(),
		},
	}), nil
}
