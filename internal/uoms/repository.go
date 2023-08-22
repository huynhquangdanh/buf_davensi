package uoms

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbUoMs "davensi.com/core/gen/uoms"
	pbUoMsConnect "davensi.com/core/gen/uoms/uomsconnect"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName  = "core.uoms"
	_uomsFields = "id, type, symbol, name, icon, managed_decimals, displayed_decimals, reporting_unit, status"
)

type UoMRepository struct {
	pbUoMsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewUoMRepository(db *pgxpool.Pool) *UoMRepository {
	return &UoMRepository{
		db: db,
	}
}

func (s *UoMRepository) QbInsert(msg *pbUoMs.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleUomValue := []any{}

	// Append required value: type, symbol
	qb.SetInsertField("type").SetInsertField("symbol")
	singleUomValue = append(singleUomValue, msg.GetType(), msg.GetSymbol())

	// Append optional fields values
	if msg.Name != nil {
		qb.SetInsertField("name")
		singleUomValue = append(singleUomValue, msg.GetName())
	}
	if msg.Icon != nil {
		qb.SetInsertField("icon")
		singleUomValue = append(singleUomValue, msg.GetIcon())
	}
	if msg.ManagedDecimals != nil {
		qb.SetInsertField("managed_decimals")
		singleUomValue = append(singleUomValue, msg.GetManagedDecimals())
	}
	if msg.DisplayedDecimals != nil {
		qb.SetInsertField("displayed_decimals")
		singleUomValue = append(singleUomValue, msg.GetDisplayedDecimals())
	}
	if msg.ReportingUnit != nil {
		qb.SetInsertField("reporting_unit")
		singleUomValue = append(singleUomValue, msg.GetReportingUnit())
	}
	if msg.Status != nil {
		qb.SetInsertField("status")
		singleUomValue = append(singleUomValue, msg.GetStatus())
	}

	_, err := qb.SetInsertValues(singleUomValue)

	return qb, err
}

func (s *UoMRepository) QbUpdate(msg *pbUoMs.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Type != nil {
		qb.SetUpdate("type", msg.GetType())
	}

	if msg.Symbol != nil {
		qb.SetUpdate("symbol", msg.GetSymbol())
	}

	if msg.Name != nil {
		qb.SetUpdate("name", msg.GetName())
	}

	if msg.Icon != nil {
		qb.SetUpdate("icon", msg.GetIcon())
	}

	if msg.ManagedDecimals != nil {
		qb.SetUpdate("managed_decimals", msg.GetManagedDecimals())
	}

	if msg.DisplayedDecimals != nil {
		qb.SetUpdate("displayed_decimals", msg.GetDisplayedDecimals())
	}

	if msg.ReportingUnit != nil {
		qb.SetUpdate("reporting_unit", msg.GetReportingUnit())
	}

	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	selectSQL, selectArgs := GetFBBySelect(msg.Select).GenerateSQL()
	qb.Where(selectSQL, selectArgs...)

	return qb, nil
}

func (s *UoMRepository) QbGetOne(msg *pbUoMs.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_uomsFields, "uoms"))

	selectSQL, selectArgs := GetFBBySelect(msg.Select).GenerateSQL()
	qb.Where(selectSQL, selectArgs...)

	qb.Where("uoms.status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *UoMRepository) QbGetBySelect(selectList *pbUoMs.SelectList) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_uomsFields, _tableName))

	if selectList != nil && len(selectList.GetList()) > 0 {
		selectListFB := util.CreateFilterBracket("OR")
		for _, selectItem := range selectList.GetList() {
			selectSQL, selectArgs := GetFBBySelect(selectItem).GenerateSQL()
			selectListFB.SetFilter(selectSQL, selectArgs...)
		}
		selectListSQL, selectListArgs := selectListFB.GenerateSQL()
		qb.Where(selectListSQL, selectListArgs...)
	}

	return qb
}

func GetFBBySelect(selectUom *pbUoMs.Select) *util.FilterBracket {
	selectFB := util.CreateFilterBracket("AND")

	if selectUom != nil && selectUom.Select != nil {
		switch selectUom.GetSelect().(type) {
		case *pbUoMs.Select_ById:
			selectFB.SetFilter("uoms.id = ?", selectUom.GetById())
		case *pbUoMs.Select_ByTypeSymbol:
			selectFB.
				SetFilter("uoms.symbol = ?", selectUom.GetByTypeSymbol().GetSymbol()).
				SetFilter("uoms.type = ?", selectUom.GetByTypeSymbol().GetType())
		}
	}

	return selectFB
}

func (s *UoMRepository) QbGetList(msg *pbUoMs.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_uomsFields, "uoms"))

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
	if msg.Symbol != nil {
		qb.Where("uoms.symbol LIKE '%' || ? || '%'", msg.GetSymbol())
	}
	if msg.Name != nil {
		qb.Where("uoms.name LIKE '%' || ? || '%'", msg.GetName())
	}
	if msg.Icon != nil {
		qb.Where("uoms.icon LIKE '%' || ? || '%'", msg.GetIcon())
	}
	if msg.ManagedDecimals != nil {
		decimalFilter := common.GetDecimalsFB(msg.GetManagedDecimals(), "uoms.managed_decimals")
		sqlStr, args := decimalFilter.GenerateSQL()

		qb.Where(sqlStr, args...)
	}
	if msg.DisplayedDecimals != nil {
		decimalFilter := common.GetDecimalsFB(msg.GetDisplayedDecimals(), "uoms.displayed_decimals")
		sqlStr, args := decimalFilter.GenerateSQL()

		qb.Where(sqlStr, args...)
	}
	if msg.ReportingUnit != nil {
		qb.Where("uoms.reporting_unit = ?", msg.GetReportingUnit())
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

	return qb
}

func (s *UoMRepository) ScanRow(row pgx.Row) (*pbUoMs.UoM, error) {
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
	)
	if err != nil {
		return nil, err
	}

	return &pbUoMs.UoM{
		Id:                id,
		Type:              uomType,
		Symbol:            symbol,
		Name:              util.GetSQLNullString(name),
		Icon:              util.GetSQLNullString(icon),
		ManagedDecimals:   managedDecimals,
		DisplayedDecimals: displayedDecimals,
		ReportingUnit:     reportingUnit,
		Status:            status,
	}, nil
}
