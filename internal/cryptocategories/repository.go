package cryptocategories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbCryptocategories "davensi.com/core/gen/cryptocategories"
	pbCryptocategoriesConnect "davensi.com/core/gen/cryptocategories/cryptocategoriesconnect"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_fields    = "id, name, icon, status"
	_tableName = "core.cryptocategories"
)

type CryptoCategoryRepository struct {
	pbCryptocategoriesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewCryptoCategoryRepository(db *pgxpool.Pool) *CryptoCategoryRepository {
	return &CryptoCategoryRepository{
		db: db,
	}
}

func handleTypicalCryptoCategoryFields(msg *pbCryptocategories.CryptoCategory, handleFn func(field string, value any)) {
	if msg.Icon != nil {
		handleFn("icon", msg.GetIcon())
	}
}

func (s *CryptoCategoryRepository) QbInsert(msg *pbCryptocategories.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("name")
	singleValue = append(singleValue, msg.GetName())

	// Append optional fields values
	handleTypicalCryptoCategoryFields(
		&pbCryptocategories.CryptoCategory{
			Icon: msg.Icon,
		},
		func(field string, value any) {
			qb.SetInsertField(field)
			singleValue = append(singleValue, value)
		},
	)
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleValue = append(singleValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleValue)

	return qb, err
}

func (s *CryptoCategoryRepository) QbUpdate(msg *pbCryptocategories.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Name != nil {
		qb.SetUpdate("name", msg.GetName())
	}
	handleTypicalCryptoCategoryFields(
		&pbCryptocategories.CryptoCategory{
			Icon: msg.Icon,
		},
		func(field string, value any) {
			qb.SetUpdate(field, value)
		},
	)
	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch msg.Select.(type) {
	case *pbCryptocategories.UpdateRequest_ById:
		qb.Where("id = ?", msg.GetById())
	case *pbCryptocategories.UpdateRequest_ByName:
		qb.Where("name = ?", msg.GetByName())
	}

	return qb, nil
}

func (s *CryptoCategoryRepository) QbGetOne(msg *pbCryptocategories.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	switch msg.Select.(type) {
	case *pbCryptocategories.GetRequest_ById:
		qb.Where("id = ?", msg.GetById())
	case *pbCryptocategories.GetRequest_ByName:
		qb.Where("name = ?", msg.GetByName())
	}

	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *CryptoCategoryRepository) QbGetList(msg *pbCryptocategories.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	if msg.Name != nil {
		qb.Where("name LIKE '%' || ? || '%'", msg.GetName())
	}
	handleTypicalCryptoCategoryFields(
		&pbCryptocategories.CryptoCategory{
			Icon: msg.Icon,
		},
		func(field string, value any) {
			qb.Where(
				fmt.Sprintf("%s LIKE '%%' || ? || '%%'", field),
				msg.GetName(),
			)
		},
	)
	if msg.Status != nil {
		statuses := msg.GetStatus().GetList()

		if len(statuses) > 0 {
			args := []any{}
			for _, v := range statuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(statuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *CryptoCategoryRepository) ScanRow(row pgx.Row) (*pbCryptocategories.CryptoCategory, error) {
	var (
		id     string
		name   string
		icon   sql.NullString
		status pbCommon.Status
	)

	err := row.Scan(
		&id,
		&name,
		&icon,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbCryptocategories.CryptoCategory{
		Id:     id,
		Name:   name,
		Icon:   util.GetSQLNullString(icon),
		Status: status,
	}, nil
}
