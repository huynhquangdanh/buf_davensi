package markets

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbMarkets "davensi.com/core/gen/markets"
	pbMarketsConnect "davensi.com/core/gen/markets/marketsconnect"
	pbTradingpairs "davensi.com/core/gen/tradingpairs"
	"davensi.com/core/gen/uoms"
	"davensi.com/core/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_tableName   = "core.markets"
	MarketFields = "id, symbol, tick_size, status, type, tradingpair_id, algorithm, price_type, state"
)

type MarketRepository struct {
	pbMarketsConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewMarketRepository(db *pgxpool.Pool) *MarketRepository {
	return &MarketRepository{
		db: db,
	}
}

func SetQBBySelect(selectMarket *pbMarkets.Select, qb *util.QueryBuilder) {
	switch selectMarket.GetSelect().(type) {
	case *pbMarkets.Select_ById:
		qb.Where("markets.id = ?", selectMarket.GetById())
	case *pbMarkets.Select_BySymbol:
		qb.Where("markets.symbol = ?", selectMarket.GetBySymbol())
	}
}

func (s *MarketRepository) QbInsert(msg *pbMarkets.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleValue := []any{}

	// Append required value: name
	qb.SetInsertField(
		"symbol", "type", "tradingpair_id", "algorithm", "price_type", "state", "status",
	)
	singleValue = append(
		singleValue,
		msg.GetSymbol(),
		msg.GetType(),
		// must validate and replace with existed tradingpair
		msg.GetTradingpair().GetById(),
		msg.GetAlgorithm(),
		msg.GetPriceType(),
		msg.GetState(),
		msg.GetStatus(),
	)

	// Append optional fields values
	if msg.TickSize != nil {
		qb.SetInsertField("tick_size")
		singleValue = append(singleValue, msg.GetTickSize().GetValue())
	}

	_, err := qb.SetInsertValues(singleValue)

	return qb, err
}

func (s *MarketRepository) QbUpdate(msg *pbMarkets.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Symbol != nil {
		qb.SetUpdate("symbol", msg.GetSymbol())
	}

	if msg.Type != nil {
		qb.SetUpdate("type", msg.GetType())
	}

	// must validate and replace with existed tradingpair
	if msg.Tradingpair != nil {
		qb.SetUpdate("tradingpair_id", msg.GetTradingpair().GetById())
	}

	if msg.Algorithm != nil {
		qb.SetUpdate("algorithm", msg.GetAlgorithm())
	}

	if msg.PriceType != nil {
		qb.SetUpdate("price_type", msg.GetPriceType())
	}

	if msg.TickSize != nil {
		qb.SetUpdate("tick_size", msg.GetTickSize().GetValue())
	}

	if msg.State != nil {
		qb.SetUpdate("state", msg.GetState())
	}

	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	SetQBBySelect(msg.Select, qb)

	return qb, nil
}

func (s *MarketRepository) QbGetOne(msg *pbMarkets.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(MarketFields, _tableName))

	SetQBBySelect(msg.Select, qb)

	qb.Where("markets.status = ?", pbCommon.Status_STATUS_ACTIVE)

	return qb
}

func createArgs[T any](slice []T) []any {
	args := []any{}
	for _, v := range slice {
		args = append(args, v)
	}
	return args
}

func appendListFilterQB(msg *pbMarkets.GetListRequest, fn func(fields string, args []any)) {
	if msg.Type != nil && len(msg.GetType().GetList()) > 0 {
		fn("markets.type", createArgs[pbMarkets.Type](msg.GetType().GetList()))
	}
	if msg.Algorithm != nil && len(msg.GetAlgorithm().GetList()) > 0 {
		fn("markets.algorithm", createArgs[pbMarkets.MatchingAlgorithm](msg.GetAlgorithm().GetList()))
	}
	if msg.PriceType != nil && len(msg.GetPriceType().GetList()) > 0 {
		fn("markets.price_type", createArgs[pbMarkets.PriceType](msg.GetPriceType().GetList()))
	}
	if msg.State != nil && len(msg.GetState().GetList()) > 0 {
		fn("markets.state", createArgs[pbMarkets.State](msg.GetState().GetList()))
	}
	if msg.Status != nil && len(msg.GetStatus().GetList()) > 0 {
		fn("markets.status", createArgs[pbCommon.Status](msg.GetStatus().GetList()))
	}
}

func (s *MarketRepository) QbGetList(msg *pbMarkets.GetListRequest) *util.QueryBuilder {
	qb := util.
		CreateQueryBuilder(util.Select, _tableName).
		Select(util.GetFieldsWithTableName(MarketFields, "markets"))

	if msg.Symbol != nil {
		qb.Where("markets.symbol LIKE '%' || ? || '%'", msg.GetSymbol())
	}

	appendListFilterQB(msg, func(field string, args []any) {
		qb.Where(
			fmt.Sprintf(
				"%s IN(%s)",
				field,
				strings.Join(strings.Split(strings.Repeat("?", len(args)), ""), ", "),
			),
			args...,
		)
	})

	if msg.TickSize != nil {
		ticketSizeFB := GetDecimalsFB(msg.GetTickSize(), "markets.tick_size")
		ticketSizeBracket, ticketSizeArgs := ticketSizeFB.GenerateSQL()
		qb.Where(ticketSizeBracket, ticketSizeArgs...)
	}

	return qb
}

