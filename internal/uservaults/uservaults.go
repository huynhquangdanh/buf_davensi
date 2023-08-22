package uservaults

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbUsers "davensi.com/core/gen/users"
	pbUserVaults "davensi.com/core/gen/uservaults"
	pbUserVaultsconnect "davensi.com/core/gen/uservaults/uservaultsconnect"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/util"
)

const (
	_package          = "uservaults"
	_entityName       = "User Vault"
	_entityNamePlural = "User Vaults"
	_fields           = "user_id, key, value_type, data, status"
	_table            = "core.uservaults"
)

// ServiceServer implements the UserVaults API
type ServiceServer struct {
	repo UserVaultsRepository
	pbUserVaultsconnect.ServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo: *NewUserVaultRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) Set(ctx context.Context, req *connect.Request[pbUserVaults.SetRequest],
) (*connect.Response[pbUserVaults.SetResponse], error) {
	// Verify that user.User, Key are specified and create hex input string and value type
	hexDataString, valueType, errno, validateErr := s.validateCreate(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr)
		return connect.NewResponse(&pbUserVaults.SetResponse{
			Response: &pbUserVaults.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errno,
					Package: _package,
					Text:    validateErr.Error(),
				},
			},
		}), validateErr
	}
	// Verify that User exists in core.users
	var (
		row    pgx.Rows
		err    error
		userID string
	)
	row, err = s.queryUser(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserVaults.SetResponse{
			Response: &pbUserVaults.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"setting", _entityName, "key="+req.Msg.GetKey()),
				},
			},
		}), err
	}
	if !row.Next() {
		err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT)],
			"setting "+_entityName, "user does not exist")
		return connect.NewResponse(&pbUserVaults.SetResponse{
			Response: &pbUserVaults.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	} else {
		if errScan := row.Scan(&userID); errScan != nil {
			err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)],
				"scanning", "core.users", "for id")
			return connect.NewResponse(&pbUserVaults.SetResponse{
				Response: &pbUserVaults.SetResponse_Error{
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

	// Building query depending on INSERT or UPDATE
	var qb *util.QueryBuilder
	row, err = s.queryUserVault(ctx, req)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserVaults.SetResponse{
			Response: &pbUserVaults.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"setting", _entityName, "key="+req.Msg.GetKey()),
				},
			},
		}), err
	}
	if !row.Next() {
		qb, err = s.repo.QbSetInsert(req.Msg, userID, hexDataString, valueType)
	} else {
		qb, err = s.repo.QbSetUpdate(req.Msg, userID, hexDataString, valueType)
	}
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserVaults.SetResponse{
			Response: &pbUserVaults.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"setting", _entityName, "key="+req.Msg.GetKey()),
				},
			},
		}), err
	}
	row.Close()

	sqlStr, args, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if row, err = tx.Query(ctx, sqlStr, args...); err != nil {
			return err
		}
		defer row.Close()
		return nil
	})
	if errTx != nil {
		err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"setting", _entityName, "key = "+req.Msg.GetKey())
		log.Error().Err(err).Msg(err.Error())
		return connect.NewResponse(&pbUserVaults.SetResponse{
			Response: &pbUserVaults.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error() + "(" + errTx.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserVaults.SetResponse{
		Response: &pbUserVaults.SetResponse_Ok{
			Ok: true,
		},
	}), nil
}

// Check if user exists for Set()
func (s *ServiceServer) queryUser(ctx context.Context, req *pbUsers.Select,
) (pgx.Rows, error) {
	var (
		row pgx.Rows
		err error
	)

	switch req.GetSelect().(type) {
	case *pbUsers.Select_ById:
		row, err = s.db.Query(ctx, "SELECT id FROM core.users WHERE id = $1", req.GetById())
	case *pbUsers.Select_ByLogin:
		row, err = s.db.Query(ctx, "SELECT id FROM core.users WHERE login = $1", req.GetByLogin())
	}

	return row, err
}

// Check if User Pref exists for Set() => INSERT or UPDATE
func (s *ServiceServer) queryUserVault(ctx context.Context, req *connect.Request[pbUserVaults.SetRequest],
) (pgx.Rows, error) {
	var (
		row pgx.Rows
		err error
	)

	switch req.Msg.GetUser().GetSelect().(type) {
	case *pbUsers.Select_ById:
		row, err = s.db.Query(ctx, "SELECT user_id, key FROM core.uservaults WHERE user_id = "+
			"(SELECT id FROM core.users WHERE id = $1) AND key = $2",
			req.Msg.GetUser().GetById(), req.Msg.GetKey())
	case *pbUsers.Select_ByLogin:
		row, err = s.db.Query(ctx, "SELECT user_id, key FROM core.uservaults  WHERE user_id = "+
			"(SELECT id FROM core.users WHERE login = $1) AND key = $2",
			req.Msg.GetUser().GetByLogin(), req.Msg.GetKey())
	}

	return row, err
}

