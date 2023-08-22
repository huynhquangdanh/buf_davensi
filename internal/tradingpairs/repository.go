package tradingpairs

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbTradingPairs "davensi.com/core/gen/tradingpairs"
	pbTradingPairsConnect "davensi.com/core/gen/tradingpairs/tradingpairsconnect"
	pbUoMs "davensi.com/core/gen/uoms"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName   = "core.tradingpairs"
	_banksFields = "id, symbol, quantity_uom_id, quantity_decimals, price_uom_id, price_decimals, volume_decimals, status"
)

type TradingPairRepository struct {
	pbTradingPairsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewTradingPairRepository(db *pgxpool.Pool) *TradingPairRepository {
	return &TradingPairRepository{
		db: db,
	}
}

func GetFBBySelect(selectTradingPair *pbTradingPairs.Select) *util.FilterBracket {
	selectFB := util.CreateFilterBracket("OR")

	if selectTradingPair != nil && selectTradingPair.Select != nil {
		switch selectTradingPair.GetSelect().(type) {
		case *pbTradingPairs.Select_ById:
			selectFB.SetFilter("tradingpairs.id = ?", selectTradingPair.GetById())
		case *pbTradingPairs.Select_BySymbol:
			selectFB.SetFilter(
				"tradingpairs.symbol = ?", selectTradingPair.GetBySymbol(),
			)
		}
	}

	return selectFB
}

func (s *TradingPairRepository) QbInsert(req *pbTradingPairs.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleTradingPairValue := []any{}

	// Append required value: login
	qb.SetInsertField("symbol")
	singleTradingPairValue = append(singleTradingPairValue, req.GetSymbol())
	qb.SetInsertField("quantity_uom_id")
	singleTradingPairValue = append(singleTradingPairValue, req.GetQuantityUom().GetById())
	qb.SetInsertField("price_uom_id")
	singleTradingPairValue = append(singleTradingPairValue, req.GetPriceUom().GetById())
	qb.SetInsertField("status")
	singleTradingPairValue = append(singleTradingPairValue, req.GetStatus())

	// Append optional fields values
	if req.QuantityDecimals != nil {
		qb.SetInsertField("quantity_decimals")
		singleTradingPairValue = append(singleTradingPairValue, req.GetQuantityDecimals())
	}
	if req.PriceDecimals != nil {
		qb.SetInsertField("price_decimals")
		singleTradingPairValue = append(singleTradingPairValue, req.GetPriceDecimals())
	}
	if req.VolumeDecimals != nil {
		qb.SetInsertField("volume_decimals")
		singleTradingPairValue = append(singleTradingPairValue, req.GetVolumeDecimals())
	}
	_, err := qb.SetInsertValues(singleTradingPairValue)

	return qb, err
}

func (s *TradingPairRepository) QbUpdate(req *pbTradingPairs.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)
	if req.Symbol != nil {
		qb.SetUpdate("symbol", req.GetSymbol())
	}
	if req.QuantityUom != nil {
		qb.SetUpdate("quantity_uom_id", req.GetQuantityUom().GetById())
	}
	if req.QuantityDecimals != nil {
		qb.SetUpdate("quantity_decimals", req.GetQuantityDecimals())
	}
	if req.PriceUom != nil {
		qb.SetUpdate("price_uom_id", req.GetPriceUom().GetById())
	}
	if req.PriceDecimals != nil {
		qb.SetUpdate("price_decimals", req.GetPriceDecimals())
	}
	if req.VolumeDecimals != nil {
		qb.SetUpdate("volume_decimals", req.GetVolumeDecimals())
	}
	if req.Status != nil {
		qb.SetUpdate("status", req.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	selectSQL, selectArgs := GetFBBySelect(req.Select).GenerateSQL()

	qb.Where(selectSQL, selectArgs...)

	return qb, nil
}

func (s *TradingPairRepository) QbGetOne(req *pbTradingPairs.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_banksFields, _tableName))

	selectSQL, selectArgs := GetFBBySelect(req.Select).GenerateSQL()

	qb.Where(selectSQL, selectArgs...)

	qb.Where("tradingpairs.status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func (s *TradingPairRepository) QbGetBySelect(selectList *pbTradingPairs.SelectList) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_banksFields, _tableName))

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

func (s *TradingPairRepository) QbGetList(req *pbTradingPairs.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(_banksFields, _tableName))

	if req.Symbol != nil {
		qb.Where("tradingpairs.symbol LIKE '%' || ? || '%'", req.GetSymbol())
	}
	if req.QuantityDecimals != nil {
		decimalFilter := setDecimals(req.GetQuantityDecimals(), "tradingpairs.quantity_decimals")
		sqlStr, args := decimalFilter.GenerateSQL()

		qb.Where(sqlStr, args...)
	}
	if req.PriceDecimals != nil {
		decimalFilter := setDecimals(req.GetPriceDecimals(), "tradingpairs.price_decimals")
		sqlStr, args := decimalFilter.GenerateSQL()

		qb.Where(sqlStr, args...)
	}
	if req.VolumeDecimals != nil {
		decimalFilter := setDecimals(req.GetVolumeDecimals(), "tradingpairs.volume_decimals")
		sqlStr, args := decimalFilter.GenerateSQL()

		qb.Where(sqlStr, args...)
	}

	if req.Status != nil {
		bankStatuses := req.GetStatus().GetList()

		if len(bankStatuses) > 0 {
			args := []any{}
			for _, v := range bankStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"tradingpairs.status IN (%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(bankStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func ScanMainEntity(row pgx.Row) (*pbTradingPairs.TradingPair, error) {
	var (
		ID               string
		symbol           string
		quantityUomID    string
		quantityDecimals uint32
		priceUomID       string
		priceDecimals    uint32
		volumeDecimals   uint32
		status           pbCommon.Status
	)

	err := row.Scan(
		&ID,
		&symbol,
		&quantityUomID,
		&quantityDecimals,
		&priceUomID,
		&priceDecimals,
		&volumeDecimals,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbTradingPairs.TradingPair{
		Id:               ID,
		Symbol:           symbol,
		QuantityUom:      &pbUoMs.UoM{Id: quantityUomID},
		QuantityDecimals: quantityDecimals,
		PriceUom:         &pbUoMs.UoM{Id: priceUomID},
		PriceDecimals:    priceDecimals,
		VolumeDecimals:   volumeDecimals,
		Status:           status,
	}, nil
}

func ScanWithRelationship(row pgx.Row) (*pbTradingPairs.TradingPair, error) {
	var (
		ID               string
		symbol           string
		quantityUomID    string
		quantityDecimals uint32
		priceUomID       string
		priceDecimals    uint32
		volumeDecimals   uint32
		status           pbCommon.Status
	)

	var (
		qUomID            string
		quantityUomType   pbUoMs.Type
		quantityUomSymbol string
		// Nullable field must use sql.* type or a scan error will be thrown
		quantityUomName              sql.NullString
		quantityUomIcon              sql.NullString
		quantityUomManagedDecimals   uint32
		quantityUomDisplayedDecimals uint32
		quantityUomReportingUnit     bool
		quantityUomStatus            pbCommon.Status
	)

	var (
		pUomID         string
		priceUomType   pbUoMs.Type
		priceUomSymbol string
		// Nullable field must use sql.* type or a scan error will be thrown
		priceUomName              sql.NullString
		priceUomIcon              sql.NullString
		priceUomManagedDecimals   uint32
		priceUomDisplayedDecimals uint32
		priceUomReportingUnit     bool
		priceUomStatus            pbCommon.Status
	)

	err := row.Scan(
		&ID,
		&symbol,
		&quantityUomID,
		&quantityDecimals,
		&priceUomID,
		&priceDecimals,
		&volumeDecimals,
		&status,
		&qUomID,
		&quantityUomType,
		&quantityUomSymbol,
		&quantityUomName,
		&quantityUomIcon,
		&quantityUomManagedDecimals,
		&quantityUomDisplayedDecimals,
		&quantityUomReportingUnit,
		&quantityUomStatus,
		&pUomID,
		&priceUomType,
		&priceUomSymbol,
		&priceUomName,
		&priceUomIcon,
		&priceUomManagedDecimals,
		&priceUomDisplayedDecimals,
		&priceUomReportingUnit,
		&priceUomStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbTradingPairs.TradingPair{
		Id:     ID,
		Symbol: symbol,
		QuantityUom: &pbUoMs.UoM{
			Id:                quantityUomID,
			Type:              quantityUomType,
			Symbol:            quantityUomSymbol,
			Name:              util.GetSQLNullString(quantityUomName),
			Icon:              util.GetSQLNullString(quantityUomIcon),
			ManagedDecimals:   quantityUomManagedDecimals,
			DisplayedDecimals: quantityUomDisplayedDecimals,
			ReportingUnit:     quantityUomReportingUnit,
			Status:            quantityUomStatus,
		},
		QuantityDecimals: quantityDecimals,
		PriceUom: &pbUoMs.UoM{
			Id:                priceUomID,
			Type:              priceUomType,
			Symbol:            priceUomSymbol,
			Name:              util.GetSQLNullString(priceUomName),
			Icon:              util.GetSQLNullString(priceUomIcon),
			ManagedDecimals:   priceUomManagedDecimals,
			DisplayedDecimals: priceUomDisplayedDecimals,
			ReportingUnit:     priceUomReportingUnit,
			Status:            priceUomStatus,
		},
		PriceDecimals:  priceDecimals,
		VolumeDecimals: volumeDecimals,
		Status:         status,
	}, nil
}

func setDecimals(
	list *pbCommon.UInt32ValueList,
	field string,
) *util.FilterBracket {
	filterBracket := util.CreateFilterBracket("OR")

	for _, v := range list.GetList() {
		switch v.GetSelect().(type) {
		case *pbCommon.UInt32Values_Single:
			filterBracket.SetFilter(
				fmt.Sprintf("%s = ?", field),
				v.GetSelect().(*pbCommon.UInt32Values_Single).Single,
			)
		case *pbCommon.UInt32Values_Range:
			from := v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From
			to := v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To
			if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From != nil && v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To != nil {
				setRangeCondition("<", field, from, filterBracket)
				setRangeCondition(">", field, to, filterBracket)
			} else if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.From != nil {
				setRangeCondition(">", field, to, filterBracket)
			} else if v.GetSelect().(*pbCommon.UInt32Values_Range).Range.To != nil {
				setRangeCondition("<", field, from, filterBracket)
			}
		}
	}
	return filterBracket
}

func setRangeCondition(
	expression, field string,
	value *pbCommon.UInt32Boundary,
	filterBracket *util.FilterBracket,
) {
	switch value.GetBoundary().(type) {
	case *pbCommon.UInt32Boundary_Incl:
		filterBracket.SetFilter(
			fmt.Sprintf("%s %s= ?", field, expression),
			value.GetBoundary().(*pbCommon.UInt32Boundary_Incl).Incl,
		)
	case *pbCommon.UInt32Boundary_Excl:
		filterBracket.SetFilter(
			fmt.Sprintf("%s %s ?", field, expression),
			value.GetBoundary().(*pbCommon.UInt32Boundary_Excl).Excl,
		)
	}
}
