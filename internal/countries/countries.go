package countries

import (
	"context"
	"fmt"
	"sync"

	"github.com/bufbuild/connect-go"
	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbCountries "davensi.com/core/gen/countries"
	pbCountriesconnect "davensi.com/core/gen/countries/countriesconnect"
	"davensi.com/core/gen/uoms"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/cryptos"
	"davensi.com/core/internal/fiats"
)

const (
	_package          = "countries"
	_entityName       = "Country"
	_entityNamePlural = "Countries"
)

// ServiceServer implements the CountriesService API
type ServiceServer struct {
	Repo CountryRepository
	pbCountriesconnect.ServiceHandler
	db        *pgxpool.Pool
	cryptosSS *cryptos.ServiceServer
	fiatsSS   *fiats.ServiceServer
}

// For singleton UoMs export module
var (
	singletonServiceServer *ServiceServer
	once                   sync.Once
)

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo:      *NewCountryRepository(db),
		db:        db,
		cryptosSS: cryptos.GetSingletonServiceServer(db),
		fiatsSS:   fiats.GetSingletonServiceServer(db),
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
	req *connect.Request[pbCountries.CreateRequest],
) (*connect.Response[pbCountries.CreateResponse], error) {
	if validateErr := s.ValidateCreate(req.Msg); validateErr != nil {
		log.Error().Err(validateErr.Err)
		return connect.NewResponse(&pbCountries.CreateResponse{
			Response: &pbCountries.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    validateErr.Code,
					Package: _package,
					Text:    validateErr.Err.Error(),
				},
			},
		}), validateErr.Err
	}
	errCreating := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	qb, err := s.Repo.QbInsert(req.Msg)
	if err != nil {
		errCreating.UpdateMessage(err.Error())
		log.Error().Err(errCreating.Err)
		return connect.NewResponse(&pbCountries.CreateResponse{
			Response: &pbCountries.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errCreating.Code,
					Package: _package,
					Text:    errCreating.Err.Error(),
				},
			},
		}), errCreating.Err
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
		return connect.NewResponse(&pbCountries.CreateResponse{
			Response: &pbCountries.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	// Start building the response from here
	newCountry, err := s.Get(ctx, &connect.Request[pbCountries.GetRequest]{
		Msg: &pbCountries.GetRequest{
			Select: &pbCountries.Select{
				Select: &pbCountries.Select_ByCode{
					ByCode: req.Msg.GetCode(),
				},
			},
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("unable to create %s",
			_entityName,
		)
		return connect.NewResponse(&pbCountries.CreateResponse{
			Response: &pbCountries.CreateResponse_Error{
				Error: newCountry.Msg.GetError(),
			},
		}), err
	}
	log.Info().Msgf("%s created successfully with id = %s",
		_entityName, newCountry.Msg.GetCountry().GetId())
	return connect.NewResponse(&pbCountries.CreateResponse{
		Response: &pbCountries.CreateResponse_Country{
			Country: newCountry.Msg.GetCountry(),
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbCountries.UpdateRequest],
) (*connect.Response[pbCountries.UpdateResponse], error) {
	if errQueryUpdate := ValidateSelect(req.Msg.Select, "updating"); errQueryUpdate != nil {
		log.Error().Err(errQueryUpdate.Err)
		return connect.NewResponse(&pbCountries.UpdateResponse{
			Response: &pbCountries.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    errQueryUpdate.Code,
					Package: _package,
					Text:    errQueryUpdate.Err.Error(),
				},
			},
		}), errQueryUpdate.Err
	}

	getCountryResponse, err := s.Get(ctx, connect.NewRequest(&pbCountries.GetRequest{
		Select: req.Msg.Select,
	}))
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbCountries.UpdateResponse{
			Response: &pbCountries.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    getCountryResponse.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + getCountryResponse.Msg.GetError().Text,
				},
			},
		}), err
	}

	countrieBeforeUpdate := getCountryResponse.Msg.GetCountry()

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbCountries.UpdateResponse{
			Response: &pbCountries.UpdateResponse_Error{
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
		return connect.NewResponse(&pbCountries.UpdateResponse{
			Response: &pbCountries.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error() + "(" + err.Error() + ")",
				},
			},
		}), err
	}

	// Start building the response from here
	updatedCountry, err := s.Get(ctx, &connect.Request[pbCountries.GetRequest]{
		Msg: &pbCountries.GetRequest{
			Select: &pbCountries.Select{
				Select: &pbCountries.Select_ById{
					ById: countrieBeforeUpdate.Id,
				},
			},
		},
	})
	if err != nil {
		log.Error().Err(err).Msgf("unable to update %s with %s", _entityName, sel)
		return connect.NewResponse(&pbCountries.UpdateResponse{
			Response: &pbCountries.UpdateResponse_Error{
				Error: updatedCountry.Msg.GetError(),
			},
		}), err
	}
	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbCountries.UpdateResponse{
		Response: &pbCountries.UpdateResponse_Country{
			Country: updatedCountry.Msg.GetCountry(),
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbCountries.GetRequest],
) (*connect.Response[pbCountries.GetResponse], error) {
	qb := s.Repo.QbGetOne(req.Msg)
	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCountries.GetResponse{
			Response: &pbCountries.GetResponse_Error{
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
		return connect.NewResponse(&pbCountries.GetResponse{
			Response: &pbCountries.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	countrie, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbCountries.GetResponse{
			Response: &pbCountries.GetResponse_Error{
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
		return connect.NewResponse(&pbCountries.GetResponse{
			Response: &pbCountries.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbCountries.GetResponse{
		Response: &pbCountries.GetResponse_Country{
			Country: countrie,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbCountries.GetListRequest],
	res *connect.ServerStream[pbCountries.GetListResponse],
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
				return res.Send(&pbCountries.GetListResponse{
					Response: &pbCountries.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		countrie, err := s.Repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbCountries.GetListResponse{
						Response: &pbCountries.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbCountries.GetListResponse{
			Response: &pbCountries.GetListResponse_Country{
				Country: countrie,
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
	req *connect.Request[pbCountries.DeleteRequest],
) (*connect.Response[pbCountries.DeleteResponse], error) {
	deletedCountry, err := s.Update(ctx, connect.NewRequest(&pbCountries.UpdateRequest{
		Select: req.Msg.Select,
		Status: pbCommon.Status_STATUS_TERMINATED.Enum(),
	}))
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbCountries.DeleteResponse{
			Response: &pbCountries.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedCountry.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedCountry.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbCountries.DeleteResponse{
		Response: &pbCountries.DeleteResponse_Country{
			Country: deletedCountry.Msg.GetCountry(),
		},
	}), nil
}

func (s *ServiceServer) SetFiats(
	ctx context.Context,
	req *connect.Request[pbCountries.SetFiatsRequest],
) (*connect.Response[pbCountries.SetFiatsResponse], error) {
	countryRes, errCountryRes := s.Get(ctx, connect.NewRequest[pbCountries.GetRequest](&pbCountries.GetRequest{
		Select: req.Msg.Select,
	}))

	if errCountryRes != nil {
		return connect.NewResponse[pbCountries.SetFiatsResponse](&pbCountries.SetFiatsResponse{
			Response: &pbCountries.SetFiatsResponse_Error{
				Error: countryRes.Msg.GetError(),
			},
		}), errCountryRes
	}

	handleFn, err := s.GenSetFiatsHandleFn(countryRes.Msg.GetCountry(), req.Msg.GetFiats())

	if err != nil {
		return connect.NewResponse[pbCountries.SetFiatsResponse](&pbCountries.SetFiatsResponse{
			Response: &pbCountries.SetFiatsResponse_Error{
				Error: &pbCommon.Error{
					Code:    err.Code,
					Package: _package,
					Text:    err.Err.Error(),
				},
			},
		}), err.Err
	}

	if excuteErr := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := handleFn(tx)

		return err
	}); excuteErr != nil {
		return &connect.Response[pbCountries.SetFiatsResponse]{
			Msg: &pbCountries.SetFiatsResponse{
				Response: &pbCountries.SetFiatsResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    excuteErr.Error(),
					},
				},
			},
		}, excuteErr
	}

	fiatsRes, errFiatsRes := s.GetFiats(ctx, connect.NewRequest[pbCountries.GetFiatsRequest](&pbCountries.GetFiatsRequest{
		Select: req.Msg.Select,
	}))

	if errFiatsRes != nil {
		return &connect.Response[pbCountries.SetFiatsResponse]{
			Msg: &pbCountries.SetFiatsResponse{
				Response: &pbCountries.SetFiatsResponse_Error{
					Error: fiatsRes.Msg.GetError(),
				},
			},
		}, errFiatsRes
	}

	return &connect.Response[pbCountries.SetFiatsResponse]{
		Msg: &pbCountries.SetFiatsResponse{
			Response: &pbCountries.SetFiatsResponse_Fiats{
				Fiats: fiatsRes.Msg.GetFiats(),
			},
		},
	}, nil
}

func (s *ServiceServer) AddFiats(
	ctx context.Context,
	req *connect.Request[pbCountries.AddFiatsRequest],
) (*connect.Response[pbCountries.AddFiatsResponse], error) {
	countryRes, errCountryRes := s.Get(ctx, connect.NewRequest[pbCountries.GetRequest](&pbCountries.GetRequest{
		Select: req.Msg.Select,
	}))

	if errCountryRes != nil {
		return &connect.Response[pbCountries.AddFiatsResponse]{
			Msg: &pbCountries.AddFiatsResponse{
				Response: &pbCountries.AddFiatsResponse_Error{
					Error: countryRes.Msg.GetError(),
				},
			},
		}, errCountryRes
	}

	country := countryRes.Msg.GetCountry()

	handleFn, err := s.GenAddFiatsHandleFn(country, req.Msg.GetFiats())

	if err != nil {
		return connect.NewResponse(&pbCountries.AddFiatsResponse{
			Response: &pbCountries.AddFiatsResponse_Error{
				Error: &pbCommon.Error{
					Code:    err.Code,
					Package: _package,
					Text:    err.Err.Error(),
				},
			},
		}), err.Err
	}

	if excuteErr := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := handleFn(tx)

		return err
	}); excuteErr != nil {
		errRes := &pbCommon.Error{
			Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			Package: _package,
			Text:    excuteErr.Error(),
		}
		return &connect.Response[pbCountries.AddFiatsResponse]{
			Msg: &pbCountries.AddFiatsResponse{
				Response: &pbCountries.AddFiatsResponse_Error{
					Error: errRes,
				},
			},
		}, excuteErr
	}

	fiatsRes, errFiatsRes := s.GetFiats(ctx, connect.NewRequest[pbCountries.GetFiatsRequest](&pbCountries.GetFiatsRequest{
		Select: req.Msg.Select,
	}))

	if errFiatsRes != nil {
		return &connect.Response[pbCountries.AddFiatsResponse]{
			Msg: &pbCountries.AddFiatsResponse{
				Response: &pbCountries.AddFiatsResponse_Error{
					Error: fiatsRes.Msg.GetError(),
				},
			},
		}, errFiatsRes
	}

	fiatsList := fiatsRes.Msg.GetFiats()

	return &connect.Response[pbCountries.AddFiatsResponse]{
		Msg: &pbCountries.AddFiatsResponse{
			Response: &pbCountries.AddFiatsResponse_Fiats{
				Fiats: fiatsList,
			},
		},
	}, nil
}

func (s *ServiceServer) GetFiats(
	ctx context.Context,
	req *connect.Request[pbCountries.GetFiatsRequest],
) (*connect.Response[pbCountries.GetFiatsResponse], error) {
	qbGetFiats := QbGetListFiats(&uoms.SelectList{})

	SetQBBySelect(
		req.Msg.Select,
		qbGetFiats.
			Join("LEFT JOIN core.countries_uoms ON uoms.id = countries_uoms.uom_id"),
		"countries_uoms",
	)

	sqlStr, args, _ := qbGetFiats.GenerateSQL()

	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		errRes := &pbCommon.Error{
			Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			Package: _package,
			Text:    err.Error(),
		}
		return connect.NewResponse(&pbCountries.GetFiatsResponse{
			Response: &pbCountries.GetFiatsResponse_Error{
				Error: errRes,
			},
		}), err
	}
	defer rows.Close()

	fiatsList := &pbCountries.FiatList{}

	for rows.Next() {
		fiat, err := fiats.ScanFiatUom(rows)
		if err != nil {
			errRes := &pbCommon.Error{
				Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
				Package: _package,
				Text:    err.Error(),
			}
			return connect.NewResponse(&pbCountries.GetFiatsResponse{
				Response: &pbCountries.GetFiatsResponse_Error{
					Error: errRes,
				},
			}), err
		}

		fiatsList.Fiats.List = append(fiatsList.Fiats.List, fiat.Uom)
	}

	return connect.NewResponse(&pbCountries.GetFiatsResponse{
		Response: &pbCountries.GetFiatsResponse_Fiats{
			Fiats: fiatsList,
		},
	}), nil
}

func (s *ServiceServer) SetCryptos(
	ctx context.Context,
	req *connect.Request[pbCountries.SetCryptosRequest],
) (*connect.Response[pbCountries.SetCryptosResponse], error) {
	countryRes, errCountryRes := s.Get(ctx, connect.NewRequest[pbCountries.GetRequest](&pbCountries.GetRequest{
		Select: req.Msg.Select,
	}))

	if errCountryRes != nil {
		return &connect.Response[pbCountries.SetCryptosResponse]{
			Msg: &pbCountries.SetCryptosResponse{
				Response: &pbCountries.SetCryptosResponse_Error{
					Error: countryRes.Msg.GetError(),
				},
			},
		}, errCountryRes
	}

	country := countryRes.Msg.GetCountry()

	handleFn, err := s.GenSetCryptosHandleFn(country, req.Msg.GetCryptos())

	if err != nil {
		return &connect.Response[pbCountries.SetCryptosResponse]{
			Msg: &pbCountries.SetCryptosResponse{
				Response: &pbCountries.SetCryptosResponse_Error{
					Error: &pbCommon.Error{
						Code:    err.Code,
						Package: _package,
						Text:    err.Err.Error(),
					},
				},
			},
		}, err.Err
	}

	if excuteErr := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := handleFn(tx)

		return err
	}); excuteErr != nil {
		return &connect.Response[pbCountries.SetCryptosResponse]{
			Msg: &pbCountries.SetCryptosResponse{
				Response: &pbCountries.SetCryptosResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    excuteErr.Error(),
					},
				},
			},
		}, excuteErr
	}

	fiatsRes, errCryptosRes := s.GetCryptos(ctx, connect.NewRequest[pbCountries.GetCryptosRequest](&pbCountries.GetCryptosRequest{
		Select: req.Msg.Select,
	}))

	if errCryptosRes != nil {
		return &connect.Response[pbCountries.SetCryptosResponse]{
			Msg: &pbCountries.SetCryptosResponse{
				Response: &pbCountries.SetCryptosResponse_Error{
					Error: fiatsRes.Msg.GetError(),
				},
			},
		}, errCryptosRes
	}

	return &connect.Response[pbCountries.SetCryptosResponse]{
		Msg: &pbCountries.SetCryptosResponse{
			Response: &pbCountries.SetCryptosResponse_Cryptos{
				Cryptos: fiatsRes.Msg.GetCryptos(),
			},
		},
	}, nil
}

func (s *ServiceServer) AddCryptos(
	ctx context.Context,
	req *connect.Request[pbCountries.AddCryptosRequest],
) (*connect.Response[pbCountries.AddCryptosResponse], error) {
	countryRes, errCountryRes := s.Get(ctx, connect.NewRequest[pbCountries.GetRequest](&pbCountries.GetRequest{
		Select: req.Msg.Select,
	}))

	if errCountryRes != nil {
		return &connect.Response[pbCountries.AddCryptosResponse]{
			Msg: &pbCountries.AddCryptosResponse{
				Response: &pbCountries.AddCryptosResponse_Error{
					Error: countryRes.Msg.GetError(),
				},
			},
		}, errCountryRes
	}

	handleFn, err := s.GenAddCryptosHandleFn(countryRes.Msg.GetCountry(), req.Msg.GetCryptos())

	if err != nil {
		return &connect.Response[pbCountries.AddCryptosResponse]{
			Msg: &pbCountries.AddCryptosResponse{
				Response: &pbCountries.AddCryptosResponse_Error{
					Error: &pbCommon.Error{
						Code:    err.Code,
						Package: _package,
						Text:    err.Err.Error(),
					},
				},
			},
		}, err.Err
	}

	if excuteErr := crdbpgx.ExecuteTx(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		_, err := handleFn(tx)

		return err
	}); excuteErr != nil {
		return &connect.Response[pbCountries.AddCryptosResponse]{
			Msg: &pbCountries.AddCryptosResponse{
				Response: &pbCountries.AddCryptosResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    excuteErr.Error(),
					},
				},
			},
		}, excuteErr
	}

	fiatsRes, errCryptosRes := s.GetCryptos(ctx, connect.NewRequest[pbCountries.GetCryptosRequest](&pbCountries.GetCryptosRequest{
		Select: req.Msg.Select,
	}))

	if errCryptosRes != nil {
		return &connect.Response[pbCountries.AddCryptosResponse]{
			Msg: &pbCountries.AddCryptosResponse{
				Response: &pbCountries.AddCryptosResponse_Error{
					Error: fiatsRes.Msg.GetError(),
				},
			},
		}, errCryptosRes
	}

	return &connect.Response[pbCountries.AddCryptosResponse]{
		Msg: &pbCountries.AddCryptosResponse{
			Response: &pbCountries.AddCryptosResponse_Cryptos{
				Cryptos: fiatsRes.Msg.GetCryptos(),
			},
		},
	}, nil
}