func (s *ServiceServer) Remove(ctx context.Context, req *connect.Request[pbUserVaults.RemoveRequest],
) (*connect.Response[pbUserVaults.RemoveResponse], error) {
	var (
		row pgx.Rows
		err error
		qb  *util.QueryBuilder
	)

	// Verify that User exists in core.users
	var userID string
	row, err = s.queryUser(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserVaults.RemoveResponse{
			Response: &pbUserVaults.RemoveResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"removing", _entityName, "key="+req.Msg.GetKey()),
				},
			},
		}), err
	}
	if !row.Next() {
		errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		err := fmt.Errorf(common.Errors[uint32(errno)],
			"removing "+_entityName, "user does not exist with that id/login")
		return connect.NewResponse(&pbUserVaults.RemoveResponse{
			Response: &pbUserVaults.RemoveResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	} else {
		if errScan := row.Scan(&userID); errScan != nil {
			err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)],
				"scanning", "core.users", "for id")
			return connect.NewResponse(&pbUserVaults.RemoveResponse{
				Response: &pbUserVaults.RemoveResponse_Error{
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

	qb, err = s.repo.QbRemove(req.Msg, userID)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserVaults.RemoveResponse{
			Response: &pbUserVaults.RemoveResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"removing", _entityName, "key="+req.Msg.GetKey()),
				},
			},
		}), err
	}
	row.Close()

	sqlStr, args, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if row, err = tx.Query(ctx, sqlStr, args...); err != nil {
			return err
		}
		defer row.Close()
		return nil
	})
	if errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"removing", _entityName, "key = "+req.Msg.GetKey())
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserVaults.RemoveResponse{
			Response: &pbUserVaults.RemoveResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + errTx.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserVaults.RemoveResponse{
		Response: &pbUserVaults.RemoveResponse_Ok{
			Ok: true,
		},
	}), nil
}

