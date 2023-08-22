package socials

import (
	"context"
	"fmt"
	"sync"

	"davensi.com/core/internal/util"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbKyc "davensi.com/core/gen/kyc"
	pbSocials "davensi.com/core/gen/socials"
	pbSocialsConnect "davensi.com/core/gen/socials/socialsconnect"
	"davensi.com/core/internal/common"
)

const (
	_package          = "socials"
	_entityName       = "KYC Social"
	_entityNamePlural = "KYC Socials"
)

// ServiceServer implements the SocialsService API
type ServiceServer struct {
	repo SocialRepository
	pbSocialsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

// For singleton Credentials export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		repo: *NewSocialRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbSocials.CreateRequest],
) (*connect.Response[pbSocials.CreateResponse], error) {
	id := uuid.New().String()
	qb, err := s.repo.QbInsert(id, req.Msg)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbSocials.CreateResponse{
			Response: &pbSocials.CreateResponse_Error{
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
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sqlStr, args...)
		return err
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbSocials.CreateResponse{
			Response: &pbSocials.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	// Start building the response from here
	newSocial, err := s.Get(ctx, &connect.Request[pbSocials.GetRequest]{
		Msg: &pbSocials.GetRequest{
			Id: id,
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("unable to create %s",
			_entityName,
		)
		return connect.NewResponse(&pbSocials.CreateResponse{
			Response: &pbSocials.CreateResponse_Error{
				Error: newSocial.Msg.GetError(),
			},
		}), err
	}
	log.Info().Msgf("%s created successfully with id = %s",
		_entityName, newSocial.Msg.GetSocial().GetId())
	return connect.NewResponse(&pbSocials.CreateResponse{
		Response: &pbSocials.CreateResponse_Social{
			Social: newSocial.Msg.GetSocial(),
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbSocials.UpdateRequest],
) (*connect.Response[pbSocials.UpdateResponse], error) {
	if errQueryUpdate := validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errQueryUpdate.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbSocials.UpdateResponse{
			Response: &pbSocials.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	getSocialResponse, err := s.getOldSocialToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbSocials.UpdateResponse{
			Response: &pbSocials.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getSocialResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getSocialResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	socialBeforeUpdate := getSocialResponse.Msg.GetSocial()

	qb, genSQLError := s.repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbSocials.UpdateResponse{
			Response: &pbSocials.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlstr, sqlArgs, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")
	if err := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sqlstr, sqlArgs...)
		return err
	}); err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			sel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbSocials.UpdateResponse{
			Response: &pbSocials.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	// Start building the response from here
	updatedSocial, err := s.Get(ctx, &connect.Request[pbSocials.GetRequest]{
		Msg: &pbSocials.GetRequest{
			Id: socialBeforeUpdate.Id,
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("unable to update %s with %s", _entityName, sel)
		return connect.NewResponse(&pbSocials.UpdateResponse{
			Response: &pbSocials.UpdateResponse_Error{
				Error: updatedSocial.Msg.GetError(),
			},
		}), err
	}
	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbSocials.UpdateResponse{
		Response: &pbSocials.UpdateResponse_Social{
			Social: updatedSocial.Msg.GetSocial(),
		},
	}), nil
}

func (s *ServiceServer) getOldSocialToUpdate(msg *pbSocials.UpdateRequest) (*connect.Response[pbSocials.GetResponse], error) {
	getUomRequest := &pbSocials.GetRequest{Id: msg.GetId()}

	getUomRes, err := s.Get(context.Background(), &connect.Request[pbSocials.GetRequest]{
		Msg: getUomRequest,
	})
	if err != nil {
		return nil, err
	}

	return getUomRes, nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbSocials.GetRequest],
) (*connect.Response[pbSocials.GetResponse], error) {
	qb := s.repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbSocials.GetResponse{
			Response: &pbSocials.GetResponse_Error{
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
		return connect.NewResponse(&pbSocials.GetResponse{
			Response: &pbSocials.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	social, err := ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbSocials.GetResponse{
			Response: &pbSocials.GetResponse_Error{
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
		return connect.NewResponse(&pbSocials.GetResponse{
			Response: &pbSocials.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbSocials.GetResponse{
		Response: &pbSocials.GetResponse_Social{
			Social: social,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbSocials.GetListRequest],
	res *connect.ServerStream[pbSocials.GetListResponse],
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
		social, err := ScanRow(rows)
		if err != nil {
			return streamingErr(res, err, pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR)
		}

		if errSend := res.Send(&pbSocials.GetListResponse{
			Response: &pbSocials.GetListResponse_Social{
				Social: social,
			},
		}); errSend != nil {
			_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
			_errSend := fmt.Errorf(common.Errors[uint32(_errnoSend.Number())], "listing", _entityName, "<Selection>")
			log.Error().Err(errSend).Msg(_errSend.Error())
		}
	}

	return rows.Err()
}

func streamingErr(res *connect.ServerStream[pbSocials.GetListResponse], err error, errorCode pbCommon.ErrorCode) error {
	_errno := errorCode
	_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "listing", _entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := res.Send(&pbSocials.GetListResponse{
		Response: &pbSocials.GetListResponse_Error{
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

func (s *ServiceServer) Delete(
	ctx context.Context,
	req *connect.Request[pbSocials.DeleteRequest],
) (*connect.Response[pbSocials.DeleteResponse], error) {
	canceledStatus := pbKyc.Status_STATUS_CANCELED
	contact := &pbSocials.UpdateRequest{
		Id:     req.Msg.GetId(),
		Status: &canceledStatus,
	}
	deleteReq := connect.NewRequest(contact)
	deletedSocial, err := s.Update(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbSocials.DeleteResponse{
			Response: &pbSocials.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedSocial.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedSocial.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbSocials.DeleteResponse{
		Response: &pbSocials.DeleteResponse_Social{
			Social: deletedSocial.Msg.GetSocial(),
		},
	}), nil
}

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

func (s *ServiceServer) MakeCreationQB(msg *pbSocials.CreateRequest) (*util.QueryBuilder, *common.ErrWithCode) {
	id := uuid.New().String()
	qb, err := s.repo.QbInsert(id, msg)
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

func (s *ServiceServer) MakeUpdateQB(msg *pbSocials.UpdateRequest, upsert bool) (*util.QueryBuilder, *common.ErrWithCode) {
	oldSocial, getOldSocialErr := s.getOldSocialToUpdate(msg)
	if getOldSocialErr != nil {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating",
			_entityName,
			getOldSocialErr.Error(),
		)
	}
	var (
		qb    *util.QueryBuilder
		qbErr error
	)
	if upsert && oldSocial == nil {
		qb, qbErr = s.repo.QbInsert(uuid.New().String(), &pbSocials.CreateRequest{
			RelationshipStatus: msg.RelationshipStatus,
			Religion:           msg.Religion,
			SocialClass:        msg.SocialClass,
			Profession:         msg.Profession,
			Status:             msg.GetStatus().Enum(),
		})
	} else {
		qb, qbErr = s.repo.QbUpdate(msg)
	}
	if qbErr != nil {
		return nil, common.CreateErrWithCode(
			pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
			"updating",
			_entityName,
			qbErr.Error(),
		)
	}
	return qb, nil
}
