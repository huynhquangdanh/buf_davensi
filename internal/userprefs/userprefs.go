package userprefs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbUserPrefs "davensi.com/core/gen/userprefs"
	pbUserPrefsconnect "davensi.com/core/gen/userprefs/userprefsconnect"
	pbUsers "davensi.com/core/gen/users"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/util"
)

const (
	_package          = "userprefs"
	_entityName       = "User Pref"
	_entityNamePlural = "User Prefs"
	_fields           = "user_id, key, value, status"
	_table            = "core.userprefs"
)

// ServiceServer implements the AddressesService API
type ServiceServer struct {
	repo UserPrefsRepository
	pbUserPrefsconnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo: *NewUserPrefRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) Set(ctx context.Context, req *connect.Request[pbUserPrefs.SetRequest],
) (*connect.Response[pbUserPrefs.SetResponse], error) {
	// TODO: get default value if value is not specified. Currently don't know what to get
	// TODO: display full user.User field in response. Currently only user.ID

	// Verify that user.User, Key are specified
	_errno, validateErr := s.ValidateCreate(req.Msg)
	if validateErr != nil {
		log.Error().Err(validateErr)
		return connect.NewResponse(&pbUserPrefs.SetResponse{
			Response: &pbUserPrefs.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    validateErr.Error(),
				},
			},
		}), validateErr
	}

	// Verify that Key does not exist in core.userprefs_default (exists OR not???)
	var (
		row pgx.Rows
		err error
	)
	row, err = s.db.Query(ctx, "SELECT key FROM core.userprefs_default WHERE key = $1", req.Msg.GetKey())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserPrefs.SetResponse{
			Response: &pbUserPrefs.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text: fmt.Sprintf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
						"setting", _entityName, "key="+req.Msg.GetKey()),
				},
			},
		}), err
	}
	if row.Next() {
		err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT)],
			"setting "+_entityName, "key already exists in default table")
		return connect.NewResponse(&pbUserPrefs.SetResponse{
			Response: &pbUserPrefs.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	row.Close()

	// Verify that User exists in core.users
	var userID string
	row, err = s.queryUser(ctx, req)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserPrefs.SetResponse{
			Response: &pbUserPrefs.SetResponse_Error{
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
		return connect.NewResponse(&pbUserPrefs.SetResponse{
			Response: &pbUserPrefs.SetResponse_Error{
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
			return connect.NewResponse(&pbUserPrefs.SetResponse{
				Response: &pbUserPrefs.SetResponse_Error{
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
	row, err = s.queryUserPref(ctx, req)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserPrefs.SetResponse{
			Response: &pbUserPrefs.SetResponse_Error{
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
		qb, err = s.repo.QbSetInsert(req.Msg, userID)
	} else {
		qb, err = s.repo.QbSetUpdate(req.Msg, userID)
	}
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserPrefs.SetResponse{
			Response: &pbUserPrefs.SetResponse_Error{
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
	var (
		setUserPref *pbUserPrefs.UserPref
	)
	log.Info().Msg("Executing SQL '" + sqlStr + "'")

	setUserPref, errTx := common.ExecuteTxWrite[pbUserPrefs.UserPref](ctx, s.db, sqlStr, args, ScanRow)
	if errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"setting", _entityName, "key = "+req.Msg.GetKey())
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserPrefs.SetResponse{
			Response: &pbUserPrefs.SetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + errTx.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserPrefs.SetResponse{
		Response: &pbUserPrefs.SetResponse_Userpref{
			Userpref: setUserPref,
		},
	}), nil
}

// Check if user exists for Set()
func (s *ServiceServer) queryUser(ctx context.Context, req *connect.Request[pbUserPrefs.SetRequest],
) (pgx.Rows, error) {
	var (
		row pgx.Rows
		err error
	)

	switch req.Msg.GetUser().GetSelect().(type) {
	case *pbUsers.Select_ById:
		row, err = s.db.Query(ctx, "SELECT id FROM core.users WHERE id = $1", req.Msg.GetUser().GetById())
	case *pbUsers.Select_ByLogin:
		row, err = s.db.Query(ctx, "SELECT id FROM core.users WHERE login = $1", req.Msg.GetUser().GetByLogin())
	}

	return row, err
}

// Check if User Pref exists for Set() => INSERT or UPDATE
func (s *ServiceServer) queryUserPref(ctx context.Context, req *connect.Request[pbUserPrefs.SetRequest],
) (pgx.Rows, error) {
	var (
		row pgx.Rows
		err error
	)

	switch req.Msg.GetUser().GetSelect().(type) {
	case *pbUsers.Select_ById:
		row, err = s.db.Query(ctx, "SELECT user_id, key FROM core.userprefs WHERE user_id = "+
			"(SELECT id FROM core.users WHERE id = $1) AND key = $2",
			req.Msg.GetUser().GetById(), req.Msg.GetKey())
	case *pbUsers.Select_ByLogin:
		row, err = s.db.Query(ctx, "SELECT user_id, key FROM core.userprefs WHERE user_id = "+
			"(SELECT id FROM core.users WHERE login = $1) AND key = $2",
			req.Msg.GetUser().GetByLogin(), req.Msg.GetKey())
	}

	return row, err
}

func (s *ServiceServer) Remove(ctx context.Context, req *connect.Request[pbUserPrefs.RemoveRequest],
) (*connect.Response[pbUserPrefs.RemoveResponse], error) {
	var (
		row pgx.Rows
		err error
		qb  *util.QueryBuilder
	)

	// Verify that User exists in core.users
	var userID string
	row, err = s.queryUserID(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserPrefs.RemoveResponse{
			Response: &pbUserPrefs.RemoveResponse_Error{
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
		return connect.NewResponse(&pbUserPrefs.RemoveResponse{
			Response: &pbUserPrefs.RemoveResponse_Error{
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
			return connect.NewResponse(&pbUserPrefs.RemoveResponse{
				Response: &pbUserPrefs.RemoveResponse_Error{
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
		return connect.NewResponse(&pbUserPrefs.RemoveResponse{
			Response: &pbUserPrefs.RemoveResponse_Error{
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

	sqlStr, args, _ := qb.SetReturnFields("*").GenerateSQL()

	var (
		removedUserPref *pbUserPrefs.UserPref
		errScan         error
	)
	log.Info().Msg("Executing SQL '" + sqlStr + "'")
	errTx := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if row, err = tx.Query(ctx, sqlStr, args...); err != nil {
			return err
		}
		defer row.Close()

		if row.Next() {
			if removedUserPref, errScan = ScanRow(row); errScan != nil {
				log.Error().Err(err).Msgf("unable to set %s with key = '%s'", _entityName, req.Msg.GetKey())
				return errScan
			}
			log.Info().Msgf("%s with key = '%s' removed successfully",
				_entityName, req.Msg.GetKey())
			return nil
		} else {
			return fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND)],
				_entityName, "id/login="+req.Msg.GetUser().String())
		}
	})
	if errTx != nil {
		_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"removing", _entityName, "key = "+req.Msg.GetKey())
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserPrefs.RemoveResponse{
			Response: &pbUserPrefs.RemoveResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    _err.Error() + "(" + errTx.Error() + ")",
				},
			},
		}), errTx
	}

	return connect.NewResponse(&pbUserPrefs.RemoveResponse{
		Response: &pbUserPrefs.RemoveResponse_Userpref{
			Userpref: removedUserPref,
		},
	}), nil
}

func (s *ServiceServer) queryUserID(ctx context.Context, req *pbUsers.Select,
) (pgx.Rows, error) {
	var row pgx.Rows
	var err error

	switch req.GetSelect().(type) {
	case *pbUsers.Select_ByLogin:
		row, err = s.db.Query(ctx, "SELECT id FROM core.users WHERE login = $1", req.GetByLogin())

	case *pbUsers.Select_ById:
		row, err = s.db.Query(ctx, "SELECT id FROM core.users WHERE id = $1", req.GetById())
	}

	return row, err
}

func (s *ServiceServer) Get(ctx context.Context, req *connect.Request[pbUserPrefs.GetRequest],
) (*connect.Response[pbUserPrefs.GetResponse], error) {
	// TODO: return default value based on key from userprefs_default table

	var (
		row    pgx.Rows
		err    error
		userID string
	)

	row, err = s.queryUserID(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbUserPrefs.GetResponse{
			Response: &pbUserPrefs.GetResponse_Error{
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
		return connect.NewResponse(&pbUserPrefs.GetResponse{
			Response: &pbUserPrefs.GetResponse_Error{
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
			return connect.NewResponse(&pbUserPrefs.GetResponse{
				Response: &pbUserPrefs.GetResponse_Error{
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
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserPrefs.GetResponse{
			Response: &pbUserPrefs.GetResponse_Error{
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
		return connect.NewResponse(&pbUserPrefs.GetResponse{
			Response: &pbUserPrefs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	userPref, err := s.repo.ScanRowWithRelationship(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbUserPrefs.GetResponse{
			Response: &pbUserPrefs.GetResponse_Error{
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
		return connect.NewResponse(&pbUserPrefs.GetResponse{
			Response: &pbUserPrefs.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbUserPrefs.GetResponse{
		Response: &pbUserPrefs.GetResponse_Userpref{
			Userpref: userPref,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbUserPrefs.GetListRequest],
	res *connect.ServerStream[pbUserPrefs.GetListResponse],
) error {
	// TODO: wait for proto to be fixed, returning individual userPref instead of a list
	// TODO: return default value from userprefs_default table if keyprefix is empty

	var (
		row    pgx.Rows
		err    error
		userID string
	)

	row, err = s.queryUserID(ctx, req.Msg.GetUser())
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
	userPrefList := pbUserPrefs.UserPrefList{}
	hasRows := false
	for rows.Next() {
		hasRows = true
		userPref, err := s.repo.ScanRowWithRelationship(rows)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}
		userPrefList.List = append(userPrefList.List, userPref)

		// errSend := res.Send(&pbUserPrefs.GetListResponse{	GONNA IMPLEMENT THIS LATER AFTER PROTO IS FIXED
		// 	Response: &pbUserPrefs.GetListResponse_Userprefs{
		// 		Userprefs: userPref,
		// 	},
		// })
		// if errSend != nil {
		// 	_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
		// 	_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityNamePlural, "<Selection>")
		// 	log.Error().Err(errSend).Msg(_errSend.Error())
		// 	return _errSend
		// }
	}

	errSend := res.Send(&pbUserPrefs.GetListResponse{
		Response: &pbUserPrefs.GetListResponse_Userprefs{
			Userprefs: &userPrefList,
		},
	})
	if errSend != nil {
		_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
		_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityNamePlural, "<Selection>")
		log.Error().Err(errSend).Msg(_errSend.Error())
		return _errSend
	}

	if !hasRows { // If there are no match
		if !rows.Next() {
			_errno := pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND
			_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], _entityNamePlural, args)
			log.Error().Err(err).Msg(_err.Error())
			errSend := res.Send(&pbUserPrefs.GetListResponse{
				Response: &pbUserPrefs.GetListResponse_Error{
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
	req *connect.Request[pbUserPrefs.ResetRequest],
	res *connect.ServerStream[pbUserPrefs.ResetResponse],
) error {
	// TODO: wait for proto to be fixed, returning individual userPref instead of a list

	var userID string

	row, err := s.queryUserID(ctx, req.Msg.GetUser())
	if err != nil {
		log.Error().Err(err)
		return fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)],
			"reseting", _entityNamePlural, "key="+req.Msg.GetKeyPrefix())
	}
	defer row.Close()

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

	qb := s.repo.QbReset(req.Msg, userID)
	sqlStr, args, _ := qb.GenerateSQL()
	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return streamingErrReset(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_ERROR)
	}
	defer rows.Close()

	// Start building the response from here
	resetUserPrefList := pbUserPrefs.UserPrefList{}
	hasRows := false
	for rows.Next() {
		hasRows = true
		resetUserPref, err := s.repo.ScanRowWithRelationship(rows)
		if err != nil {
			return streamingErrReset(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}
		resetUserPrefList.List = append(resetUserPrefList.List, resetUserPref)
	}

	errSend := res.Send(&pbUserPrefs.ResetResponse{
		Response: &pbUserPrefs.ResetResponse_Userprefs{
			Userprefs: &resetUserPrefList,
		},
	})
	if errSend != nil {
		_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
		_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "reseting", _entityNamePlural, "<Selection>")
		log.Error().Err(errSend).Msg(_errSend.Error())
		return _errSend
	}

	if !hasRows { // If there are no match
		if !rows.Next() {
			_err := fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND)], _entityNamePlural, args)
			log.Error().Err(err).Msg(_err.Error())
			errSend := res.Send(&pbUserPrefs.ResetResponse{
				Response: &pbUserPrefs.ResetResponse_Error{
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

func streamingErr(res *connect.ServerStream[pbUserPrefs.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbUserPrefs.GetListResponse{
		Response: &pbUserPrefs.GetListResponse_Error{
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

func streamingErrReset(res *connect.ServerStream[pbUserPrefs.ResetResponse], err error, errorCode pbCommon.ErrorCode) error {
	_err := fmt.Errorf(common.Errors[uint32(errorCode)], "reseting", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbUserPrefs.ResetResponse{
		Response: &pbUserPrefs.ResetResponse_Error{
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
