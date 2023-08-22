package userids

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	pbAddresses "davensi.com/core/gen/addresses"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/credentials"
	"davensi.com/core/internal/livelinesses"
	"davensi.com/core/internal/physiques"
	"davensi.com/core/internal/socials"
	"davensi.com/core/internal/users"
	"davensi.com/core/internal/util"

	pbCommon "davensi.com/core/gen/common"
	pbCredential "davensi.com/core/gen/credentials"
	pbKyc "davensi.com/core/gen/kyc"
	pbLiveliness "davensi.com/core/gen/liveliness"
	pbPhysiques "davensi.com/core/gen/physiques"
	pbSocials "davensi.com/core/gen/socials"
	pbUserIDs "davensi.com/core/gen/userids"
	pbUserIDsConnect "davensi.com/core/gen/userids/useridsconnect"
	pbUsers "davensi.com/core/gen/users"
	internalAddress "davensi.com/core/internal/addresses"
	internalContact "davensi.com/core/internal/contacts"
	internalIncome "davensi.com/core/internal/incomes"
	internalUserAddress "davensi.com/core/internal/useraddresses"
	internalUserContact "davensi.com/core/internal/usercontacts"
	internalUserIncome "davensi.com/core/internal/userincomes"
)

// ServiceServer implements the AddressesService API
type ServiceServer struct {
	repo UserIDsRepository
	pbUserIDsConnect.ServiceHandler
	db              *pgxpool.Pool
	addressSS       internalAddress.ServiceServer
	contactSS       internalContact.ServiceServer
	incomeSS        internalIncome.ServiceServer
	userAddressSS   internalUserAddress.ServiceServer
	userAddressRepo internalUserAddress.UserAddressesRepository
	addressRepo     internalAddress.AddressRepository
	incomeRepo      internalIncome.IncomeRepository
	contactRepo     internalContact.ContactRepository
	userContactSS   internalUserContact.ServiceServer
	userIncomeSS    internalUserIncome.ServiceServer
	userContactRepo internalUserContact.UserContactsRepository
	userIncomeRepo  internalUserIncome.UserIncomesRepository
}

type QueryUtility struct {
	SQLStr  string
	SQLArgs []any
}

type ModifyRequest struct {
	CreateRequest      *pbUserIDs.CreateRequest
	UpdateRequest      *pbUserIDs.UpdateRequest
	AddContactsRequest *pbUserIDs.AddContactsRequest
	SetContactsRequest *pbUserIDs.SetContactsRequest
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo:            *NewUserIDRepository(db),
		db:              db,
		addressSS:       *internalAddress.GetSingletonServiceServer(db),
		contactSS:       *internalContact.GetSingletonServiceServer(db),
		incomeSS:        *internalIncome.GetSingletonServiceServer(db),
		userAddressSS:   *internalUserAddress.GetSingletonServiceServer(db),
		userContactSS:   *internalUserContact.GetSingletonServiceServer(db),
		userIncomeSS:    *internalUserIncome.GetSingletonServiceServer(db),
		addressRepo:     *internalAddress.GetSingletonRepository(db),
		userAddressRepo: *internalUserAddress.GetSingletonRepository(db),
		userContactRepo: *internalUserContact.GetSingletonRepository(db),
		userIncomeRepo:  *internalUserIncome.GetSingletonRepository(db),
	}
}

