package recipients

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbOrgs "davensi.com/core/gen/orgs"
	pbRecipients "davensi.com/core/gen/recipients"
	pbRecipientsConnect "davensi.com/core/gen/recipients/recipientsconnect"
	pbUsers "davensi.com/core/gen/users"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/legalentities"
	"davensi.com/core/internal/orgs"
	"davensi.com/core/internal/users"
	"davensi.com/core/internal/util"
)

const (
	_fields           = "id, legalentity_id, user_id, label, type, org_id, status"
	_package          = "recipients"
	_entityName       = "Recipient"
	_entityNamePlural = "Recipients"
)

// ServiceServer implements the RecipientsService API
type ServiceServer struct {
	Repo Repository
	pbRecipientsConnect.UnimplementedServiceHandler
	db              *pgxpool.Pool
	legalEntitiesSS *legalentities.ServiceServer
	usersSS         *users.ServiceServer
	orgsSS          *orgs.ServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:            *NewRecipientsRepository(db),
		db:              db,
		legalEntitiesSS: legalentities.GetSingletonServiceServer(db),
		usersSS:         users.GetSingletonServiceServer(db),
		orgsSS:          orgs.GetSingletonServiceServer(db),
	}
}

// For singleton Recipient export module
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
	req *connect.Request[pbRecipients.GetRequest],
) (*connect.Response[pbRecipients.GetResponse], error) {
	// Validate input
	pkRes, validateErr := s.ValidateGet(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbRecipients.GetResponse{
			Response: &pbRecipients.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}
	type magicType string
	var magicKey magicType = "magicValue"
	// Query database
	var qb *util.QueryBuilder
	if ctx == context.WithValue(context.Background(), magicKey, magicKey) {
		qb = s.Repo.QbGetOne(req.Msg, pkRes, false, false)
	} else {
		qb = s.Repo.QbGetOne(req.Msg, pkRes, true, false)
	}
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL " + sqlstr + req.Msg.GetSelect().GetById() + " with args " + fmt.Sprint(sqlArgs))
	rows, errQuery := s.db.Query(ctx, sqlstr, sqlArgs...)
	if errQuery != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(errQuery).Msg(_err.Error())
		return connect.NewResponse(&pbRecipients.GetResponse{
			Response: &pbRecipients.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + " (" + errQuery.Error() + ")",
				},
			},
		}), errQuery
	}

	defer rows.Close()
	// Check if no record
	if !rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityName, sel)
		log.Error().Err(_err).Msg(_err.Error())
		return connect.NewResponse(&pbRecipients.GetResponse{
			Response: &pbRecipients.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	var recipient *pbRecipients.Recipient
	var errScan error
	if ctx == context.WithValue(context.Background(), magicKey, magicKey) {
		recipient, errScan = s.Repo.ScanRow(rows)
	} else {
		recipient, errScan = s.Repo.SuperScanRow(rows)
	}
	if errScan != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(errScan).Msg(_err.Error())
		return connect.NewResponse(&pbRecipients.GetResponse{
			Response: &pbRecipients.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + " (" + errScan.Error() + ")",
				},
			},
		}), errScan
	}
	// Check if one more record
	if rows.Next() {
		_errno := pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, sel)
		log.Error().Err(_err).Msg(_err.Error())
		return connect.NewResponse(&pbRecipients.GetResponse{
			Response: &pbRecipients.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	// Building the response from here
	return connect.NewResponse(&pbRecipients.GetResponse{
		Response: &pbRecipients.GetResponse_Recipient{
			Recipient: recipient,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbRecipients.GetListRequest],
	res *connect.ServerStream[pbRecipients.GetListResponse],
) error {
	if req.Msg.LegalEntity == nil {
		req.Msg.LegalEntity = &pbLegalEntities.GetListRequest{}
	}

	if req.Msg.User == nil {
		req.Msg.User = &pbUsers.GetListRequest{}
	}

	if req.Msg.Org == nil {
		req.Msg.Org = &pbOrgs.GetListRequest{}
	}

	var (
		qbLegalEntity = s.legalEntitiesSS.Repo.QbGetList(req.Msg.LegalEntity)
		qbUser        = s.usersSS.Repo.QbGetList(req.Msg.User)
		qbOrg         = s.orgsSS.Repo.QbGetList(req.Msg.Org)

		filterLegalEntityStr, filterLegalEntityArgs = qbLegalEntity.Filters.GenerateSQL()
		filterUserStr, filterUserArgs               = qbUser.Filters.GenerateSQL()
		filterOrgStr, filterOrgArgs                 = qbOrg.Filters.GenerateSQL()
	)

	qb := s.Repo.QbGetList(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON recipients.legalentity_id = legalentities.id", qbLegalEntity.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON recipients.user_id = users.id", qbUser.TableName)).
		Join(fmt.Sprintf("LEFT JOIN %s ON recipients.org_id  = orgs.id", qbOrg.TableName)).
		Select(strings.Join(qbLegalEntity.SelectFields, ", ")).
		Select(strings.Join(qbUser.SelectFields, ", ")).
		Select(strings.Join(qbOrg.SelectFields, ", ")).
		Where(filterLegalEntityStr, filterLegalEntityArgs...).
		Where(filterUserStr, filterUserArgs...).
		Where(filterOrgStr, filterOrgArgs...)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)
	}

	defer rows.Close()

	hasRows := false
	for rows.Next() {
		hasRows = true
		// Scan recipient field
		recipient, errScan := s.Repo.ScanWithRelationships(rows)
		if errScan != nil {
			return streamingErr(res, errScan, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}
		errSend := res.Send(&pbRecipients.GetListResponse{
			Response: &pbRecipients.GetListResponse_Recipient{
				Recipient: recipient,
			},
		})
		if errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityNamePlural, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	if !hasRows { // If there are no match
		if !rows.Next() {
			_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, args)
			log.Error().Err(err).Msg(_err.Error())
			errSend := res.Send(&pbRecipients.GetListResponse{
				Response: &pbRecipients.GetListResponse_Error{
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
			return nil
		}
	}

	return rows.Err()
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbRecipients.CreateRequest],
) (*connect.Response[pbRecipients.CreateResponse], error) {
	PKRes, validateErr := s.ValidateCreate(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbRecipients.CreateResponse{
			Response: &pbRecipients.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	qb, err := s.Repo.QbInsert(req.Msg, PKRes)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], err.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbRecipients.CreateResponse{
			Response: &pbRecipients.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	newRecipient, insertErr := common.ExecuteTxWrite[pbRecipients.Recipient](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanRow,
	)
	if insertErr != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating", _entityName, "label = '"+req.Msg.GetLabel()+"'")
		log.Error().Err(insertErr).Msg(_err.Error())
		return connect.NewResponse(&pbRecipients.CreateResponse{
			Response: &pbRecipients.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + insertErr.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with label = '%s' created successfully with id = %s",
		_entityName, req.Msg.GetLabel(), newRecipient.GetId())
	return connect.NewResponse(&pbRecipients.CreateResponse{
		Response: &pbRecipients.CreateResponse_Recipient{
			Recipient: newRecipient,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbRecipients.UpdateRequest],
) (*connect.Response[pbRecipients.UpdateResponse], error) {
	pkResOld, pkResNew, validateErr := s.ValidateUpdate(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbRecipients.UpdateResponse{
			Response: &pbRecipients.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg, pkResOld, pkResNew)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbRecipients.UpdateResponse{
			Response: &pbRecipients.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlStr, args, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL " + sqlStr + "")
	updateRecipient, updateErr := common.ExecuteTxWrite[pbRecipients.Recipient](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanRow,
	)
	if updateErr != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "updating", _entityName, sel)
		log.Error().Err(updateErr).Msg(_err.Error())
		return connect.NewResponse(&pbRecipients.UpdateResponse{
			Response: &pbRecipients.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + updateErr.Error() + ")",
				},
			},
		}), updateErr
	}

	// Start building the response from here
	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbRecipients.UpdateResponse{
		Response: &pbRecipients.UpdateResponse_Recipient{
			Recipient: updateRecipient,
		},
	}), nil
}

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbRecipients.DeleteRequest],
) (*connect.Response[pbRecipients.DeleteResponse], error) {
	statusTerminated := pbCommon.Status_STATUS_TERMINATED

	recipientInput := &pbRecipients.UpdateRequest{
		Select: req.Msg.GetSelect(),
		Status: &statusTerminated,
	}

	updateRes, err := s.Update(ctx, connect.NewRequest(recipientInput))
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		log.Error().Err(err).Msg(err.Error())
		return connect.NewResponse(&pbRecipients.DeleteResponse{
			Response: &pbRecipients.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s deleted successfully", _entityName, "id = "+updateRes.Msg.GetRecipient().Id)
	return connect.NewResponse(&pbRecipients.DeleteResponse{
		Response: &pbRecipients.DeleteResponse_Recipient{
			Recipient: updateRes.Msg.GetRecipient(),
		},
	}), nil
}

func streamingErr(
	res *connect.ServerStream[pbRecipients.GetListResponse],
	err error, errorCode pbCommon.ErrorCode,
) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbRecipients.GetListResponse{
		Response: &pbRecipients.GetListResponse_Error{
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

func (s *ServiceServer) MakeCreationQB(msg *pbRecipients.CreateRequest) (*util.QueryBuilder, *common.ErrWithCode) {
	PKRes, validateErr := s.ValidateCreate(msg)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return nil, validateErr
	}

	qb, err := s.Repo.QbInsert(msg, PKRes)
	if err != nil {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"creating",
			_entityName,
			err.Error(),
		)
	}
	return qb, nil
}

func (s *ServiceServer) MakeUpdateQB(msg *pbRecipients.UpdateRequest) (*util.QueryBuilder, *common.ErrWithCode) {
	pkResOld, pkResNew, validateErr := s.ValidateUpdate(msg)
	if validateErr != nil {
		return nil, validateErr
	}

	qb, qbUpdateErr := s.Repo.QbUpdate(msg, pkResOld, pkResNew)
	if qbUpdateErr != nil {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating",
			_entityName,
			qbUpdateErr.Error(),
		)
	}
	return qb, nil
}

func (s *ServiceServer) GenCreateFunc(req *pbRecipients.CreateRequest, recipientUUID string) (
	func(tx pgx.Tx) (*pbRecipients.Recipient, error), *common.ErrWithCode,
) {
	errGenFn := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR, "creating", _entityName, "")

	PKRes, validateErr := s.ValidateCreate(req)
	if validateErr != nil {
		log.Error().Err(validateErr.Err)
		return nil, validateErr
	}

	qb, err := s.Repo.QbInsertWithUUID(req, PKRes, recipientUUID)
	if err != nil {
		errGenFn.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)
		errGenFn.UpdateMessage(fmt.Errorf(common.Errors[uint32((pbCommon.ErrorCode_ERROR_CODE_DB_ERROR))], err.Error()).Error())
		log.Error().Err(errGenFn.Err)
		return nil, errGenFn
	}

	sqlStr, args, _ := qb.SetReturnFields("*").GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	return func(tx pgx.Tx) (*pbRecipients.Recipient, error) {
		executedRecipient, errWriteRecipient := common.TxWrite[pbRecipients.Recipient](
			context.Background(),
			tx,
			sqlStr,
			args,
			s.Repo.ScanRow,
		)

		if errWriteRecipient != nil {
			return nil, errWriteRecipient
		}

		return executedRecipient, nil
	}, nil
}

func (s *ServiceServer) GenUpdateFunc(req *pbRecipients.UpdateRequest) (
	updateFn func(tx pgx.Tx) (*pbRecipients.Recipient, error), sel string, errorWithCode *common.ErrWithCode,
) {
	commonErr := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, "updating", _entityName, "")

	pkResOld, pkResNew, errQueryUpdate := s.ValidateUpdate(req)
	if errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		commonErr.UpdateCode(errQueryUpdate.Code).UpdateMessage(errQueryUpdate.Err.Error())
		return nil, "", commonErr
	}

	qb, genSQLError := s.Repo.QbUpdate(req, pkResOld, pkResNew)

	return func(tx pgx.Tx) (*pbRecipients.Recipient, error) {
		if genSQLError == nil { // Execute Update SQL if there is something to update
			sqlStr, args, _ := qb.GenerateSQL()
			log.Info().Msg("Executing SQL " + sqlStr + "")
			return common.TxWrite(
				context.Background(),
				tx,
				sqlStr,
				args,
				s.Repo.ScanRow,
			)
		} else {
			// Else only SELECT to get the recipient record's info
			// (needed for updating only bankaccount table without updating recipient table)
			qb := s.Repo.QbGetOne(&pbRecipients.GetRequest{
				Select: req.Select,
			}, pkResOld, false, false)
			sqlStr, args, _ := qb.GenerateSQL()
			log.Info().Msg("Executing SQL " + sqlStr + "")
			return common.TxWrite(
				context.Background(),
				tx,
				sqlStr,
				args,
				s.Repo.ScanRow,
			)
		}
	}, sel, nil
}