func (s *ServiceServer) Get(ctx context.Context, req *connect.Request[pbUserVaults.GetRequest],
) (*connect.Response[pbUserVaults.GetResponse], error) {
	var (
		row    pgx.Rows
		err    error
		userID string
	)

	row, err = s.queryUser(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserVaults.GetResponse{
			Response: &pbUserVaults.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"fetching", _entityName, "key="+req.Msg.GetKey()),
				},
			},
		}), err
	}
	if !row.Next() {
		errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		err := fmt.Errorf(common.Errors[uint32(errno)],
			_entityName, "key="+req.Msg.GetKey()+" and id/login="+req.Msg.GetUser().String())
		return connect.NewResponse(&pbUserVaults.GetResponse{
			Response: &pbUserVaults.GetResponse_Error{
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
			return connect.NewResponse(&pbUserVaults.GetResponse{
				Response: &pbUserVaults.GetResponse_Error{
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

	qb := s.repo.QbGetOne(req.Msg, userID)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL '" + sqlstr + "'")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserVaults.GetResponse{
			Response: &pbUserVaults.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	defer rows.Close()

	if !rows.Next() {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND)], _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserVaults.GetResponse{
			Response: &pbUserVaults.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	data, valueType, err := ScanRowValue(rows)
	if err != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)],
			"fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserVaults.GetResponse{
			Response: &pbUserVaults.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
					Package: _package,
					Text:    _err.Error() + " (" + err.Error() + ")",
				},
			},
		}), err
	}
	if rows.Next() {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND)],
			_entityNamePlural, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserVaults.GetResponse{
			Response: &pbUserVaults.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	decryptedBytes, err := decrypt(data, []byte(encryptKey))
	if err != nil {
		return connect.NewResponse(&pbUserVaults.GetResponse{
			Response: &pbUserVaults.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	value, err := bytesToValue(decryptedBytes, pbUserVaults.ValueType(valueType))
	if err != nil {
		return connect.NewResponse(&pbUserVaults.GetResponse{
			Response: &pbUserVaults.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}

	return connect.NewResponse(&pbUserVaults.GetResponse{
		Response: &pbUserVaults.GetResponse_Value{
			Value: value,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbUserVaults.GetListRequest],
	res *connect.ServerStream[pbUserVaults.GetListResponse],
) error {
	var (
		row    pgx.Rows
		err    error
		userID string
	)

	row, err = s.queryUser(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"listing", _entityNamePlural, "key="+req.Msg.GetKeyPrefix())
	}
	if !row.Next() {
		errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		err := fmt.Errorf(common.Errors[uint32(errno)],
			"listing "+_entityName, "user does not exist with that id/login")
		return err
	} else {
		if errScan := row.Scan(&userID); errScan != nil {
			err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)],
				"scanning", "core.users", "for id")
			return err
		}
	}
	row.Close()

	qb := s.repo.QbGetList(req.Msg, userID)
	sqlStr, args, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)
	}
	defer rows.Close()

	// Start building the response from here
	hasRows := false
	for rows.Next() {
		hasRows = true
		data, valueType, key, err := ScanRow(rows)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}

		decryptedBytes, err := decrypt(data, []byte(encryptKey))
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED)
		}
		keyValue, err := bytesToKeyValue(decryptedBytes, pbUserVaults.ValueType(valueType))
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED)
		}

		errSend := res.Send(&pbUserVaults.GetListResponse{
			Response: &pbUserVaults.GetListResponse_KeyValue{
				KeyValue: &pbUserVaults.KeyValue{
					Key:   key,
					Value: keyValue.Value,
				},
			},
		})
		if errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityNamePlural, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
			return _errSend
		}
	}

	if !hasRows { // If there are no match
		if !rows.Next() {
			_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, args)
			log.Error().Err(err).Msg(_err.Error())
			errSend := res.Send(&pbUserVaults.GetListResponse{
				Response: &pbUserVaults.GetListResponse_Error{
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
			return _err
		}
	}

	return rows.Err()
}

func (s *ServiceServer) Reset(
	ctx context.Context,
	req *connect.Request[pbUserVaults.ResetRequest],
	res *connect.ServerStream[pbUserVaults.ResetResponse],
) error {
	var userID string

	row, err := s.queryUser(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"reseting", _entityNamePlural, "key="+req.Msg.GetKeyPrefix())
	}
	if !row.Next() {
		errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
		err := fmt.Errorf(common.Errors[uint32(errno)],
			"reseting "+_entityName, "user does not exist with that id/login")
		return err
	} else {
		if errScan := row.Scan(&userID); errScan != nil {
			err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)],
				"scanning", "core.users", "for id")
			return err
		}
	}
	row.Close()

	qb := s.repo.QbReset(req.Msg, userID)
	qb.SetReturnFields("*")
	sqlStr, args, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return streamingErrReset(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)
	}
	defer rows.Close()

	// Start building the response from here
	hasRows := false
	for rows.Next() {
		hasRows = true
		_, _, err := ScanRowValue(rows)
		if err != nil {
			return streamingErrReset(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}
		errSend := res.Send(&pbUserVaults.ResetResponse{
			Response: &pbUserVaults.ResetResponse_Ok{
				Ok: true,
			},
		})
		if errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "reseting", _entityNamePlural, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
			return _errSend
		}
	}

	if !hasRows { // If there are no match
		if !rows.Next() {
			_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND)], _entityNamePlural, args)
			log.Error().Err(err).Msg(_err.Error())
			errSend := res.Send(&pbUserVaults.ResetResponse{
				Response: &pbUserVaults.ResetResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND,
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
			return _err
		}
	}

	return rows.Err()
}

func streamingErr(res *connect.ServerStream[pbUserVaults.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbUserVaults.GetListResponse{
		Response: &pbUserVaults.GetListResponse_Error{
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

func streamingErrReset(res *connect.ServerStream[pbUserVaults.ResetResponse], err error, errorCode pbCommon.ErrorCode) error {
	_err := fmt.Errorf(common.Errors[uint32(errorCode)], "reseting", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbUserVaults.ResetResponse{
		Response: &pbUserVaults.ResetResponse_Error{
			Error: &pbCommon.Error{
				Code: errorCode,
				Text: _err.Error() + " (" + err.Error() + ")",
			},
		},
	}); errSend != nil {
		_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
		_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "reseting", _entityName, "<Selection>")
		log.Error().Err(errSend).Msg(_errSend.Error())
		_err = _errSend
	}
	return _err
}