func (s *ServiceServer) Create(ctx context.Context, req *connect.Request[pbUserIDs.CreateRequest],
) (*connect.Response[pbUserIDs.CreateResponse], error) {
	// TODO: relationship fields
	// TODO: check duplicate primary key (only user_id column)

	// Verify that user.User is specified
	_errno, validateErr := s.validateCreate(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr)
		return connect.NewResponse(&pbUserIDs.CreateResponse{
			Response: &pbUserIDs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    validateErr.Error(),
				},
			},
		}), validateErr
	}
	isRecordExist, isRecordExistErr := s.IsRecordWithUserExist(ctx, req.Msg.GetUser())
	if isRecordExistErr != nil || isRecordExist {
		log.Error().Err(isRecordExistErr)
		return connect.NewResponse(&pbUserIDs.CreateResponse{
			Response: &pbUserIDs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"check exist", _entityName, "with user_id/user_login="+req.Msg.GetUser().String()),
				},
			},
		}), isRecordExistErr
	}

	// Verify that User exists in core.users
	var userID string
	if userIDScanErr := s.scanForExistUser(ctx, req.Msg.GetUser(), &userID); userIDScanErr != nil {
		return connect.NewResponse(&pbUserIDs.CreateResponse{
			Response: &pbUserIDs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
					Package: _package,
					Text:    userIDScanErr.Error(),
				},
			},
		}), userIDScanErr
	}
	queryMap, queryMapErr := s.createQueryMap(
		&ModifyRequest{
			CreateRequest: req.Msg,
		}, "CREATE", nil)
	if queryMapErr != nil {
		return connect.NewResponse(&pbUserIDs.CreateResponse{
			Response: &pbUserIDs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    queryMapErr.Err.Error(),
				},
			},
		}), queryMapErr.Err
	}
	newUserID := &pbUserIDs.UserId{}
	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.ExecuteModifyUserID(ctx, tx, true, newUserID, nil, userID, req.Msg.GetStatus(), queryMap); err != nil {
			return err
		}
		return nil
	}); errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"creating", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.CreateResponse{
			Response: &pbUserIDs.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserIDs.CreateResponse{
		Response: &pbUserIDs.CreateResponse_Userid{
			Userid: newUserID,
		},
	}), nil
}
func handleModifySubEntity[T any](
	ctx context.Context,
	tx pgx.Tx,
	queryMap map[string]*QueryUtility,
	entityKey string, isCreate bool, optionalCols *[]any,
	oldUserID, newUserID *pbUserIDs.UserId,
	scanner func(row pgx.Row) (*T, error),
) error {
	var (
		result  any
		scanErr error
	)
	if queryMap[entityKey] != nil {
		result, scanErr = common.TxWrite[T](
			ctx, tx, queryMap[entityKey].SQLStr, queryMap[entityKey].SQLArgs,
			scanner,
		)
		if scanErr != nil {
			return scanErr
		}
		switch entityKey {
		case "credentials":
			newUserID.Credentials = result.(*pbCredential.Credentials)
			if isCreate || newUserID.Credentials.GetId() != oldUserID.GetCredentials().GetId() {
				*optionalCols = append(*optionalCols, result)
			}
		case "physiques":
			newUserID.Physique = result.(*pbPhysiques.Physique)
			if isCreate || newUserID.Physique.GetId() != oldUserID.GetPhysique().GetId() {
				*optionalCols = append(*optionalCols, result)
			}
		case "liveliness":
			newUserID.Liveliness = result.(*pbLiveliness.Liveliness)
			if isCreate || newUserID.Liveliness.GetId() != oldUserID.GetLiveliness().GetId() {
				*optionalCols = append(*optionalCols, result)
			}
		case "socials":
			newUserID.Social = result.(*pbSocials.Social)
			if isCreate || newUserID.Social.GetId() != oldUserID.GetSocial().GetId() {
				*optionalCols = append(*optionalCols, result)
			}
		}
	}
	return nil
}
func (s *ServiceServer) ExecuteModifyUserID(
	ctx context.Context, tx pgx.Tx, isCreate bool, newUserID, oldUserID *pbUserIDs.UserId,
	scanUserID string, newStatus pbKyc.Status,
	queryMap map[string]*QueryUtility) (err error) {
	var (
		optionalCols []any
		userIDQB     *util.QueryBuilder
	)
	if err := handleModifySubEntity[pbCredential.Credentials](
		ctx, tx, queryMap, "credentials",
		isCreate, &optionalCols, oldUserID, newUserID,
		credentials.GetSingletonServiceServer(s.db).Repo.ScanRow); err != nil {
		return err
	}
	if err := handleModifySubEntity[pbPhysiques.Physique](
		ctx, tx, queryMap, "physiques",
		isCreate, &optionalCols, oldUserID, newUserID,
		physiques.GetSingletonServiceServer(s.db).Repo.ScanRow); err != nil {
		return err
	}
	if err := handleModifySubEntity[pbLiveliness.Liveliness](
		ctx, tx, queryMap, "liveliness",
		isCreate, &optionalCols, oldUserID, newUserID,
		livelinesses.GetSingletonServiceServer(s.db).Repo.ScanRow); err != nil {
		return err
	}
	if err := handleModifySubEntity[pbSocials.Social](
		ctx, tx, queryMap, "socials",
		isCreate, &optionalCols, oldUserID, newUserID,
		socials.ScanRow); err != nil {
		return err
	}

	if !isCreate && (newStatus != pbKyc.Status_STATUS_UNSPECIFIED || len(optionalCols) > 0) {
		userIDQB, err = s.repo.QbUpdate(scanUserID, newStatus, _table, optionalCols...)
	} else {
		userIDQB, err = s.repo.QbInsert(scanUserID, newStatus, _table, optionalCols...)
	}
	if err != nil {
		return err
	}
	userIDSQLStr, userIDArgs, _ := userIDQB.GenerateSQL()
	userIDRes, err := common.TxWrite[pbUserIDs.UserId](
		ctx, tx, userIDSQLStr, userIDArgs, s.repo.ScanRow)
	if err != nil {
		return err
	}
	newUserID.Status = userIDRes.Status
	newUserID.User = userIDRes.User
	return nil
}

func (s *ServiceServer) IsRecordWithUserExist(ctx context.Context, req *pbUsers.Select) (bool, error) {
	var (
		row pgx.Rows
		err error
	)

	switch req.GetSelect().(type) {
	case *pbUsers.Select_ByLogin:
		row, err = s.db.Query(ctx, "SELECT 1 FROM core.userids "+
			"LEFT JOIN core.users "+
			"ON core.userids.user_id = core.users.id "+
			"WHERE core.users.login = $1", req.GetByLogin())

	case *pbUsers.Select_ById:
		row, err = s.db.Query(ctx, "SELECT 1 FROM core.userids WHERE user_id = $1", req.GetById())
	}

	if err != nil {
		return false, err
	}
	if row.Next() {
		return true, nil
	}
	return false, nil
}

func (s *ServiceServer) Update(ctx context.Context, req *connect.Request[pbUserIDs.UpdateRequest],
) (*connect.Response[pbUserIDs.UpdateResponse], error) {
	errorNo, errValidate := s.validateUpdate(req.Msg)
	if errValidate != nil {
		errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err := fmt.Errorf(common.Errors[uint32(errno)],
			"updating '"+_entityName+"'", errValidate.Error())
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.UpdateResponse{
			Response: &pbUserIDs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	// Verify that User exists in core.users
	var userID string

	if userIDScanErr := s.scanForExistUser(ctx, req.Msg.GetUser(), &userID); userIDScanErr != nil {
		return connect.NewResponse(&pbUserIDs.UpdateResponse{
			Response: &pbUserIDs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
					Package: _package,
					Text:    userIDScanErr.Error(),
				},
			},
		}), userIDScanErr
	}
	isRecordExist, isRecordExistErr := s.IsRecordWithUserExist(ctx, req.Msg.GetUser())
	if isRecordExistErr != nil || !isRecordExist {
		log.Error().Err(isRecordExistErr)
		return connect.NewResponse(&pbUserIDs.UpdateResponse{
			Response: &pbUserIDs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    "update failed: record with user_id not exist",
				},
			},
		}), isRecordExistErr
	}
	// get old userid records
	oldUserID, oldUserIDErr := s.Get(ctx, &connect.Request[pbUserIDs.GetRequest]{
		Msg: &pbUserIDs.GetRequest{
			User: req.Msg.GetUser(),
		},
	})
	if oldUserIDErr != nil {
		return connect.NewResponse(&pbUserIDs.UpdateResponse{
			Response: &pbUserIDs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    oldUserIDErr.Error(),
				},
			},
		}), oldUserIDErr
	}
	queryMap, queryMapErr := s.createQueryMap(
		&ModifyRequest{
			UpdateRequest: req.Msg,
		}, "UPDATE", oldUserID.Msg.GetUserid())
	if queryMapErr != nil {
		return connect.NewResponse(&pbUserIDs.UpdateResponse{
			Response: &pbUserIDs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    queryMapErr.Err.Error(),
				},
			},
		}), queryMapErr.Err
	}
	updatedUserID := &pbUserIDs.UserId{
		User:        oldUserID.Msg.GetUserid().User,
		Credentials: oldUserID.Msg.GetUserid().Credentials,
		Physique:    oldUserID.Msg.GetUserid().Physique,
		Liveliness:  oldUserID.Msg.GetUserid().Liveliness,
		Social:      oldUserID.Msg.GetUserid().Social,
		Status:      oldUserID.Msg.GetUserid().Status,
	}
	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// check length of newAdditionalCols
		if err := s.ExecuteModifyUserID(
			ctx, tx, false,
			updatedUserID, oldUserID.Msg.GetUserid(),
			userID, req.Msg.GetStatus(), queryMap,
		); err != nil {
			return err
		}
		return nil
	}); errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"creating", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.UpdateResponse{
			Response: &pbUserIDs.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errorNo,
					Package: _package,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserIDs.UpdateResponse{
		Response: &pbUserIDs.UpdateResponse_Userid{
			Userid: updatedUserID,
		},
	}), nil
}

