package incomes

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"davensi.com/core/gen/addresses"
	"davensi.com/core/gen/countries"
	pbKyc "davensi.com/core/gen/kyc"

	pbCommon "davensi.com/core/gen/common"
	pbIncomes "davensi.com/core/gen/incomes"

	"davensi.com/core/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_fields = "id, type, amount_year, amount_month" +
		", amount_week, amount_day, amount_hour, currency_id" +
		", description, employer, industry, occupation, employment_type" +
		", employment_status, employment_start_date, company, investment_vehicle," +
		" property_address_id, from_country_id, status"
	_tableName = "core.kyc_incomes"
)

type IncomeRepository struct {
	db *pgxpool.Pool
}

func NewIncomeRepository(db *pgxpool.Pool) *IncomeRepository {
	return &IncomeRepository{
		db: db,
	}
}

func setSalary(salary *pbIncomes.SetSalary, cb func(field string, value any)) {
	if salary.AmountYear != nil {
		cb("amount_year", salary.GetAmountYear().GetValue())
	}
	if salary.AmountMonth != nil {
		cb("amount_month", salary.GetAmountMonth().GetValue())
	}
	if salary.AmountWeek != nil {
		cb("amount_week", salary.GetAmountWeek().GetValue())
	}
	if salary.AmountDay != nil {
		cb("amount_day", salary.GetAmountDay().GetValue())
	}
	if salary.AmountHour != nil {
		cb("amount_hour", salary.GetAmountHour().GetValue())
	}
	if salary.Employer != nil {
		cb("employer", salary.GetEmployer())
	}
	if salary.Industry != nil {
		cb("industry", salary.GetIndustry())
	}
	if salary.Occupation != nil {
		cb("occupation", salary.GetOccupation())
	}
	if salary.EmploymentType != nil {
		cb("employment_type", salary.GetEmploymentType())
	}
	if salary.EmploymentStatus != nil {
		cb("employment_status", salary.GetEmploymentStatus())
	}
	if salary.EmploymentStartDate != nil {
		cb("employment_start_date", salary.GetEmploymentStartDate())
	}
}

func setFreelancing(freelancing *pbIncomes.SetFreelancing, cb func(field string, value any)) {
	if freelancing.AmountYear != nil {
		cb("amount_year", freelancing.GetAmountYear().GetValue())
	}
	if freelancing.AmountMonth != nil {
		cb("amount_month", freelancing.GetAmountMonth().GetValue())
	}
	if freelancing.AmountWeek != nil {
		cb("amount_week", freelancing.GetAmountWeek().GetValue())
	}
	if freelancing.AmountDay != nil {
		cb("amount_day", freelancing.GetAmountDay().GetValue())
	}
	if freelancing.AmountHour != nil {
		cb("amount_hour", freelancing.GetAmountHour().GetValue())
	}
	if freelancing.Occupation != nil {
		cb("occupation", freelancing.GetOccupation())
	}
}

func setDividends(dividends *pbIncomes.SetDividends, cb func(field string, value any)) {
	if dividends.AmountYear != nil {
		cb("amount_year", dividends.GetAmountYear().GetValue())
	}
	if dividends.Company != nil {
		cb("company", dividends.GetCompany())
	}
	if dividends.Industry != nil {
		cb("industry", dividends.GetIndustry())
	}
}

func setInvestment(investment *pbIncomes.SetInvestment, cb func(field string, value any)) {
	if investment.AmountYear != nil {
		cb("amount_year", investment.GetAmountYear().GetValue())
	}
	if investment.AmountMonth != nil {
		cb("amount_month", investment.GetAmountMonth().GetValue())
	}
	if investment.InvestmentVehicle != nil {
		cb("investment_vehicle", investment.GetInvestmentVehicle())
	}
}