func setRangeCondition(
	expression, field string,
	value *pbCommon.DecimalBoundary,
	filterBracket *util.FilterBracket,
) {
	switch value.GetBoundary().(type) {
	case *pbCommon.DecimalBoundary_Incl:
		filterBracket.SetFilter(
			fmt.Sprintf("%s %s= ?", field, expression),
			value.GetBoundary().(*pbCommon.DecimalBoundary_Incl).Incl,
		)
	case *pbCommon.DecimalBoundary_Excl:
		filterBracket.SetFilter(
			fmt.Sprintf("%s %s ?", field, expression),
			value.GetBoundary().(*pbCommon.DecimalBoundary_Excl).Excl,
		)
	}
}

func GetDecimalsFB(
	list *pbCommon.DecimalValueList,
	field string,
) *util.FilterBracket {
	filterBracket := util.CreateFilterBracket("OR")

	for _, v := range list.GetList() {
		switch v.GetSelect().(type) {
		case *pbCommon.DecimalValues_Single:
			filterBracket.SetFilter(
				fmt.Sprintf("%s = ?", field),
				v.GetSelect().(*pbCommon.DecimalValues_Single).Single,
			)
		case *pbCommon.DecimalValues_Range:
			from := v.GetSelect().(*pbCommon.DecimalValues_Range).Range.From
			to := v.GetSelect().(*pbCommon.DecimalValues_Range).Range.To
			if v.GetSelect().(*pbCommon.DecimalValues_Range).Range.From != nil && v.GetSelect().(*pbCommon.DecimalValues_Range).Range.To != nil {
				setRangeCondition("<", field, from, filterBracket)
				setRangeCondition(">", field, to, filterBracket)
			} else if v.GetSelect().(*pbCommon.DecimalValues_Range).Range.From != nil {
				setRangeCondition(">", field, to, filterBracket)
			} else if v.GetSelect().(*pbCommon.DecimalValues_Range).Range.To != nil {
				setRangeCondition("<", field, from, filterBracket)
			}
		}
	}

	return filterBracket
}

func (s *MarketRepository) ScanMainEntity(row pgx.Row) (*pbMarkets.Market, error) {
	var (
		marketID            string
		marketSymbol        string
		marketTickSize      sql.NullFloat64
		marketStatus        pbCommon.Status
		marketType          pbMarkets.Type
		marketTradingpairID string
		marketAlgorithm     pbMarkets.MatchingAlgorithm
		marketPriceType     pbMarkets.PriceType
		marketState         pbMarkets.State
	)

	err := row.Scan(
		&marketID,
		&marketSymbol,
		&marketTickSize,
		&marketStatus,
		&marketType,
		&marketTradingpairID,
		&marketAlgorithm,
		&marketPriceType,
		&marketState,
	)
	if err != nil {
		return nil, err
	}

	return &pbMarkets.Market{
		Id:     marketID,
		Symbol: marketSymbol,
		Type:   marketType,
		Tradingpair: &pbTradingpairs.TradingPair{
			Id:               marketTradingpairID,
			Symbol:           "",
			QuantityUom:      nil,
			QuantityDecimals: 0,
			PriceUom:         nil,
			PriceDecimals:    0,
			VolumeDecimals:   0,
			Status:           0,
		},
		Algorithm: marketAlgorithm,
		PriceType: marketPriceType,
		TickSize:  util.FloatToDecimal(util.GetSQLNullFloat(marketTickSize)),
		State:     marketState,
		Status:    marketStatus,
	}, nil
}

func (s *MarketRepository) ScanWithRelationship(row pgx.Row) (*pbMarkets.Market, error) {
	var (
		marketID            string
		marketSymbol        string
		marketTickSize      sql.NullFloat64
		marketStatus        pbCommon.Status
		marketType          pbMarkets.Type
		marketTradingpairID string
		marketAlgorithm     pbMarkets.MatchingAlgorithm
		marketPriceType     pbMarkets.PriceType
		marketState         pbMarkets.State
	)

	var (
		tradingPairID             string
		tradingPairSymbol         string
		tradingPairQuantityUomID  string
		tradingQuantityDecimals   uint32
		tradingPairPriceUomID     string
		tradingPairPriceDecimals  uint32
		tradingPairVolumeDecimals uint32
		tradingPairStatus         pbCommon.Status
	)

	err := row.Scan(
		&marketID,
		&marketSymbol,
		&marketTickSize,
		&marketStatus,
		&marketType,
		&marketTradingpairID,
		&marketAlgorithm,
		&marketPriceType,
		&marketState,
		&tradingPairID,
		&tradingPairSymbol,
		&tradingPairQuantityUomID,
		&tradingQuantityDecimals,
		&tradingPairPriceUomID,
		&tradingPairPriceDecimals,
		&tradingPairVolumeDecimals,
		&tradingPairStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbMarkets.Market{
		Id:     marketID,
		Symbol: marketSymbol,
		Type:   marketType,
		Tradingpair: &pbTradingpairs.TradingPair{
			Id:     tradingPairID,
			Symbol: tradingPairSymbol,
			QuantityUom: &uoms.UoM{
				Id: tradingPairQuantityUomID,
			},
			QuantityDecimals: tradingQuantityDecimals,
			PriceUom: &uoms.UoM{
				Id: tradingPairPriceUomID,
			},
			PriceDecimals:  tradingPairPriceDecimals,
			VolumeDecimals: tradingPairVolumeDecimals,
			Status:         tradingPairStatus,
		},
		Algorithm: marketAlgorithm,
		PriceType: marketPriceType,
		TickSize:  util.FloatToDecimal(util.GetSQLNullFloat(marketTickSize)),
		State:     marketState,
		Status:    marketStatus,
	}, nil
}
