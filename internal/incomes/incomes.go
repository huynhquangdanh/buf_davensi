package incomes

import (
	"context"
	"fmt"
	"sync"

	pbAddresses "davensi.com/core/gen/addresses"
	pbCountries "davensi.com/core/gen/countries"
	pbUoms "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/util"

	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	pbCommon "davensi.com/core/gen/common"
	pbIncomes "davensi.com/core/gen/incomes"
	pbIncomesconnect "davensi.com/core/gen/incomes/incomesconnect"
	"davensi.com/core/internal/common"
)

const (
	_package          = "incomes"
	_entityName       = "Income"
	_entityNamePlural = "Incomes"
)

var (
	once                   sync.Once
	singletonServiceServer *ServiceServer
)

func GetSingletonServiceServer(db *pgxpool.Pool) *ServiceServer {
	once.Do(func() {
		singletonServiceServer = NewServiceServer(db)
	})
	return singletonServiceServer
}

// ServiceServer implements the IncomesService API
type ServiceServer struct {
	Repo IncomeRepository
	pbIncomesconnect.ServiceHandler
	db *pgxpool.Pool
}

func NewServiceServer(db *pgxpool.Pool) *ServiceServer {
	return &ServiceServer{
		Repo: *NewIncomeRepository(db),
		db:   db,
	}
}