func setRent(rent *pbIncomes.SetRent, cb func(field string, value any)) {
	if rent.AmountYear != nil {
		cb("amount_year", rent.GetAmountYear().GetValue())
	}
	if rent.AmountHour != nil {
		cb("amount_hour", rent.GetAmountHour().GetValue())
	}
	if rent.AmountMonth != nil {
		cb("amount_month", rent.GetAmountMonth().GetValue())
	}
	if rent.AmountWeek != nil {
		cb("amount_week", rent.GetAmountWeek().GetValue())
	}
	if rent.AmountDay != nil {
		cb("amount_day", rent.GetAmountDay().GetValue())
	}
}

func setPension(pension *pbIncomes.SetPension, cb func(field string, value any)) {
	if pension.FromCountry != nil {
		country := pension.GetFromCountry()
		cb("from_country", country.GetById())
	}
	if pension.AmountMonth != nil {
		cb("amount_month", pension.GetAmountMonth().GetValue())
	}
	if pension.AmountWeek != nil {
		cb("amount_week", pension.GetAmountWeek().GetValue())
	}
	if pension.AmountDay != nil {
		cb("amount_day", pension.GetAmountDay().GetValue())
	}
	if pension.AmountYear != nil {
		cb("amount_year", pension.GetAmountYear().GetValue())
	}
}

func setOther(other *pbIncomes.SetOther, cb func(field string, value any)) {
	if other.AmountYear != nil {
		cb("amount_year", other.GetAmountYear().GetValue())
	}
	if other.AmountMonth != nil {
		cb("amount_month", other.GetAmountMonth().GetValue())
	}
	if other.AmountWeek != nil {
		cb("amount_week", other.GetAmountWeek().GetValue())
	}
	if other.Description != nil {
		cb("description", other.GetDescription())
	}
	if other.AmountDay != nil {
		cb("amount_day", other.GetAmountDay().GetValue())
	}
	if other.AmountHour != nil {
		cb("amount_hour", other.GetAmountHour().GetValue())
	}
}

func (s *IncomeRepository) QbInsert(msg *pbIncomes.CreateRequest) (*util.QueryBuilder, error) {
	qb, err := s.QbInsertWithControlID(msg, "")
	return qb, err
}

func (s *IncomeRepository) QbInsertWithControlID(msg *pbIncomes.CreateRequest, idString string) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleValue := []any{}
	if idString != "" {
		qb.SetInsertField("id", "from_country_id", "property_address_id")
		singleValue = append(singleValue, idString, uuid.NewString(), uuid.NewString())
	} else {
		qb.SetInsertField("from_country_id", "property_address_id")
		singleValue = append(singleValue, uuid.NewString(), uuid.NewString())
	}
	switch msg.GetSelect().(type) {
	case *pbIncomes.CreateRequest_Salary:
		setSalary(msg.GetSalary(), func(field string, value any) {
			qb.SetInsertField(field)
			singleValue = append(singleValue, value)
		})
	case *pbIncomes.CreateRequest_Freelancing:
		setFreelancing(msg.GetFreelancing(), func(field string, value any) {
			qb.SetInsertField(field)
			singleValue = append(singleValue, value)
		})
	case *pbIncomes.CreateRequest_Dividends:
		setDividends(msg.GetDividends(), func(field string, value any) {
			qb.SetInsertField(field)
			singleValue = append(singleValue, value)
		})
	case *pbIncomes.CreateRequest_Investment:
		setInvestment(msg.GetInvestment(), func(field string, value any) {
			qb.SetInsertField(field)
			singleValue = append(singleValue, value)
		})
	case *pbIncomes.CreateRequest_Rent:
		setRent(msg.GetRent(), func(field string, value any) {
			qb.SetInsertField(field)
			singleValue = append(singleValue, value)
		})
	case *pbIncomes.CreateRequest_Pension:
		setPension(msg.GetPension(), func(field string, value any) {
			qb.SetInsertField(field)
			singleValue = append(singleValue, value)
		})
	case *pbIncomes.CreateRequest_Other:
		setOther(msg.GetOther(), func(field string, value any) {
			qb.SetInsertField(field)
			singleValue = append(singleValue, value)
		})
	}

	if msg.Status != nil {
		qb.SetInsertField("status")
		singleValue = append(singleValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleValue)

	qb.SetReturnFields("*")

	return qb, err
}
func (s *IncomeRepository) QbBulkInsert(
	msgs []*pbIncomes.CreateRequest,
	uuidSlices []string,
) (qbList []*util.QueryBuilder, qbListErr error) {
	var qb *util.QueryBuilder
	for index, value := range msgs {
		log.Info().Msgf("value: %v uuidSlices at index %v is %v", value, index, uuidSlices[index])
		qb, qbListErr = s.QbInsertWithControlID(value, uuidSlices[index])
		if qbListErr != nil {
			return nil, qbListErr
		}
		qbList = append(qbList, qb)
	}
	return qbList, nil
}

