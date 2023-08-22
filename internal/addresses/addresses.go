package addresses

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbAddresses "davensi.com/core/gen/addresses"
	pbAddressesconnect "davensi.com/core/gen/addresses/addressesconnect"
	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/countries"
)

const (
	_package          = "addresses"
	_entityName       = "Address"
	_entityNamePlural = "addresses"
)

// For singleton Addresses export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

// ServiceServer implements the AddressesService API
type ServiceServer struct {
	Repo AddressRepository
	pbAddressesconnect.ServiceHandler
	db             *pgxpool.Pool
	countryService countries.ServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:           *NewAddressRepository(db),
		db:             db,
		countryService: *countries.GetSingletonServiceServer(db),
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
	req *connect.Request[pbAddresses.CreateRequest],
) (*connect.Response[pbAddresses.CreateResponse], error) {
	if errCreation := s.validateCreate(req.Msg); errCreation != nil {
		log.Error().Err(errCreation.Err)
		return connect.NewResponse(&pbAddresses.CreateResponse{
			Response: &pbAddresses.CreateResponse_Error{
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
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbAddresses.CreateResponse{
			Response: &pbAddresses.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlStr, args, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	newAddress, err := common.ExecuteTxWrite(
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanRow,
	)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbAddresses.CreateResponse{
			Response: &pbAddresses.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s created successfully with id = %s",
		_entityName, newAddress.Id)
	return connect.NewResponse(&pbAddresses.CreateResponse{
		Response: &pbAddresses.CreateResponse_Address{
			Address: newAddress,
		},
	}), nil
}

func (s *ServiceServer) ExecuteInsertManyAddresses(
	ctx context.Context,
	tx pgx.Tx,
	msg *pbAddresses.SetLabeledAddressList,
	insertAddresses []*pbAddresses.SetLabeledAddress,
) (map[string]*pbAddresses.SetLabeledAddress, error) {
	qbInsertAddresses, err := s.Repo.QbInsertManyAddress(msg)
	res := map[string]*pbAddresses.SetLabeledAddress{}
	if err != nil {
		return res, err
	}
	var rows pgx.Rows
	sqlInsertAddresses, argsInsertAddress, _ := qbInsertAddresses.GenerateSQL()
	log.Log().Msgf("sql %s with args %v", sqlInsertAddresses, argsInsertAddress)
	if tx != nil {
		rows, err = tx.Query(ctx, sqlInsertAddresses, argsInsertAddress...)
		if err != nil {
			return nil, err
		}
	} else {
		rows, err = s.db.Query(ctx, sqlInsertAddresses, argsInsertAddress...)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()
	countInsert := 0
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		res[id] = insertAddresses[countInsert]
		countInsert++
	}
	return res, nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbAddresses.UpdateRequest],
) (*connect.Response[pbAddresses.UpdateResponse], error) {
	commonErr := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
		"updating",
		_entityName,
		"",
	)

	if errQueryUpdate := s.validateUpdateQuery(req.Msg); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbAddresses.UpdateResponse{
			Response: &pbAddresses.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	err := s.getOldAddressToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbAddresses.UpdateResponse{
			Response: &pbAddresses.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
					Package: _package,
					Text:    "update failed: address does not exist for that ID",
				},
			},
		}), err
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		commonErr.
			UpdateCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR).
			UpdateMessage(genSQLError.Error())
		log.Error().Err(commonErr.Err)
		return connect.NewResponse(&pbAddresses.UpdateResponse{
			Response: &pbAddresses.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    commonErr.Code,
					Package: _package,
					Text:    commonErr.Err.Error(),
				},
			},
		}), commonErr.Err
	}

	sqlStr, args, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	updatedAddress, err := common.ExecuteTxWrite[pbAddresses.Address](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanRow,
	)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			sel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbAddresses.UpdateResponse{
			Response: &pbAddresses.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s updated successfully with id = %s",
		_entityName, updatedAddress.Id)
	return connect.NewResponse(&pbAddresses.UpdateResponse{
		Response: &pbAddresses.UpdateResponse_Address{
			Address: updatedAddress,
		},
	}), nil
}