func (s *ServiceServer) Create(
	ctx context.Context,
	req *connect.Request[pbIncomes.CreateRequest],
) (*connect.Response[pbIncomes.CreateResponse], error) {
	if _errno, validateErr := s.ValidateCreation(req.Msg); validateErr != nil {
		log.Error().Err(validateErr)
		return connect.NewResponse(&pbIncomes.CreateResponse{
			Response: &pbIncomes.CreateResponse_Error{
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
		return connect.NewResponse(&pbIncomes.CreateResponse{
			Response: &pbIncomes.CreateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	var newIncome *pbIncomes.Income

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL \"" + sqlStr + "\"")
	newIncome, errExcute := common.ExecuteTxWrite[pbIncomes.Income](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanRow,
	)
	if errExcute != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"creating",
			_entityName,
		)
		log.Error().Err(errExcute).Msg(_err.Error())
		return connect.NewResponse(&pbIncomes.CreateResponse{
			Response: &pbIncomes.CreateResponse_Error{
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
		_entityName, newIncome.GetId(),
	)
	return connect.NewResponse(&pbIncomes.CreateResponse{
		Response: &pbIncomes.CreateResponse_Income{
			Income: newIncome,
		},
	}), nil
}

func (s *ServiceServer) Update(
	ctx context.Context,
	req *connect.Request[pbIncomes.UpdateRequest],
) (*connect.Response[pbIncomes.UpdateResponse], error) {
	if errQueryUpdate := validateQueryUpdate(req.Msg); errQueryUpdate != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", errQueryUpdate.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbIncomes.UpdateResponse{
			Response: &pbIncomes.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	qb, genSQLError := s.Repo.QbUpdate(req.Msg)
	if genSQLError != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating '"+_entityName+"'", genSQLError.Error())
		log.Error().Err(_err)
		return connect.NewResponse(&pbIncomes.UpdateResponse{
			Response: &pbIncomes.UpdateResponse_Error{
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
	updatedIncome, errExcute := common.ExecuteTxWrite[pbIncomes.Income](
		ctx,
		s.db,
		sqlStr,
		args,
		s.Repo.ScanRow,
	)

	if errExcute != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())],
			"updating",
			_entityName,
			sel,
		)
		log.Error().Msg(_err.Error())
		return connect.NewResponse(&pbIncomes.UpdateResponse{
			Response: &pbIncomes.UpdateResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	log.Info().Msgf("%s with %s updated successfully", _entityName, sel)
	return connect.NewResponse(&pbIncomes.UpdateResponse{
		Response: &pbIncomes.UpdateResponse_Income{
			Income: updatedIncome,
		},
	}), nil
}

func (s *ServiceServer) Get(
	ctx context.Context,
	req *connect.Request[pbIncomes.GetRequest],
) (*connect.Response[pbIncomes.GetResponse], error) {
	qb := s.Repo.QbGetOne(req.Msg)
	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, err := s.db.Query(ctx, sqlstr, sqlArgs...)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbIncomes.GetResponse{
			Response: &pbIncomes.GetResponse_Error{
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
		return connect.NewResponse(&pbIncomes.GetResponse{
			Response: &pbIncomes.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}
	icome, err := s.Repo.ScanRow(rows)
	if err != nil {
		_errno := pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR
		_err := fmt.Errorf(common.Errors[uint32(_errno.Number())], "fetching", _entityName, sel)
		log.Error().Err(err).Msg(_err.Error())
		return connect.NewResponse(&pbIncomes.GetResponse{
			Response: &pbIncomes.GetResponse_Error{
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
		return connect.NewResponse(&pbIncomes.GetResponse{
			Response: &pbIncomes.GetResponse_Error{
				Error: &pbCommon.Error{
					Code:    _errno,
					Package: _package,
					Text:    _err.Error(),
				},
			},
		}), _err
	}

	// Start building the response from here
	return connect.NewResponse(&pbIncomes.GetResponse{
		Response: &pbIncomes.GetResponse_Income{
			Income: icome,
		},
	}), nil
}

func (s *ServiceServer) GetList(
	ctx context.Context,
	req *connect.Request[pbIncomes.GetListRequest],
	res *connect.ServerStream[pbIncomes.GetListResponse],
) error {
	qb := s.Repo.QbGetList(req.Msg)

	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	sqlStr, args, _ := qb.GenerateSQL()

	log.Info().Msg("Executing SQL: " + sqlStr)
	rows, err := s.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return common.StreamError(
			_entityName,
			pbCommon.ErrorCode_ERROR_CODE_DB_ERROR,
			err,
			func(errStream *pbCommon.Error) error {
				return res.Send(&pbIncomes.GetListResponse{
					Response: &pbIncomes.GetListResponse_Error{
						Error: errStream,
					},
				})
			},
		)
	}

	defer rows.Close()

	// Start building the response from here
	for rows.Next() {
		icome, err := s.Repo.ScanRow(rows)
		if err != nil {
			return common.StreamError(
				_entityName,
				pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR,
				err,
				func(errStream *pbCommon.Error) error {
					return res.Send(&pbIncomes.GetListResponse{
						Response: &pbIncomes.GetListResponse_Error{
							Error: errStream,
						},
					})
				},
			)
		}

		if errSend := res.Send(&pbIncomes.GetListResponse{
			Response: &pbIncomes.GetListResponse_Income{
				Income: icome,
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
	req *connect.Request[pbIncomes.DeleteRequest],
) (*connect.Response[pbIncomes.DeleteResponse], error) {
	terminatedStatus := pbCommon.Status_STATUS_TERMINATED
	icome := &pbIncomes.UpdateRequest{
		Id:     req.Msg.GetId(),
		Status: &terminatedStatus,
	}
	deleteReq := connect.NewRequest(icome)
	deletedIncome, err := s.Update(ctx, deleteReq)
	if err != nil {
		log.Error().Err(err)
		return connect.NewResponse(&pbIncomes.DeleteResponse{
			Response: &pbIncomes.DeleteResponse_Error{
				Error: &pbCommon.Error{
					Code:    deletedIncome.Msg.GetError().Code,
					Package: _package,
					Text:    "update failed: " + deletedIncome.Msg.GetError().Text,
				},
			},
		}), err
	}

	return connect.NewResponse(&pbIncomes.DeleteResponse{
		Response: &pbIncomes.DeleteResponse_Income{
			Income: deletedIncome.Msg.GetIncome(),
		},
	}), nil
}
func (s *ServiceServer) MapFromCreateRequestToIncome(req *pbIncomes.CreateRequest) (income *pbIncomes.Income) {
	switch convertedReq := req.GetSelect().(type) {
	case *pbIncomes.CreateRequest_Dividends:
		data := convertedReq.Dividends
		income = &pbIncomes.Income{
			AmountYear: data.AmountYear,
			Company:    data.Company,
			Industry:   data.Industry,
			Currency:   getCurrencySelectData(data.Currency),
		}
	case *pbIncomes.CreateRequest_Freelancing:
		data := convertedReq.Freelancing
		income = &pbIncomes.Income{
			AmountYear:  data.AmountYear,
			AmountMonth: data.AmountMonth,
			AmountWeek:  data.AmountWeek,
			AmountDay:   data.AmountDay,
			AmountHour:  data.AmountHour,
			Currency:    getCurrencySelectData(data.Currency),
			Occupation:  data.Occupation,
		}
	case *pbIncomes.CreateRequest_Investment:
		data := convertedReq.Investment
		income = &pbIncomes.Income{
			AmountYear:        data.AmountYear,
			AmountMonth:       data.AmountMonth,
			Currency:          getCurrencySelectData(data.Currency),
			InvestmentVehicle: data.InvestmentVehicle,
		}
	case *pbIncomes.CreateRequest_Pension:
		data := convertedReq.Pension
		income = &pbIncomes.Income{
			AmountYear:  data.AmountYear,
			AmountMonth: data.AmountMonth,
			AmountWeek:  data.AmountWeek,
			AmountDay:   data.AmountDay,
			Currency:    getCurrencySelectData(data.Currency),
			FromCountry: getFromCountrySelectData(data.FromCountry),
		}
	case *pbIncomes.CreateRequest_Rent:
		data := convertedReq.Rent
		income = &pbIncomes.Income{
			AmountYear:  data.AmountYear,
			AmountMonth: data.AmountMonth,
			AmountWeek:  data.AmountWeek,
			AmountDay:   data.AmountDay,
			AmountHour:  data.AmountHour,
			Currency:    getCurrencySelectData(data.Currency),
			PropertyAddress: &pbAddresses.Address{
				Type:       data.PropertyAddress.GetType(),
				Country:    getFromCountrySelectData(data.PropertyAddress.Country),
				Building:   data.PropertyAddress.Building,
				Floor:      data.PropertyAddress.Floor,
				Unit:       data.PropertyAddress.Unit,
				StreetNum:  data.PropertyAddress.StreetNum,
				StreetName: data.PropertyAddress.StreetName,
				District:   data.PropertyAddress.District,
				Locality:   data.PropertyAddress.Locality,
				ZipCode:    data.PropertyAddress.ZipCode,
				Region:     data.PropertyAddress.Region,
				State:      data.PropertyAddress.State,
				Status:     data.PropertyAddress.GetStatus(),
			},
		}
	case *pbIncomes.CreateRequest_Salary:
		data := convertedReq.Salary
		income = &pbIncomes.Income{
			AmountYear:          data.AmountYear,
			AmountMonth:         data.AmountMonth,
			AmountWeek:          data.AmountWeek,
			AmountDay:           data.AmountDay,
			AmountHour:          data.AmountHour,
			Currency:            getCurrencySelectData(data.Currency),
			Employer:            data.Employer,
			Industry:            data.Industry,
			Occupation:          data.Occupation,
			EmploymentType:      data.EmploymentType,
			EmploymentStatus:    data.EmploymentStatus,
			EmploymentStartDate: data.EmploymentStartDate,
		}
	case *pbIncomes.CreateRequest_Other:
		data := convertedReq.Other
		income = &pbIncomes.Income{
			AmountYear:  data.AmountYear,
			AmountMonth: data.AmountMonth,
			AmountWeek:  data.AmountWeek,
			AmountDay:   data.AmountDay,
			AmountHour:  data.AmountHour,
			Currency:    getCurrencySelectData(data.Currency),
			Description: data.Description,
		}
	}
	return income
}

func getCurrencySelectData(uomSelect *pbUoms.Select) (currencyData *pbUoms.UoM) {
	switch uomSelect.GetSelect().(type) {
	case *pbUoms.Select_ById:
		currencyData = &pbUoms.UoM{
			Id: uomSelect.GetById(),
		}
	case *pbUoms.Select_ByTypeSymbol:
		currencyData = &pbUoms.UoM{
			Type:   uomSelect.GetByTypeSymbol().GetType(),
			Symbol: uomSelect.GetByTypeSymbol().GetSymbol(),
		}
	}
	return currencyData
}
func getFromCountrySelectData(countrySelect *pbCountries.Select) (countryData *pbCountries.Country) {
	switch countrySelect.GetSelect().(type) {
	case *pbCountries.Select_ById:
		countryData = &pbCountries.Country{
			Id: countrySelect.GetById(),
		}
	case *pbCountries.Select_ByCode:
		countryData = &pbCountries.Country{
			Code: countrySelect.GetByCode(),
		}
	}
	return countryData
}
func (s *ServiceServer) GetSpecificIncome(ctx context.Context, incomeID string) (income *pbIncomes.Income, getIncomeErr error) {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields).Where("id = ? AND status != ?", incomeID, pbCommon.Status_STATUS_TERMINATED)

	sqlstr, sqlArgs, sel := qb.GenerateSQL()
	log.Info().Msg("Executing SQL \"" + sqlstr + "\"")

	rows, getIncomeErr := s.db.Query(ctx, sqlstr, sqlArgs...)
	if getIncomeErr != nil {
		return nil, getIncomeErr
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND.Number())], _entityName, sel)
	}
	income, getIncomeErr = s.Repo.ScanRow(rows)
	if getIncomeErr != nil {
		return nil, fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_DB_FIELD_SCAN_ERROR.Number())], "fetching", _entityName, sel)
	}
	if rows.Next() {
		return nil, fmt.Errorf(common.Errors[uint32(pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND.Number())], _entityNamePlural, sel)
	}
	return income, nil
}