func (s *ServiceServer) queryUserID(ctx context.Context, req *pbUsers.Select) (pgx.Rows, error) {
	var (
		row pgx.Rows
		err error
	)

	switch req.GetSelect().(type) {
	case *pbUsers.Select_ByLogin:
		row, err = s.db.Query(ctx, "SELECT id FROM core.users WHERE login = $1", req.GetByLogin())

	case *pbUsers.Select_ById:
		row, err = s.db.Query(ctx, "SELECT id FROM core.users WHERE id = $1", req.GetById())
	}

	return row, err
}

func (s *ServiceServer) Get(ctx context.Context, req *connect.Request[pbUserIDs.GetRequest],
) (*connect.Response[pbUserIDs.GetResponse], error) {
	var (
		row    pgx.Rows
		err    error
		userID string
	)

	row, err = s.queryUserID(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.GetResponse{
			Response: &pbUserIDs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"fetching", _entityName, "user_id/login="+req.Msg.GetUser().String()),
				},
			},
		}), err
	}
	if !row.Next() {
		errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		err := fmt.Errorf(common.Errors[uint32(errno)],
			_entityName, "user_login/id="+req.Msg.GetUser().String()+" and user_id/login="+req.Msg.GetUser().String())
		return connect.NewResponse(&pbUserIDs.GetResponse{
			Response: &pbUserIDs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	} else {
		if errScan := row.Scan(&userID); errScan != nil {
			err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)],
				"scanning", "core.users", "for id")
			return connect.NewResponse(&pbUserIDs.GetResponse{
				Response: &pbUserIDs.GetResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), err
		}
	}
	row.Close()

	qb := s.repo.QbGetOne(userID, _table, _userIDsFields, []string{})
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.GetResponse{
			Response: &pbUserIDs.GetResponse_Error{
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
		return connect.NewResponse(&pbUserIDs.GetResponse{
			Response: &pbUserIDs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	userIDRes, err := s.repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.GetResponse{
			Response: &pbUserIDs.GetResponse_Error{
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
		return connect.NewResponse(&pbUserIDs.GetResponse{
			Response: &pbUserIDs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	userIDRes = s.GetAdditionalInfo(ctx, userIDRes)

	addresses, addressesErr := s.GetAddresses(ctx, userID)
	if addressesErr == nil {
		userIDRes.Addresses = addresses
	}

	contacts, contactsErr := s.GetContacts(ctx, userID)
	if contactsErr == nil {
		userIDRes.Contacts = contacts
	}

	incomes, incomesErr := s.GetIncomes(ctx, userID)
	if incomesErr == nil {
		userIDRes.Incomes = incomes
	}
	// Start building the response from here
	return connect.NewResponse(&pbUserIDs.GetResponse{
		Response: &pbUserIDs.GetResponse_Userid{
			Userid: userIDRes,
		},
	}), nil
}

func (s *ServiceServer) GetAdditionalInfo(ctx context.Context, userIDRes *pbUserIDs.UserId,
) *pbUserIDs.UserId {
	credentialsServiceServer := credentials.GetSingletonServiceServer(s.db)
	livelinessesServiceServer := livelinesses.GetSingletonServiceServer(s.db)
	physiqueServiceServer := physiques.GetSingletonServiceServer(s.db)
	socialsServiceServer := socials.GetSingletonServiceServer(s.db)
	usersServiceServer := users.GetSingletonServiceServer(s.db)

	credential, credentialErr := credentialsServiceServer.Get(ctx, connect.NewRequest[pbCredential.GetRequest](
		&pbCredential.GetRequest{
			Id: userIDRes.GetCredentials().GetId(),
		},
	))
	if credentialErr == nil {
		userIDRes.Credentials = credential.Msg.GetCredentials()
	}

	liveliness, livelinessErr := livelinessesServiceServer.Get(ctx, connect.NewRequest[pbLiveliness.GetRequest](
		&pbLiveliness.GetRequest{
			Id: userIDRes.GetLiveliness().GetId(),
		},
	))
	if livelinessErr == nil {
		userIDRes.Liveliness = liveliness.Msg.GetLiveliness()
	}

	physique, physiqueErr := physiqueServiceServer.Get(ctx, connect.NewRequest[pbPhysiques.GetRequest](
		&pbPhysiques.GetRequest{
			Id: userIDRes.GetPhysique().GetId(),
		},
	))
	if physiqueErr == nil {
		userIDRes.Physique = physique.Msg.GetPhysique()
	}

	social, socialErr := socialsServiceServer.Get(ctx, connect.NewRequest[pbSocials.GetRequest](
		&pbSocials.GetRequest{
			Id: userIDRes.GetCredentials().GetId(),
		},
	))
	if socialErr == nil {
		userIDRes.Social = social.Msg.GetSocial()
	}

	user, usersErr := usersServiceServer.Get(ctx, connect.NewRequest[pbUsers.GetRequest](
		&pbUsers.GetRequest{
			Select: &pbUsers.GetRequest_ById{
				ById: userIDRes.GetUser().GetId(),
			},
		},
	))
	if usersErr == nil {
		userIDRes.User = user.Msg.GetUser()
	}

	// Start building the response from here
	return userIDRes
}

func (s *ServiceServer) CheckRow(row pgx.Rows, userID string) (err *pbCommon.Error) {
	if !row.Next() {
		err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT)],
			"deleting "+_entityName, "user not found for that id/login")
		return &pbCommon.Error{
			Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			Package: _package,
			Text:    err.Error(),
		}
	} else {
		if errScan := row.Scan(&userID); errScan != nil {
			errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
			err := fmt.Errorf(common.Errors[uint32(errno)],
				"scanning", "core.users", "for id")
			return &pbCommon.Error{
				Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
				Package: _package,
				Text:    err.Error(),
			}
		}
	}
	row.Close()
	return err
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbUserIDs.DeleteRequest],
) (*connect.Response[pbUserIDs.DeleteResponse], error) {
	var (
		row pgx.Rows
		err error
		qb  *util.QueryBuilder
	)

	// Verify that User exists in core.users
	var userID string
	if userIDScanErr := s.scanForExistUser(ctx, req.Msg.GetUser(), &userID); userIDScanErr != nil {
		return connect.NewResponse(&pbUserIDs.DeleteResponse{
			Response: &pbUserIDs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
					Package: _package,
					Text:    userIDScanErr.Error(),
				},
			},
		}), userIDScanErr
	}
	_qb := s.repo.QbGetOne(userID, _table, _userIDsFields, []string{})
	sqlstr, sqlArgs, sel := _qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.DeleteResponse{
			Response: &pbUserIDs.DeleteResponse_Error{
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
		return connect.NewResponse(&pbUserIDs.DeleteResponse{
			Response: &pbUserIDs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	userIDRes, err := s.repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.DeleteResponse{
			Response: &pbUserIDs.DeleteResponse_Error{
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
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.DeleteResponse{
			Response: &pbUserIDs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	cancelledStatus := pbKyc.StatusList{
		List: []pbKyc.Status{*pbKyc.Status_STATUS_CANCELED.Enum()},
	}

	queryMap, queryMapErr := s.createQueryMap(
		&ModifyRequest{
			UpdateRequest: &pbUserIDs.UpdateRequest{
				Credentials: &pbCredential.UpdateCredentials{
					Status: &cancelledStatus,
				},
				Physique: &pbPhysiques.UpdatePhysique{
					Status: &cancelledStatus,
				},
				Social: &pbSocials.UpdateSocial{
					Status: &cancelledStatus,
				},
				User: &pbUsers.Select{
					Select: &pbUsers.Select_ById{
						ById: userIDRes.GetUser().GetId(),
					},
				},
				Status: pbKyc.Status_STATUS_CANCELED.Enum(),
			},
		}, "UPDATE", userIDRes)
	if queryMapErr != nil {
		return connect.NewResponse(&pbUserIDs.DeleteResponse{
			Response: &pbUserIDs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    queryMapErr.Err.Error(),
				},
			},
		}), queryMapErr.Err
	}

	qb, err = s.repo.QbDelete(userID, _table, []string{})
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.DeleteResponse{
			Response: &pbUserIDs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"deleting", _entityName, "user_id/login="+req.Msg.GetUser().String()),
				},
			},
		}), err
	}
	sqlStr, args, _ := qb.SetReturnFields("*").GenerateSQL()

	var errScan error
	deletedUserID := userIDRes
	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.ExecuteModifyUserID(ctx, tx, false, deletedUserID, userIDRes, userID, pbKyc.Status_STATUS_CANCELED, queryMap); err != nil {
			return err
		}
		if row, err = tx.Query(ctx, sqlStr, args...); err != nil {
			return err
		}
		defer row.Close()

		if row.Next() {
			if deletedUserID, errScan = s.repo.ScanRow(row); errScan != nil {
				log.Error().Err(err).Msgf("unable to delete %s with id/login = '%s'",
					_entityName, req.Msg.GetUser().String())
				return errScan
			}
			log.Info().Msgf("%s with id/login = '%s' removed successfully",
				_entityName, req.Msg.GetUser().String())
			return nil
		} else {
			return fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND)],
				_entityName, "id/login="+req.Msg.GetUser().String())
		}
	}); errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"creating", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.DeleteResponse{
			Response: &pbUserIDs.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserIDs.DeleteResponse{
		Response: &pbUserIDs.DeleteResponse_Userid{
			Userid: deletedUserID,
		},
	}), nil
}
func getQueryUtility[T any](
	obj *T,
	isCreate bool,
	makeCreationQBFunc func(*T) (*util.QueryBuilder, *common.ErrWithCode),
	makeUpsertQBFunc func(msg *T, upsert bool) (*util.QueryBuilder, *common.ErrWithCode),
) (*QueryUtility, *common.ErrWithCode) {
	var (
		queryBuilder    *util.QueryBuilder
		queryBuilderErr *common.ErrWithCode
	)
	if obj != nil && isCreate {
		queryBuilder, queryBuilderErr = makeCreationQBFunc(obj)
	}
	if obj != nil && !isCreate {
		queryBuilder, queryBuilderErr = makeUpsertQBFunc(obj, true)
	}
	if queryBuilderErr != nil {
		return nil, queryBuilderErr
	}
	sqlStr, sqlArgs, _ := queryBuilder.GenerateSQL()
	return &QueryUtility{
		SQLStr:  sqlStr,
		SQLArgs: sqlArgs,
	}, nil
}