func (s *ServiceServer) getOldAddressToUpdate(msg *pbAddresses.UpdateRequest) error {
	getAddressRequest := &pbAddresses.GetRequest{Id: msg.GetId()}

	_, err := s.Get(context.Background(), &connect.Request[pbAddresses.GetRequest]{
		Msg: getAddressRequest,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbAddresses.GetRequest],
) (*connect.Response[pbAddresses.GetResponse], error) {
	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbAddresses.GetResponse{
			Response: &pbAddresses.GetResponse_Error{
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
		return connect.NewResponse(&pbAddresses.GetResponse{
			Response: &pbAddresses.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	address, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbAddresses.GetResponse{
			Response: &pbAddresses.GetResponse_Error{
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
		return connect.NewResponse(&pbAddresses.GetResponse{
			Response: &pbAddresses.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbAddresses.GetResponse{
		Response: &pbAddresses.GetResponse_Address{
			Address: address,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbAddresses.GetListRequest],
	res *connect.ServerStream[pbAddresses.GetListResponse],
) error {
	if req.Msg.Country == nil {
		req.Msg.Country = &pbCountries.GetListRequest{}
	}
	qbCountries := s.countryService.Repo.QbGetList(req.Msg.Country)
	countryFB, countryArgs := qbCountries.Filters.GenerateSQL()

	qb := s.Repo.QbGetList(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON countries.id = addresses.country_id", qbCountries.TableName)).
		Select(strings.Join(qbCountries.SelectFields, ", ")).
		Where(countryFB, countryArgs...)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbAddresses.GetListResponse{
					Response: &pbAddresses.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		address, err := s.Repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbAddresses.GetListResponse{
						Response: &pbAddresses.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbAddresses.GetListResponse{
			Response: &pbAddresses.GetListResponse_Address{
				Address: address,
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
	req *connect.Request[pbAddresses.DeleteRequest],
) (*connect.Response[pbAddresses.DeleteResponse], error) {
	terminatedStatus := pbCommon.Status_STATUS_TERMINATED
	address := &pbAddresses.UpdateRequest{
		Id:     req.Msg.GetId(),
		Status: &terminatedStatus,
	}
	deleteReq := connect.NewRequest(address)
	deletedAddress, err := s.Update(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbAddresses.DeleteResponse{
			Response: &pbAddresses.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedAddress.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedAddress.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbAddresses.DeleteResponse{
		Response: &pbAddresses.DeleteResponse_Address{
			Address: deletedAddress.Msg.GetAddress(),
		},
	}), nil
}

func (s *ServiceServer) GenCreateFunc(req *pbAddresses.CreateRequest, addressUUID string) (
	func(tx pgx.Tx) (*pbAddresses.Address, error), *common.ErrWithCode,
) {
	errGenFn := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR, "creating", _entityName, "")

	if validateErr := s.validateCreate(req); validateErr != nil {
		return nil, validateErr
	}

	qb, errInsert := s.Repo.QbInsertWithUUID(req, addressUUID)
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
	return func(tx pgx.Tx) (*pbAddresses.Address, error) {
		executedAddress, errWriteAddress := common.TxWrite[pbAddresses.Address](
			context.Background(),
			tx,
			sqlStr,
			args,
			s.Repo.ScanRow,
		)

		if errWriteAddress != nil {
			return nil, errWriteAddress
		}

		return executedAddress, nil
	}, nil
}

func (s *ServiceServer) GenUpdateFunc(req *pbAddresses.UpdateRequest) (
	updateFn func(tx pgx.Tx) (*pbAddresses.Address, error), sel string, errorWithCode *common.ErrWithCode,
) {
	commonErr := common.CreateErrWithCode(pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, "updating", _entityName, "")

	if errQueryUpdate := s.validateUpdateQuery(req); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		commonErr.UpdateCode(errQueryUpdate.Code).UpdateMessage(errQueryUpdate.Err.Error())
		return nil, "", commonErr
	}

	err := s.getOldAddressToUpdate(req)
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
	return func(tx pgx.Tx) (*pbAddresses.Address, error) {
		return common.TxWrite(
			context.Background(),
			tx,
			sqlStr,
			args,
			s.Repo.ScanRow,
		)
	}, sel, nil
}