func (s *ServiceServer) GetCryptos(
	ctx context.Context,
	req *connect.Request[pbCountries.GetCryptosRequest],
) (*connect.Response[pbCountries.GetCryptosResponse], error) {
	qbGetCryptos := QbGetListCryptos(&uoms.SelectList{})

	SetQBBySelect(
		req.Msg.Select,
		qbGetCryptos.
			Join("LEFT JOIN core.countries_uoms ON uoms.id = countries_uoms.uom_id"),
		"countries_uoms",
	)

	sqlStr, args, _ := qbGetCryptos.GenerateSQL()

	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return connect.NewResponse(&pbCountries.GetCryptosResponse{
			Response: &pbCountries.GetCryptosResponse_Error{
				Error: &pbCommon.Error{
					Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
					Package: _package,
					Text:    err.Error(),
				},
			},
		}), err
	}
	defer rows.Close()

	cryptosList := &pbCountries.CryptoList{}

	for rows.Next() {
		fiat, err := cryptos.ScanCryptoUom(rows)
		if err != nil {
			return connect.NewResponse(&pbCountries.GetCryptosResponse{
				Response: &pbCountries.GetCryptosResponse_Error{
					Error: &pbCommon.Error{
						Code:    pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
						Package: _package,
						Text:    err.Error(),
					},
				},
			}), err
		}

		cryptosList.Cryptos.List = append(cryptosList.Cryptos.List, fiat.Uom)
	}

	return connect.NewResponse(&pbCountries.GetCryptosResponse{
		Response: &pbCountries.GetCryptosResponse_Cryptos{
			Cryptos: cryptosList,
		},
	}), nil
}