func (s *ServiceServer) createQueryMap(
	createReq *ModifyRequest,
	modifyType string,
	oldUserIDToUpdate *pbUserIDs.UserId,
) (map[string]*QueryUtility, *common.ErrWithCode) {
	fmt.Println(oldUserIDToUpdate.String())
	queryMap := map[string]*QueryUtility{}
	var (
		queryUtility    *QueryUtility
		queryUtilityErr *common.ErrWithCode
	)
	switch modifyType {
	case _createSymbol:
		queryUtility, queryUtilityErr = getQueryUtility[pbCredential.CreateRequest](
			createReq.CreateRequest.GetCredentials(), true, credentials.GetSingletonServiceServer(s.db).MakeCreationQB, nil)
		if queryUtilityErr != nil {
			return nil, queryUtilityErr
		}
		queryMap["credentials"] = queryUtility
		queryUtility, queryUtilityErr = getQueryUtility[pbLiveliness.CreateRequest](
			createReq.CreateRequest.GetLiveliness(), true, livelinesses.GetSingletonServiceServer(s.db).MakeCreationQB, nil)
		if queryUtilityErr != nil {
			return nil, queryUtilityErr
		}
		queryMap["liveliness"] = queryUtility
		queryUtility, queryUtilityErr = getQueryUtility[pbPhysiques.CreateRequest](
			createReq.CreateRequest.GetPhysique(), true, physiques.GetSingletonServiceServer(s.db).MakeCreationQB, nil)
		if queryUtilityErr != nil {
			return nil, queryUtilityErr
		}
		queryMap["physiques"] = queryUtility
		queryUtility, queryUtilityErr = getQueryUtility[pbSocials.CreateRequest](
			createReq.CreateRequest.GetSocial(), true, socials.GetSingletonServiceServer(s.db).MakeCreationQB, nil)
		if queryUtilityErr != nil {
			return nil, queryUtilityErr
		}
		queryMap["socials"] = queryUtility

	case _updateSymbol:
		if createReq.UpdateRequest.GetCredentials() != nil {
			conversedReq := createReq.UpdateRequest.GetCredentials()
			queryUtility, queryUtilityErr = getQueryUtility[pbCredential.UpdateRequest](
				&pbCredential.UpdateRequest{
					Id:          oldUserIDToUpdate.GetCredentials().GetId(),
					Photo:       conversedReq.Photo,
					Gender:      conversedReq.Gender,
					Title:       conversedReq.Title,
					FirstName:   conversedReq.FirstName,
					MiddleNames: conversedReq.MiddleNames,
					LastName:    conversedReq.LastName,
					Status:      &conversedReq.GetStatus().List[0],
				}, false, nil, credentials.GetSingletonServiceServer(s.db).MakeUpdateQB)
			if queryUtilityErr != nil {
				return nil, queryUtilityErr
			}
			queryMap["credentials"] = queryUtility
		}
		if createReq.UpdateRequest.GetLiveliness() != nil {
			log.Info().Msgf("update liveliness with id = '%s'", oldUserIDToUpdate.GetLiveliness().GetId())
			conversedReq := createReq.UpdateRequest.GetLiveliness()
			queryUtility, queryUtilityErr = getQueryUtility[pbLiveliness.UpdateRequest](
				&pbLiveliness.UpdateRequest{
					Id:                       oldUserIDToUpdate.GetLiveliness().GetId(),
					LivelinessVideoFile:      conversedReq.LivelinessVideoFile,
					LivelinessVideoFileType:  conversedReq.LivelinessVideoFileType,
					TimestampVideoFile:       conversedReq.TimestampVideoFile,
					TimestampVideoFileType:   conversedReq.TimestampVideoFileType,
					IdOwnershipPhotoFile:     conversedReq.IdOwnershipPhotoFile,
					IdOwnershipPhotoFileType: conversedReq.IdOwnershipPhotoFileType,
					Status:                   &conversedReq.GetStatus().List[0],
				}, false, nil, livelinesses.GetSingletonServiceServer(s.db).MakeUpdateQB)
			if queryUtilityErr != nil {
				return nil, queryUtilityErr
			}
			queryMap["liveliness"] = queryUtility
		}
		if createReq.UpdateRequest.GetPhysique() != nil {
			conversedReq := createReq.UpdateRequest.GetPhysique()
			queryUtility, queryUtilityErr = getQueryUtility[pbPhysiques.UpdateRequest](
				&pbPhysiques.UpdateRequest{
					Id:        oldUserIDToUpdate.GetPhysique().GetId(),
					Race:      conversedReq.Race,
					Ethnicity: conversedReq.Ethnicity,
					EyesColor: conversedReq.EyesColor,
					HairColor: conversedReq.HairColor,
					BodyShape: conversedReq.BodyShape,
					Height:    conversedReq.Height,
					Weight:    conversedReq.Weight,
					Status:    &conversedReq.GetStatus().List[0],
				}, false, nil, physiques.GetSingletonServiceServer(s.db).MakeUpdateQB)
			if queryUtilityErr != nil {
				return nil, queryUtilityErr
			}
			queryMap["physiques"] = queryUtility
		}
		if createReq.UpdateRequest.GetSocial() != nil {
			conversedReq := createReq.UpdateRequest.GetSocial()
			queryUtility, queryUtilityErr = getQueryUtility[pbSocials.UpdateRequest](
				&pbSocials.UpdateRequest{
					Id:                 oldUserIDToUpdate.GetSocial().GetId(),
					RelationshipStatus: conversedReq.RelationshipStatus,
					Religion:           conversedReq.Religion,
					SocialClass:        conversedReq.SocialClass,
					Profession:         conversedReq.Profession,
					Status:             &conversedReq.GetStatus().List[0],
				}, false, nil, socials.GetSingletonServiceServer(s.db).MakeUpdateQB)
			if queryUtilityErr != nil {
				return nil, queryUtilityErr
			}
			queryMap["socials"] = queryUtility
		}
	}

	return queryMap, nil
}

