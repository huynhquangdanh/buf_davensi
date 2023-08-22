package ibans

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbIbans "davensi.com/core/gen/ibans"
	pbIbansconnect "davensi.com/core/gen/ibans/ibansconnect"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/countries"
)

const (
	_package          = "ibans"
	_entityName       = "Iban"
	_entityNamePlural = "Ibans"
)

// ServiceServer implements the IbansService API
type ServiceServer struct {
	Repo IbanRepository
	pbIbansconnect.ServiceHandler
	db        *pgxpool.Pool
	countrySS *countries.ServiceServer
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:      *NewIbanRepository(db),
		db:        db,
		countrySS: countries.GetSingletonServiceServer(db),
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbIbans.CreateRequest],
) (*connect.Response[pbIbans.CreateResponse], error) {
	if _errno, validateErr := s.validateCreation(req.Msg); validateErr != nil {
		log.Error().Err(validateErr)
		return connect.NewResponse(&pbIbans.CreateResponse{
			Response: &pbIbans.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    validateErr.Error(),
				},
			},
		}), validateErr
	}

	qb, err := s.Repo.QbInsert(req.Msg)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(
			common.Errors[uint32(_errno.Number())],
			err.Error(),
		)
		log.Error().Err(_err)
		return connect.NewResponse(&pbIbans.CreateResponse{
			Response: &pbIbans.CreateResponse_Error{
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
	newIban, errExcute := common.ExecuteTxWrite[pbIbans.IBAN](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanMainEntity,
	)
	if errExcute != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
		)
		log.Error().Err(errExcute).Msg(_err.Error())
		return connect.NewResponse(&pbIbans.CreateResponse{
			Response: &pbIbans.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + errExcute.Error() + ")",
				},
			},
		}), errExcute
	}

	log.Info().Msgf(
		"%s created successfully with id = %s",
		_entityName, newIban.GetId(),
	)
	return connect.NewResponse(&pbIbans.CreateResponse{
		Response: &pbIbans.CreateResponse_Iban{
			Iban: newIban,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbIbans.UpdateRequest],
) (*connect.Response[pbIbans.UpdateResponse], error) {
	if errQueryUpdate := validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errQueryUpdate.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbIbans.UpdateResponse{
			Response: &pbIbans.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	getIbanResponse, err := s.getOldIbanToUpdate(req.Msg)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbIbans.UpdateResponse{
			Response: &pbIbans.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getIbanResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getIbanResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	ibanBeforeUpdate := getIbanResponse.Msg.GetIban()

	if _errno, errUpdateValue := s.validateMsgUpdate(ibanBeforeUpdate, req.Msg); errUpdateValue != nil {
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errUpdateValue.Error())
		log.Error().Err(_err)
		log.Error().Err(errUpdateValue)
		return connect.NewResponse(&pbIbans.UpdateResponse{
			Response: &pbIbans.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    errUpdateValue.Error(),
				},
			},
		}), errUpdateValue
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbIbans.UpdateResponse{
			Response: &pbIbans.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	sqlStr, args, sel := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	updatedIban, errExcute := common.ExecuteTxWrite[pbIbans.IBAN](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanMainEntity,
	)

	if errExcute != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			sel,
		)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbIbans.UpdateResponse{
			Response: &pbIbans.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbIbans.UpdateResponse{
		Response: &pbIbans.UpdateResponse_Iban{
			Iban: updatedIban,
		},
	}), nil
}

func (s *ServiceServer) getOldIbanToUpdate(msg *pbIbans.UpdateRequest) (*connect.Response[pbIbans.GetResponse], error) {
	getIbanRequest := &pbIbans.GetRequest{}

	switch msg.Select.(type) {
	case *pbIbans.UpdateRequest_ById:
		getIbanRequest.Select = &pbIbans.Select{
			Select: &pbIbans.Select_ById{
				ById: msg.GetById(),
			},
		}
	case *pbIbans.UpdateRequest_ByCountryValidity:
		getIbanRequest.Select = &pbIbans.Select{
			Select: &pbIbans.Select_ByCountryValidity{
				ByCountryValidity: msg.GetByCountryValidity(),
			},
		}
	}

	getIbanRes, err := s.Get(context.Background(), &connect.Request[pbIbans.GetRequest]{
		Msg: getIbanRequest,
	})
	if err != nil {
		return nil, err
	}

	return getIbanRes, nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbIbans.GetRequest],
) (*connect.Response[pbIbans.GetResponse], error) {
	if errQueryGet := ValidateSelect(req.Msg.Select, "fetching"); errQueryGet != nil {
		log.Error().Err(errQueryGet.Err)
		return connect.NewResponse(&pbIbans.GetResponse{
			Response: &pbIbans.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryGet.Code,
					Package: _package,
					Text:    errQueryGet.Err.Error(),
				},
			},
		}), errQueryGet.Err
	}

	qbCountry := s.countrySS.Repo.QbGetList(&pbCountries.GetListRequest{})
	countryFB, countryArgs := qbCountry.Filters.GenerateSQL()

	qb := s.Repo.QbGetOne(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON countries.id = ibans.country_id", qbCountry.TableName)).
		Select(strings.Join(qbCountry.SelectFields, ", ")).
		Where(countryFB, countryArgs...)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbIbans.GetResponse{
			Response: &pbIbans.GetResponse_Error{
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
		return connect.NewResponse(&pbIbans.GetResponse{
			Response: &pbIbans.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	iban, err := s.Repo.ScanWithRelationship(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbIbans.GetResponse{
			Response: &pbIbans.GetResponse_Error{
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
		return connect.NewResponse(&pbIbans.GetResponse{
			Response: &pbIbans.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbIbans.GetResponse{
		Response: &pbIbans.GetResponse_Iban{
			Iban: iban,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbIbans.GetListRequest],
	res *connect.ServerStream[pbIbans.GetListResponse],
) error {
	if req.Msg.Country == nil {
		req.Msg.Country = &pbCountries.GetListRequest{}
	}
	qbCountry := s.countrySS.Repo.QbGetList(req.Msg.Country)
	countryFB, countryArgs := qbCountry.Filters.GenerateSQL()

	qb := s.Repo.QbGetList(req.Msg).
		Join(fmt.Sprintf("LEFT JOIN %s ON countries.id = ibans.country_id", qbCountry.TableName)).
		Select(strings.Join(qbCountry.SelectFields, ", ")).
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
				return res.Send(&pbIbans.GetListResponse{
					Response: &pbIbans.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		iban, err := s.Repo.ScanWithRelationship(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbIbans.GetListResponse{
						Response: &pbIbans.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbIbans.GetListResponse{
			Response: &pbIbans.GetListResponse_Iban{
				Iban: iban,
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
	req *connect.Request[pbIbans.DeleteRequest],
) (*connect.Response[pbIbans.DeleteResponse], error) {
	updateIbanRequest := &pbIbans.UpdateRequest{
		Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
	}

	switch req.Msg.Select.(type) {
	case *pbIbans.DeleteRequest_ById:
		updateIbanRequest.Select = &pbIbans.UpdateRequest_ById{
			ById: req.Msg.GetById(),
		}
	case *pbIbans.DeleteRequest_ByCountryValidity:
		updateIbanRequest.Select = &pbIbans.UpdateRequest_ByCountryValidity{
			ByCountryValidity: req.Msg.GetByCountryValidity(),
		}
	}
	terminatedStatus := pbCommon.Status_STATUS_TERMINATED
	iban := &pbIbans.UpdateRequest{
		Select: &pbIbans.UpdateRequest_ById{
			ById: req.Msg.GetById(),
		},
		Status: &terminatedStatus,
	}
	deleteReq := connect.NewRequest(iban)
	deletedIban, err := s.Update(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbIbans.DeleteResponse{
			Response: &pbIbans.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedIban.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedIban.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbIbans.DeleteResponse{
		Response: &pbIbans.DeleteResponse_Iban{
			Iban: deletedIban.Msg.GetIban(),
		},
	}), nil
}