func (s *IncomeRepository) QbUpdate(msg *pbIncomes.UpdateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Update, _tableName)

	qb.SetReturnFields("*")

	switch msg.GetSelect().(type) {
	case *pbIncomes.UpdateRequest_Salary:
		setSalary(msg.GetSalary(), func(field string, value any) {
			qb.SetUpdate(field, value)
		})
	case *pbIncomes.UpdateRequest_Freelancing:
		setFreelancing(msg.GetFreelancing(), func(field string, value any) {
			qb.SetUpdate(field, value)
		})
	case *pbIncomes.UpdateRequest_Dividends:
		setDividends(msg.GetDividends(), func(field string, value any) {
			qb.SetUpdate(field, value)
		})
	case *pbIncomes.UpdateRequest_Investment:
		setInvestment(msg.GetInvestment(), func(field string, value any) {
			qb.SetUpdate(field, value)
		})
	case *pbIncomes.UpdateRequest_Rent:
		setRent(msg.GetRent(), func(field string, value any) {
			qb.SetUpdate(field, value)
		})
	case *pbIncomes.UpdateRequest_Pension:
		setPension(msg.GetPension(), func(field string, value any) {
			qb.SetUpdate(field, value)
		})
	case *pbIncomes.UpdateRequest_Other:
		setOther(msg.GetOther(), func(field string, value any) {
			qb.SetUpdate(field, value)
		})
	}

	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	qb.Where("id = ?", msg.GetId())

	return qb, nil
}

func (s *IncomeRepository) QbGetOne(msg *pbIncomes.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)
	qb.Where("id = ?", msg.GetId())

	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func addFilterGetList(msg *pbIncomes.GetListRequest, qb *util.QueryBuilder) {
	if msg.AmountYear != nil {
		qb.Where("amount_year = ?", msg.GetAmountYear().GetValue())
	}
	if msg.AmountMonth != nil {
		qb.Where("amount_month = ?", msg.GetAmountMonth().GetValue())
	}
	if msg.AmountWeek != nil {
		qb.Where("amount_week = ?", msg.GetAmountWeek().GetValue())
	}
	if msg.AmountDay != nil {
		qb.Where("amount_day = ?", msg.GetAmountDay().GetValue())
	}
	if msg.AmountHour != nil {
		qb.Where("amount_hour = ?", msg.GetAmountHour().GetValue())
	}
	if msg.Description != nil {
		qb.Where("description = ?", msg.GetDescription())
	}
	if msg.Employer != nil {
		qb.Where("employer = ?", msg.GetEmployer())
	}
	if msg.Industry != nil {
		qb.Where("industry = ?", msg.GetIndustry())
	}
	if msg.Occupation != nil {
		qb.Where("occupation = ?", msg.GetOccupation())
	}
	if msg.EmploymentType != nil {
		qb.Where("employment_type = ?", msg.GetEmploymentType())
	}
	if msg.EmploymentStatus != nil {
		qb.Where("employment_status = ?", msg.GetEmploymentStatus())
	}
	if msg.EmploymentStartDate != nil {
		qb.Where("employment_start_date = ?", msg.GetEmploymentStartDate())
	}
	if msg.Company != nil {
		qb.Where("company = ?", msg.GetCompany())
	}
	if msg.InvestmentVehicle != nil {
		qb.Where("investment_vehicle = ?", msg.GetInvestmentVehicle())
	}
}