func (s *ServiceServer) scanForExistUser(ctx context.Context, req *pbUsers.Select, userID *string) error {
	row, queryErr := s.queryUserID(ctx, req)
	if queryErr != nil {
		log.Error().Err(queryErr)
		return queryErr
	}
	defer row.Close()
	if !row.Next() {
		return fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT)],
			"creating/updating "+_entityName, "user does not exist")
	} else {
		if errScan := row.Scan(userID); errScan != nil {
			return fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)],
				"scanning", "core.users", "for id")
		}
	}
	return nil
}

func (s *ServiceServer) GetUserID(ctx context.Context, req *pbUsers.Select) (string, error) {
	var userID string
	row, err := s.queryUserID(ctx, req)
	if err != nil {
		return "", err
	}
	defer row.Close()
	if row.Next() {
		userID, err = s.repo.ScanID(row)
		if err != nil {
			return "", err
		}
	}
	return userID, nil
}

func (s *ServiceServer) SeparateAddressQuery(req *pbUserIDs.SetAddressesRequest) (
	updateUserAddresses []*pbAddresses.SetLabeledAddress,
	insertAddresses []*pbAddresses.SetLabeledAddress,
) {
	setLabelAddresses := req.Addresses.GetList()
	for _, labelAddress := range setLabelAddresses {
		if labelAddress.GetId() != "" {
			// id is not null
			insertAddresses = append(insertAddresses, labelAddress)
		} else if labelAddress.GetAddress() != nil {
			updateUserAddresses = append(updateUserAddresses, labelAddress)
		}
	}
	return updateUserAddresses, insertAddresses
}

