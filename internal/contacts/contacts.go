package contacts

import (
	"context"
	"fmt"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	pbContactsconnect "davensi.com/core/gen/contacts/contactsconnect"
	"davensi.com/core/internal/common"
)

const (
	_package          = "contacts"
	_entityName       = "Contact"
	_entityNamePlural = "Contacts"
)

var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

// ServiceServer implements the ContactsService API
type ServiceServer struct {
	Repo ContactRepository
	pbContactsconnect.ServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewContactRepository(db),
		db:   db,
	}
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbContacts.CreateRequest],
) (*connect.Response[pbContacts.CreateResponse], error) {
	if errCreation := s.ValidateCreate(req.Msg); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbContacts.CreateResponse{
			Response: &pbContacts.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	qb, err := s.Repo.QbInsert(req.Msg)
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"create",
			_package,
			err.Error(),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbContacts.CreateResponse{
			Response: &pbContacts.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	newContact, err := common.ExecuteTxWrite(ctx, s.db, sqlStr, args, s.Repo.ScanRow)
	if err != nil {
		errCreation := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"creating",
			_package,
			fmt.Sprintf(
				"value = '%s'",
				req.Msg.GetValue(),
			),
		)
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbContacts.CreateResponse{
			Response: &pbContacts.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreation.Code,
					Package: _package,
					Text:    errCreation.Err.Error(),
				},
			},
		}), errCreation.Err
	}

	log.Info().Msgf("%s created successfully with id = %s",
		_entityName, newContact.Id)
	return connect.NewResponse(&pbContacts.CreateResponse{
		Response: &pbContacts.CreateResponse_Contact{
			Contact: newContact,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbContacts.UpdateRequest],
) (*connect.Response[pbContacts.UpdateResponse], error) {
	if errQueryUpdate := validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbContacts.UpdateResponse{
			Response: &pbContacts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getContactResponse, err := s.getOldContactToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbContacts.UpdateResponse{
			Response: &pbContacts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getContactResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getContactResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		errGenSQL := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			err.Error(),
		)
		log.Error().Err(errGenSQL.Err)
		return connect.NewResponse(&pbContacts.UpdateResponse{
			Response: &pbContacts.UpdateResponse_Error{
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
	updatedContact, err := common.ExecuteTxWrite[pbContacts.Contact](
		ctx,
		s.db,
		sqlstr,
		sqlArgs,
		s.Repo.ScanRow,
	)
	if err != nil {
		errUpdate := common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			"updating",
			_package,
			fmt.Sprintf("%s, %s", sel, err.Error()),
		)
		log.Error().Err(errUpdate.Err)
		return connect.NewResponse(&pbContacts.UpdateResponse{
			Response: &pbContacts.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errUpdate.Code,
					Package: _package,
					Text:    errUpdate.Err.Error(),
				},
			},
		}), errUpdate.Err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbContacts.UpdateResponse{
		Response: &pbContacts.UpdateResponse_Contact{
			Contact: updatedContact,
		},
	}), nil
}

func (s *ServiceServer) getOldContactToUpdate(msg *pbContacts.UpdateRequest) (*connect.Response[pbContacts.GetResponse], error) {
	getContactRequest := &pbContacts.GetRequest{Id: msg.GetId()}

	return s.Get(context.Background(), &connect.Request[pbContacts.GetRequest]{
		Msg: getContactRequest,
	})
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbContacts.GetRequest],
) (*connect.Response[pbContacts.GetResponse], error) {
	if errQueryGet := validateQueryGet(req.Msg); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbContacts.GetResponse{
			Response: &pbContacts.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbContacts.GetResponse{
			Response: &pbContacts.GetResponse_Error{
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
		return connect.NewResponse(&pbContacts.GetResponse{
			Response: &pbContacts.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	contact, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbContacts.GetResponse{
			Response: &pbContacts.GetResponse_Error{
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
		return connect.NewResponse(&pbContacts.GetResponse{
			Response: &pbContacts.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbContacts.GetResponse{
		Response: &pbContacts.GetResponse_Contact{
			Contact: contact,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbContacts.GetListRequest],
	res *connect.ServerStream[pbContacts.GetListResponse],
) error {
	qb := s.Repo.QbGetList(req.Msg)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbContacts.GetListResponse{
					Response: &pbContacts.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		contact, err := s.Repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbContacts.GetListResponse{
						Response: &pbContacts.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbContacts.GetListResponse{
			Response: &pbContacts.GetListResponse_Contact{
				Contact: contact,
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
	req *connect.Request[pbContacts.DeleteRequest],
) (*connect.Response[pbContacts.DeleteResponse], error) {
	terminatedStatus := pbCommon.Status_STATUS_TERMINATED
	contact := &pbContacts.UpdateRequest{
		Id:     req.Msg.GetId(),
		Status: &terminatedStatus,
	}
	deleteReq := connect.NewRequest(contact)
	deletedContact, err := s.Update(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbContacts.DeleteResponse{
			Response: &pbContacts.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedContact.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedContact.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbContacts.DeleteResponse{
		Response: &pbContacts.DeleteResponse_Contact{
			Contact: deletedContact.Msg.GetContact(),
		},
	}), nil
}

func (s *ServiceServer) GenCreateFunc(req *pbContacts.CreateRequest, contactUUID string) (
	func(tx pgx.Tx) (*pbContacts.Contact, error), *common.ErrWithCode,
) {
	errGenFn := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR, "creating", _entityName, "")

	if validateErr := s.ValidateCreate(req); validateErr != nil {
		return nil, validateErr
	}

	qb, errInsert := s.Repo.QbInsertWithUUID(req, contactUUID)
	if errInsert != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			errInsert.Error(),
		)
		log.Error().Err(_err)
		return nil, errGenFn.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT).
			UpdateMessage(errInsert.Error())
	}

	sqlStr, args, _ := qb.SetReturnFields("*").GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	return func(tx pgx.Tx) (*pbContacts.Contact, error) {
		executedContact, errWriteContact := common.TxWrite[pbContacts.Contact](
			context.Background(),
			tx,
			sqlStr,
			args,
			s.Repo.ScanRow,
		)

		if errWriteContact != nil {
			return nil, errWriteContact
		}

		return executedContact, nil
	}, nil
}

func (s *ServiceServer) GenUpdateFunc(req *pbContacts.UpdateRequest) (
	updateFn func(tx pgx.Tx) (*pbContacts.Contact, error), sel string, errorWithCode *common.ErrWithCode,
) {
	commonErr := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, "updating", _entityName, "")

	if errQueryUpdate := validateQueryUpdate(req); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		commonErr.UpdateCode(errQueryUpdate.Code).UpdateMessage(errQueryUpdate.Err.Error())
		return nil, "", commonErr
	}

	_, err := s.getOldContactToUpdate(req)
	if err != nil {
		log.Error().Err(err)
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).UpdateMessage(err.Error())
		return nil, "", commonErr
	}

	qb, genSQLError := s.Repo.QbUpdate(req)
	if genSQLError != nil {
		commonErr.UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).UpdateMessage(genSQLError.Error())
		log.Error().Err(commonErr.Err)
		return nil, "", commonErr
	}

	sqlStr, args, sel := qb.SetReturnFields("*").GenerateSQL()

	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	return func(tx pgx.Tx) (*pbContacts.Contact, error) {
		return common.TxWrite(
			context.Background(),
			tx,
			sqlStr,
			args,
			s.Repo.ScanRow,
		)
	}, sel, nil
}
