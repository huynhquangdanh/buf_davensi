package userincomes

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"davensi.com/core/gen/common"
	pbIncomes "davensi.com/core/gen/incomes"
	pbKyc "davensi.com/core/gen/kyc"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName         = "core.users_incomes"
	_userIncomesFields = "user_id, label, income_id, status"
)

type UserIncomesRepository struct {
	db *pgxpool.Pool
}

var (
	singleRepo *UserIncomesRepository
	once       sync.Once
)

func NewUserContactRepository(db *pgxpool.Pool) *UserIncomesRepository {
	return &UserIncomesRepository{
		db: db,
	}
}

func GetSingletonRepository(db *pgxpool.Pool) *UserIncomesRepository {
	once.Do(func() {
		singleRepo = NewUserContactRepository(db)
	})
	return singleRepo
}

type GetUserIncome struct {
	UserID   string
	Label    string
	IncomeID string
	Status   common.Status
}

func NewUserIncomeRepository(db *pgxpool.Pool) *UserIncomesRepository {
	return &UserIncomesRepository{
		db: db,
	}
}

func (*UserIncomesRepository) QbGet(
	userIncome GetUserIncome,
) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)

	if userIncome.UserID != "" {
		qb.Where("user_id = ?", userIncome.UserID)
	}
	if userIncome.Label != "" {
		qb.Where("label = ?", userIncome.Label)
	}
	if userIncome.IncomeID != "" {
		qb.Where("income_id = ?", userIncome.IncomeID)
	}
	qb.Where("status = ?", userIncome.Status)

	return qb
}

func (*UserIncomesRepository) QbUpdate(
	userID string,
	income *pbIncomes.SetLabeledIncome,
) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Update, _tableName)

	qb.SetUpdate("status", income.GetStatus())
	qb.Where("user_id = ?", userID)
	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	return qb, nil
}

func (*UserIncomesRepository) QbUpsertUserIncomes(
	userID string,
	incomes []*pbIncomes.SetLabeledIncome,
	command string,
) (*util.QueryBuilder, error) {
	var queryType util.QueryType
	if command == "upsert" {
		queryType = util.Upsert
	} else {
		queryType = util.Insert
	}
	qb := util.CreateQueryBuilder(queryType, _tableName)
	qb.SetInsertField("user_id", "label", "income_id", "status")

	for _, income := range incomes {
		_, err := qb.SetInsertValues([]any{
			userID,
			income.GetLabel(),
			income.GetId(),
			income.GetStatus().Enum(),
		})
		if err != nil {
			return nil, err
		}
	}

	return qb, nil
}

func (*UserIncomesRepository) ScanRow(row pgx.Row) (*pbIncomes.LabeledIncome, error) {
	var (
		userID   string
		label    string
		incomeID string
		status   *pbKyc.Status
	)
	err := row.Scan(&userID, &label, &incomeID, &status)
	if err != nil {
		return nil, err
	}
	return &pbIncomes.LabeledIncome{
		Label: label,
		Income: &pbIncomes.Income{
			Id: incomeID,
		},
		Status: *status.Enum(),
	}, nil
}

func (*UserIncomesRepository) ScanMultiRows(rows pgx.Rows) (*pbIncomes.LabeledIncomeList, error) {
	var (
		userID   string
		incomeID pgtype.Text
		label    string
		status   uint32
	)

	userIncomes := []*pbIncomes.LabeledIncome{}
	for rows.Next() {
		err := rows.Scan(
			&userID,
			&label,
			&incomeID,
			&status,
		)
		if err != nil {
			return nil, err
		}
		userIncomes = append(userIncomes, &pbIncomes.LabeledIncome{
			Label: label,
			Income: &pbIncomes.Income{
				Id: incomeID.String,
			},
			Status: pbKyc.Status(status),
		})
	}
	return &pbIncomes.LabeledIncomeList{
		List: userIncomes,
	}, nil
}
func (*UserIncomesRepository) QbBulkInsert(userID string, valueSlices []*pbIncomes.SetLabeledIncome) (*util.QueryBuilder, error) {
	// Split the string using the comma as a delimiter
	fieldNames := strings.Split(_userIncomeFields, ",")

	// Trim any leading or trailing whitespaces from the field names
	for i, fieldName := range fieldNames {
		fieldNames[i] = strings.TrimSpace(fieldName)
	}
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	qb.SetInsertField(fieldNames...)
	if !reflect.ValueOf(valueSlices).IsNil() {
		for _, value := range valueSlices {
			_, err := qb.SetInsertValues([]any{
				userID,
				value.GetLabel(),
				value.GetId(),
				value.GetStatus().Enum(),
			})
			if err != nil {
				return nil, err
			}
		}
		return qb, nil
	}
	return nil, errors.New("value slices is not supported for bulk insert")
}
func (*UserIncomesRepository) QbUserAdditionInfoUpdateMany(
	userID string,
	valueSlices []*pbIncomes.SetLabeledIncome,
) ([]*util.QueryBuilder, error) {
	var qbList []*util.QueryBuilder
	if !reflect.ValueOf(valueSlices).IsNil() {
		for _, value := range valueSlices {
			qb := util.CreateQueryBuilder(util.Update, _tableName)
			if value.Label != "" {
				qb.SetUpdate("label", value.Label)
			}
			if value.GetStatus() != pbKyc.Status_STATUS_UNSPECIFIED {
				qb.SetUpdate("status", value.GetStatus())
			}
			if qb.IsUpdatable() {
				qb.Where("user_id = ? AND income_id = ? AND status != ?", userID, value.GetId(), pbKyc.Status_STATUS_CANCELED)
				qbList = append(qbList, qb)
			} else {
				return nil, errors.New("cannot update without new value")
			}
		}
		return qbList, nil
	}
	return nil, errors.New("value slices is not supported for bulk insert")
}
func (*UserIncomesRepository) ScanMultiLabelRows(rows pgx.Rows) ([]*pbIncomes.LabeledIncome, error) {
	var (
		userID   string
		label    pgtype.Text
		incomeID pgtype.Text
		status   uint32
	)
	var result []*pbIncomes.LabeledIncome
	var i int
	for rows.Next() {
		err := rows.Scan(
			&userID,
			&label,
			&incomeID,
			&status,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &pbIncomes.LabeledIncome{
			Label: label.String,
			Income: &pbIncomes.Income{
				Id: incomeID.String,
			},
			Status: pbKyc.Status(status),
		})
		i++
	}
	return result, nil
}