func (s *ServiceServer) SetAddresses(ctx context.Context, req *connect.Request[pbUserIDs.SetAddressesRequest]) (
	*connect.Response[pbUserIDs.SetAddressesResponse], error,
) {
	var userID string
	if errCode, err := s.ValidateSetAddresses(req.Msg); err != nil {
		return connect.NewResponse(&pbUserIDs.SetAddressesResponse{
			Response: &pbUserIDs.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	pgxTx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbUserIDs.SetAddressesResponse{
			Response: &pbUserIDs.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if _err := pgxTx.Rollback(ctx); _err != nil {
			log.Err(_err)
		}
	}()

	userID, err = s.GetUserID(ctx, req.Msg.GetUser())
	if err != nil {
		return connect.NewResponse(&pbUserIDs.SetAddressesResponse{
			Response: &pbUserIDs.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	// get query builder for upsert user_addresses
	// separate update query and insert query with address
	updateUserAddresses, insertAddresses := s.SeparateAddressQuery(req.Msg)
	// insert address first and get id
	res, err := s.addressSS.ExecuteInsertManyAddresses(ctx, pgxTx, req.Msg.GetAddresses(), insertAddresses)
	if err != nil {
		return connect.NewResponse(&pbUserIDs.SetAddressesResponse{
			Response: &pbUserIDs.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	// query update
	for _, address := range updateUserAddresses {
		res[address.GetId()] = address
	}
	upsertLabelAddresses := []*pbAddresses.SetLabeledAddress{}
	for k, v := range res {
		labelAddress := &pbAddresses.SetLabeledAddress{
			Label: v.Label,
			Select: &pbAddresses.SetLabeledAddress_Id{
				Id: k,
			},
			Status:          v.Status,
			MainAddress:     v.MainAddress,
			OwnershipStatus: v.OwnershipStatus,
		}
		upsertLabelAddresses = append(upsertLabelAddresses, labelAddress)
	}
	result, err := s.userAddressSS.UpsertUserAddresses(ctx, pgxTx, userID, upsertLabelAddresses, "upsert")
	if err != nil {
		return connect.NewResponse(&pbUserIDs.SetAddressesResponse{
			Response: &pbUserIDs.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	if err = pgxTx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbUserIDs.SetAddressesResponse{
			Response: &pbUserIDs.SetAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbUserIDs.SetAddressesResponse{
		Response: &pbUserIDs.SetAddressesResponse_Addresses{
			Addresses: &pbUserIDs.AddressList{
				User: &pbUsers.Select{
					Select: &pbUsers.Select_ById{
						ById: userID,
					},
				},
				Name: "",
				Addresses: &pbAddresses.LabeledAddressList{
					List: result,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) AddAddresses(ctx context.Context, req *connect.Request[pbUserIDs.AddAddressesRequest]) (
	*connect.Response[pbUserIDs.AddAddressesResponse], error,
) {
	var userID string
	if errCode, err := s.ValidateAddAddresses(req.Msg); err != nil {
		return connect.NewResponse(&pbUserIDs.AddAddressesResponse{
			Response: &pbUserIDs.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbUserIDs.AddAddressesResponse{
			Response: &pbUserIDs.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	// get user id from req.user
	userID, err = s.GetUserID(ctx, req.Msg.GetUser())
	if err != nil {
		return connect.NewResponse(&pbUserIDs.AddAddressesResponse{
			Response: &pbUserIDs.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Error().Msg(err.Error())
		}
	}()
	res, err := s.addressSS.ExecuteInsertManyAddresses(ctx, tx, req.Msg.GetAddresses(), req.Msg.GetAddresses().GetList())
	if err != nil {
		return connect.NewResponse(&pbUserIDs.AddAddressesResponse{
			Response: &pbUserIDs.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	// query update
	insertLabelAddresses := []*pbAddresses.SetLabeledAddress{}
	for k, v := range res {
		labelAddress := &pbAddresses.SetLabeledAddress{
			Label: v.Label,
			Select: &pbAddresses.SetLabeledAddress_Id{
				Id: k,
			},
			Status:          v.Status,
			MainAddress:     v.MainAddress,
			OwnershipStatus: v.OwnershipStatus,
		}
		insertLabelAddresses = append(insertLabelAddresses, labelAddress)
	}
	result, err := s.userAddressSS.UpsertUserAddresses(ctx, tx, userID, insertLabelAddresses, "insert")
	if err != nil {
		return connect.NewResponse(&pbUserIDs.AddAddressesResponse{
			Response: &pbUserIDs.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}

	if err = tx.Commit(ctx); err != nil {
		return connect.NewResponse(&pbUserIDs.AddAddressesResponse{
			Response: &pbUserIDs.AddAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), nil
	}
	return connect.NewResponse(&pbUserIDs.AddAddressesResponse{
		Response: &pbUserIDs.AddAddressesResponse_Addresses{
			Addresses: &pbUserIDs.AddressList{
				User: &pbUsers.Select{
					Select: &pbUsers.Select_ById{
						ById: userID,
					},
				},
				Name: "",
				Addresses: &pbAddresses.LabeledAddressList{
					List: result,
				},
			},
		},
	}), nil
}

func (s *ServiceServer) GetUserAddressByIDLabel(
	ctx context.Context,
	userID, label string,
) (*pbAddresses.LabeledAddress, error) {
	qb := s.userAddressRepo.QbGetUserAddressByIDLabel(userID, label)
	sqlStr, sqlArgs, _ := qb.GenerateSQL()
	queryRows, queryErr := s.db.Query(ctx, sqlStr, sqlArgs...)
	if queryErr != nil {
		return nil, queryErr
	}

	defer queryRows.Close()
	userAddress, userAddressScanErr := s.userAddressRepo.ScanRow(queryRows)
	if userAddressScanErr != nil {
		return nil, userAddressScanErr
	}
	if userAddress == nil {
		return nil, errors.New("user contact with user_id: " + userID + " and label: " + label + "not found")
	}
	if len(userAddress) > 0 &&
		(userAddress[0].GetStatus() == pbKyc.Status_STATUS_UNSPECIFIED ||
			userAddress[0].GetStatus() == pbKyc.Status_STATUS_CANCELED) {
		return nil, errors.New("user contact with user_id: " + userID + " and label: " + label + "is deleted or not active")
	}
	// get old contact detail
	addressDetail, getaddressDetailErr := s.addressSS.Get(ctx, &connect.Request[pbAddresses.GetRequest]{
		Msg: &pbAddresses.GetRequest{
			Id: userAddress[0].GetAddress().GetId(),
		},
	})
	if getaddressDetailErr != nil {
		return nil, getaddressDetailErr
	}
	userAddress[0].Address = addressDetail.Msg.GetAddress()

	return userAddress[0], nil
}

func (s *ServiceServer) UpdateAddress(ctx context.Context, req *connect.Request[pbUserIDs.UpdateAddressRequest]) (
	*connect.Response[pbUserIDs.UpdateAddressResponse], error,
) {
	var (
		userID              string
		err                 error
		updateUserAddressQB *util.QueryBuilder
		updateAddressQB     *util.QueryBuilder
	)
	if errCode, validateUpdateErr := s.ValidateUpdateAddress(req.Msg); validateUpdateErr != nil {
		return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
			Response: &pbUserIDs.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCode,
					Package: _package,
					Text:    validateUpdateErr.Error(),
				},
			},
		}), validateUpdateErr
	}
	err = s.scanForExistUser(ctx, req.Msg.GetUser(), &userID)
	if err != nil {
		return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
			Response: &pbUserIDs.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	// get old user address to update
	oldUserAddress, err := s.GetUserAddressByIDLabel(
		ctx, userID, req.Msg.GetAddress().GetByLabel())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
			Response: &pbUserIDs.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	isUpdateable, isSubAddressUpdateable := s.userAddressRepo.CheckUpdateability(oldUserAddress, req.Msg.GetAddress())
	if !isUpdateable {
		return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
			Response: &pbUserIDs.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    "address is not updateable with provided data",
				},
			},
		}), errors.New("address is not updateable with provided data")
	}
	updateAddressReq := req.Msg.GetAddress()
	updateUserAddressQB, err = s.userAddressRepo.QbUpdate(userID, updateAddressReq)
	if err != nil {
		return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
			Response: &pbUserIDs.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	if isSubAddressUpdateable {
		updateAddressQB, err = s.addressRepo.QbUpdate(&pbAddresses.UpdateRequest{
			Id:         oldUserAddress.GetAddress().GetId(),
			Country:    updateAddressReq.GetAddress().Country,
			Status:     updateAddressReq.GetAddress().Status,
			Type:       updateAddressReq.GetAddress().Type,
			Building:   updateAddressReq.GetAddress().Building,
			Floor:      updateAddressReq.GetAddress().Floor,
			Unit:       updateAddressReq.GetAddress().Unit,
			StreetNum:  updateAddressReq.GetAddress().StreetNum,
			StreetName: updateAddressReq.GetAddress().StreetName,
			District:   updateAddressReq.GetAddress().District,
			Locality:   updateAddressReq.GetAddress().Locality,
			ZipCode:    updateAddressReq.GetAddress().ZipCode,
			Region:     updateAddressReq.GetAddress().Region,
			State:      updateAddressReq.GetAddress().State,
		})
		if err != nil {
			return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
				Response: &pbUserIDs.UpdateAddressResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), err
		}
	}
	updateUserAddressSQLStr, updateUserAddressSQLArgs, _ := updateUserAddressQB.GenerateSQL()
	updateUserAddress := &pbUserIDs.Address{
		User:    req.Msg.GetUser(),
		Address: &pbAddresses.LabeledAddress{},
	}
	pgxTx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
			Response: &pbUserIDs.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	defer func() {
		if _err := pgxTx.Rollback(ctx); _err != nil {
			log.Err(_err)
		}
	}()
	updatedlabelAddress, updatelabelAddressErr := common.TxBulkWrite[pbAddresses.LabeledAddress](
		ctx, pgxTx, updateUserAddressSQLStr, updateUserAddressSQLArgs, s.userAddressRepo.ScanRow)
	if updatelabelAddressErr != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"updating", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
			Response: &pbUserIDs.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    updatelabelAddressErr.Error() + "(" + _err.Error() + ")",
				},
			},
		}), updatelabelAddressErr
	}
	updateUserAddress.Address = updatedlabelAddress[0]
	updateUserAddress.Address.Address = oldUserAddress.GetAddress()
	if !reflect.ValueOf(updateAddressQB).IsNil() {
		updateAddressSQLStr, updateContractSQLArgs, _ := updateAddressQB.GenerateSQL()
		updatedAddress, updateAddressErr := common.TxWrite[pbAddresses.Address](
			ctx, pgxTx, updateAddressSQLStr, updateContractSQLArgs, s.addressRepo.ScanRow)
		if updateAddressErr != nil {
			_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
				"updating", _entityName, "user_id/login = "+req.Msg.GetUser().String())
			return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
				Response: &pbUserIDs.UpdateAddressResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    _err.Error() + "(" + updateAddressErr.Error() + ")",
					},
				},
			}), updateAddressErr
		}
		updateUserAddress.Address.Address = updatedAddress
	}
	if err = pgxTx.Commit(ctx); err != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"updating", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
			Response: &pbUserIDs.UpdateAddressResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	return connect.NewResponse(&pbUserIDs.UpdateAddressResponse{
		Response: &pbUserIDs.UpdateAddressResponse_Address{
			Address: updateUserAddress,
		},
	}), nil
}

func (s *ServiceServer) RemoveAddresses(
	ctx context.Context,
	req *connect.Request[pbUserIDs.RemoveAddressesRequest],
) (*connect.Response[pbUserIDs.RemoveAddressesResponse], error) {
	var (
		row pgx.Rows
		err error
		qb  *util.QueryBuilder
	)

	// Verify that User exists in core.users
	var userAddress string
	row, err = s.queryUserID(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.RemoveAddressesResponse{
			Response: &pbUserIDs.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _addressPackage,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"deleting", _entityName, "user_id/login="+req.Msg.GetUser().String()),
				},
			},
		}), err
	}
	_err := s.CheckRow(row, userAddress)
	if _err != nil {
		return connect.NewResponse(&pbUserIDs.RemoveAddressesResponse{
			Response: &pbUserIDs.RemoveAddressesResponse_Error{
				Error: _err,
			},
		}), err
	}

	qb = s.repo.QbGetOne(userAddress, _addressTableName, _userAddressFields, req.Msg.GetLabels().GetList())
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveAddressesResponse{
			Response: &pbUserIDs.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _addressPackage,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	defer rows.Close()

	userAddressesRes, err := s.userAddressRepo.ScanMultiRows(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveAddressesResponse{
			Response: &pbUserIDs.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _addressPackage,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	if len(userAddressesRes.GetList()) == 0 {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveAddressesResponse{
			Response: &pbUserIDs.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _addressPackage,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	ids := []string{}
	list := userAddressesRes.GetList()
	for _, c := range list {
		ids = append(ids, c.GetAddress().GetId())
	}

	qb, err = s.repo.QbDelete(userAddress, _addressTableName, req.Msg.GetLabels().GetList())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserIDs.RemoveAddressesResponse{
			Response: &pbUserIDs.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _addressPackage,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"deleting", _entityName, "user_id/login="+req.Msg.GetUser().String()),
				},
			},
		}), err
	}
	sqlstr, args, _ := qb.SetReturnFields("*").GenerateSQL()

	var errScan error
	log.Info().Msg("Executing SQL '" + sqlstr + "'")
	if errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_qb, _err := s.addressRepo.QbDeleteMany(ids)
		if _err != nil {
			log.Error().Err(err)
			return _err
		}

		_sqlstr, _args, _ := _qb.SetReturnFields("*").GenerateSQL()
		log.Info().Msg("Executing SQL '" + _sqlstr + "'")
		if row, err = tx.Query(ctx, _sqlstr, _args...); err != nil {
			return err
		}
		row.Close()

		if rows, err = tx.Query(ctx, sqlstr, args...); err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			if _, errScan = s.userAddressRepo.ScanMultiRows(rows); errScan != nil {
				log.Error().Err(err).Msgf("unable to delete %s with id/login = '%s'",
					_entityName, req.Msg.GetUser().String())
				return errScan
			}
			log.Info().Msgf("%s with id/login = '%s' removed successfully",
				_entityName, req.Msg.GetUser().String())
			return nil
		} else {
			return fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND)],
				_entityName, "id/login="+req.Msg.GetUser().String())
		}
	}); errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"remove", _entityName, "user_id/login = "+req.Msg.GetUser().String())
		log.Error().Err(errTx).Msg(_err.Error())
		return connect.NewResponse(&pbUserIDs.RemoveAddressesResponse{
			Response: &pbUserIDs.RemoveAddressesResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _addressPackage,
					Text:    _err.Error() + "(" + _err.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserIDs.RemoveAddressesResponse{
		Response: &pbUserIDs.RemoveAddressesResponse_Addresses{
			Addresses: &pbUserIDs.AddressList{
				User:      req.Msg.GetUser(),
				Addresses: userAddressesRes,
			},
		},
	}), nil
}

func (s *ServiceServer) GetUser(
	ctx context.Context, userID string, table string, fields string,
) (pgx.Rows, error) {
	qb := s.repo.QbGetOne(userID, table, fields, []string{})
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return nil, err
	}
	return rows, nil
}

func (s *ServiceServer) GetAddresses(
	ctx context.Context, userID string,
) (*pbAddresses.LabeledAddressList, error) {
	rows, err := s.GetUser(ctx, userID, _addressTableName, _userAddressFields)
	if err != nil {
		return nil, err
	}
	userAddressesRes, err := s.userAddressRepo.ScanMultiRows(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName)
		log.Error().Err(err).Msg(_err.Error())
		return nil, err
	}
	if len(userAddressesRes.GetList()) == 0 {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName)
		log.Error().Err(err).Msg(_err.Error())
		return nil, _err
	}

	list := userAddressesRes.GetList()
	newList := []*pbAddresses.LabeledAddress{}

	for _, c := range list {
		id := c.GetAddress().GetId()
		addressRes, addressErr := s.addressSS.Get(ctx, connect.NewRequest[pbAddresses.GetRequest](&pbAddresses.GetRequest{Id: id}))
		if addressErr == nil {
			c.Address = addressRes.Msg.GetAddress()
			newList = append(newList, c)
		}
	}
	userAddressesRes.List = newList

	return userAddressesRes, nil
}
