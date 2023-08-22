package ledgers

import (
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbLedgers "davensi.com/core/gen/ledgers"
	pbLedgersConnect "davensi.com/core/gen/ledgers/ledgersconnect"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName = "core.ledgers"
	_fields    = "id, name, status"
)

type LedgerRepository struct {
	pbLedgersConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewLedgerRepository(db *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{
		db: db,
	}
}

func (s *LedgerRepository) QbInsert(msg *pbLedgers.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleLedgerValue := []any{}

	// Append required value: name
	qb.SetInsertField("name")
	singleLedgerValue = append(singleLedgerValue, msg.GetName())

	// Append optional fields values
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleLedgerValue = append(singleLedgerValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleLedgerValue)

	return qb, err
}

func (s *LedgerRepository) QbUpdate(msg *pbLedgers.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Name != nil {
		qb.SetUpdate("name", msg.GetName())
	}

	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch msg.GetSelect().(type) {
	case *pbLedgers.UpdateRequest_ById:
		qb.Where("id = ?", msg.GetById())
	case *pbLedgers.UpdateRequest_ByName:
		qb.Where("name = ?", msg.GetByName())
	}

	return qb, nil
}

func (s *LedgerRepository) QbGetOne(msg *pbLedgers.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	switch msg.GetSelect().(type) {
	case *pbLedgers.GetRequest_ById:
		qb.Where("id = ?", msg.GetById())
	case *pbLedgers.GetRequest_ByName:
		qb.Where("name = ?", msg.GetByName())
	}

	qb.Where("status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *LedgerRepository) QbGetList(msg *pbLedgers.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	if msg.Name != nil {
		qb.Where("name LIKE '%' || ? || '%'", msg.GetName())
	}
	if msg.Status != nil {
		ledgerStatuses := msg.GetStatus().GetList()

		if len(ledgerStatuses) > 0 {
			args := []any{}
			for _, v := range ledgerStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(ledgerStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *LedgerRepository) ScanRow(row pgx.Row) (*pbLedgers.Ledger, error) {
	var (
		id     string
		name   string
		status pbCommon.Status
	)

	err := row.Scan(
		&id,
		&name,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbLedgers.Ledger{
		Id:     id,
		Name:   name,
		Status: status,
	}, nil
}