func (s *IncomeRepository) QbGetList(msg *pbIncomes.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	if msg.Type != nil {
		uomTypes := msg.GetType().GetList()

		if len(uomTypes) > 0 {
			args := []any{}
			for _, v := range uomTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"uoms.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(uomTypes)), ""), ", "),
				),
				args...,
			)
		}
	}

	if msg.Status != nil {
		uomStatuses := msg.GetStatus().GetList()

		if len(uomStatuses) > 0 {
			args := []any{}
			for _, v := range uomStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"uoms.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(uomStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	addFilterGetList(msg, qb)

	return qb
}

func (s *IncomeRepository) QbDeleteMany(ids []string) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)
	qb.SetUpdate("status", pbKyc.Status_STATUS_CANCELED)

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	args := []any{}
	for _, v := range ids {
		args = append(args, v)
	}
	qb.Where(
		fmt.Sprintf(
			"id IN (%s)",
			strings.Join(strings.Split(strings.Repeat("?", len(ids)), ""), ", "),
		),
		args...,
	)

	return qb, nil
}

func (s *IncomeRepository) ScanRow(row pgx.Row) (*pbIncomes.Income, error) {
	var (
		id                  string
		incommeType         pbIncomes.Type
		amountYear          sql.NullFloat64
		amountMonth         sql.NullFloat64
		amountWeek          sql.NullFloat64
		amountDay           sql.NullFloat64
		amountHour          sql.NullFloat64
		currencyID          sql.NullString
		description         sql.NullString
		employer            sql.NullString
		industry            sql.NullString
		occupation          sql.NullString
		employmentType      sql.NullString
		employmentStatus    sql.NullString
		employmentStartDate sql.NullString
		company             sql.NullString
		investmentVehicle   sql.NullString
		propertyAddressID   string
		fromCountryID       string
		status              pbCommon.Status
	)

	err := row.Scan(
		&id,
		&incommeType,
		&amountYear,
		&amountMonth,
		&amountWeek,
		&amountDay,
		&amountHour,
		&currencyID,
		&description,
		&employer,
		&industry,
		&occupation,
		&employmentType,
		&employmentStatus,
		&employmentStartDate,
		&company,
		&investmentVehicle,
		&propertyAddressID,
		&fromCountryID,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbIncomes.Income{
		Id:   id,
		Type: incommeType,

		AmountYear:  util.FloatToDecimal(util.GetSQLNullFloat(amountYear)),
		AmountMonth: util.FloatToDecimal(util.GetSQLNullFloat(amountMonth)),
		AmountWeek:  util.FloatToDecimal(util.GetSQLNullFloat(amountWeek)),
		AmountDay:   util.FloatToDecimal(util.GetSQLNullFloat(amountDay)),
		AmountHour:  util.FloatToDecimal(util.GetSQLNullFloat(amountHour)),

		Description:         util.GetSQLNullString(description),
		Employer:            util.GetSQLNullString(employer),
		Industry:            util.GetSQLNullString(industry),
		Occupation:          util.GetSQLNullString(occupation),
		EmploymentType:      util.GetSQLNullString(employmentType),
		EmploymentStatus:    util.GetSQLNullString(employmentStatus),
		EmploymentStartDate: util.GetSQLNullString(employmentStartDate),
		Company:             util.GetSQLNullString(company),
		InvestmentVehicle:   util.GetSQLNullString(investmentVehicle),

		PropertyAddress: &addresses.Address{
			Id: propertyAddressID,
		},
		FromCountry: &countries.Country{
			Id: fromCountryID,
		},

		Status: status,
	}, nil
}
func (s *IncomeRepository) ScanRows(rows pgx.Rows) ([]*pbIncomes.Income, error) {
	var result []*pbIncomes.Income
	for rows.Next() {
		income, err := s.ScanRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, income)
	}
	return result, nil
}
