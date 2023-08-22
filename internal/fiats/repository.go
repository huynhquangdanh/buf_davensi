package fiats

import (
	"database/sql"
	"errors"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbFiats "davensi.com/core/gen/fiats"
	pbUoMs "davensi.com/core/gen/uoms"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName = "core.fiats"
)

type FiatRepository struct {
	db *pgxpool.Pool
}

func NewFiatRepository(db *pgxpool.Pool) *FiatRepository {
	return &FiatRepository{
		db: db,
	}
}

func (s *FiatRepository) QbGetOne(_ *pbUoMs.GetRequest, uomsQb *util.QueryBuilder) *util.QueryBuilder {
	uomsQb.Join(fmt.Sprintf("LEFT JOIN %s ON fiats.id = uoms.id", _tableName))
	uomsQb.Select("fiats.iso4217_num")
	uomsQb.Where("uoms.type = ?", pbUoMs.Type_TYPE_FIAT)
	uomsQb.Where("uoms.status = ?", pbCommon.Status_STATUS_ACTIVE)

	return uomsQb
}

func (s *FiatRepository) QbGetList(msg *pbFiats.GetListRequest, uomsQb *util.QueryBuilder) *util.QueryBuilder {
	uomsQb.Join(fmt.Sprintf("LEFT JOIN %s ON fiats.id = uoms.id", _tableName))
	uomsQb.Select("fiats.iso4217_num")
	uomsQb.Where("uoms.type = ?", pbUoMs.Type_TYPE_FIAT)

	if msg.Iso4217Num != nil {
		uomsQb.Where("fiats.iso4217_num = ?", msg.GetIso4217Num())
	}

	return uomsQb
}

func (s *FiatRepository) QbUpdate(msg *pbFiats.UpdateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Iso4217Num != nil {
		qb.SetUpdate("fiats.iso4217_num", msg.GetIso4217Num())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch msg.GetUom().GetSelect().Select.(type) {
	case *pbUoMs.Select_ById:
		qb.Where("fiats.id = ?", msg.GetUom().GetSelect().GetById())
	case *pbUoMs.Select_ByTypeSymbol:
		qb.Where(
			"fiats.id = (SELECT uoms.id FROM uoms WHERE uoms.type = ? AND uoms.symbol = ?)",
			msg.GetUom().GetSelect().GetByTypeSymbol().GetType(),
			msg.GetUom().GetSelect().GetByTypeSymbol().GetSymbol(),
		)
	}

	return qb, nil
}

func (s *FiatRepository) QbInsert(msg *pbFiats.CreateRequest, uom *pbUoMs.UoM) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("id")
	singleValue = append(singleValue, uom.GetId())

	// Append optional fields values
	if msg.Iso4217Num != nil {
		qb.SetInsertField("iso4217_num")
		singleValue = append(singleValue, msg.GetIso4217Num())
	}

	_, err := qb.SetInsertValues(singleValue)

	return qb, err
}

func ScanFiatUom(row pgx.Row) (*pbFiats.Fiat, error) {
	var (
		id      string
		uomType pbUoMs.Type
		symbol  string
		// Nullable field must use sql.* type or a scan error will be thrown
		name              sql.NullString
		icon              sql.NullString
		managedDecimals   uint32
		displayedDecimals uint32
		reportingUnit     bool
		status            pbCommon.Status
		iso4217Num        sql.NullString
	)

	err := row.Scan(
		&id,
		&uomType,
		&symbol,
		&name,
		&icon,
		&managedDecimals,
		&displayedDecimals,
		&reportingUnit,
		&status,
		&iso4217Num,
	)
	if err != nil {
		return nil, err
	}

	return &pbFiats.Fiat{
		Uom: &pbUoMs.UoM{
			Id:                id,
			Type:              uomType,
			Symbol:            symbol,
			Name:              util.GetSQLNullString(name),
			Icon:              util.GetSQLNullString(icon),
			ManagedDecimals:   managedDecimals,
			DisplayedDecimals: displayedDecimals,
			ReportingUnit:     reportingUnit,
			Status:            status,
		},
		Iso4217Num: util.GetSQLNullString(iso4217Num),
	}, nil
}

func ScanFiat(row pgx.Row) (*pbFiats.Fiat, error) {
	var (
		id         string
		iso4217Num string
	)

	err := row.Scan(
		&id,
		&iso4217Num,
	)
	if err != nil {
		return nil, err
	}

	return &pbFiats.Fiat{
		Uom: &pbUoMs.UoM{
			Id: id,
		},
		Iso4217Num: &iso4217Num,
	}, nil
}
