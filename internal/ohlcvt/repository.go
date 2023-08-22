package ohlcvt

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbDataSources "davensi.com/core/gen/datasources"
	pbFSProvider "davensi.com/core/gen/fsproviders"
	pbMarkets "davensi.com/core/gen/markets"
	pbOhlcvt "davensi.com/core/gen/ohlcvt"
	pbOhlcvtConnect "davensi.com/core/gen/ohlcvt/ohlcvtconnect"
	pbTradingpairs "davensi.com/core/gen/tradingpairs"

	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OhlcvtRepository struct {
	pbOhlcvtConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func NewOhlcvtRepository(db *pgxpool.Pool) *OhlcvtRepository {
	return &OhlcvtRepository{
		db: db,
	}
}

func (s *OhlcvtRepository) QbInsert(msg *pbOhlcvt.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singleOhlcvtValue := []any{}

	qb.SetInsertField("source_id").
		SetInsertField("market_id").
		SetInsertField("price_type").
		SetInsertField("timestamp").
		SetInsertField("open").
		SetInsertField("high").
		SetInsertField("low").
		SetInsertField("close").
		SetInsertField("volume_in_quantity_uom").
		SetInsertField("volume_in_price_uom").
		SetInsertField("trades").
		SetInsertField("status")

	singleOhlcvtValue = append(singleOhlcvtValue,
		msg.GetSource().GetById(),
		msg.GetMarket().GetById(),
		msg.GetPriceType(),
		msg.GetTimestamp(),
		msg.GetOpen(),
		msg.GetHigh(),
		msg.GetLow(),
		msg.GetOpen(),
		msg.GetOpen(),
		msg.GetOpen(),
		msg.GetOpen(),
		msg.GetStatus(),
	)
	_, err := qb.SetInsertValues(singleOhlcvtValue)

	return qb, err
}

func (s *OhlcvtRepository) QbUpdate(msg *pbOhlcvt.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	handleTypicalOhlcvtUpdateFields(
		msg,
		func(field string, value any) {
			qb.Where(fmt.Sprintf("%s LIKE '%%' || ? || '%%'", field), value)
		},
	)

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	switch msg.GetSelect().Select.(type) {
	case *pbOhlcvt.Select_ById:
		qb.Where("prices.id = ?", msg.GetSelect().GetById())
	case *pbOhlcvt.Select_ByOhlcvtKey:
		qb.Where(
			"ohlcvt.source_id = ? AND ohlcvt.market_id = ? AND ohlcvt.price_type = ? AND ohlcvt.timestamp = ? ",
			msg.GetSelect().GetByOhlcvtKey().GetSource().GetId(),
			msg.GetSelect().GetByOhlcvtKey().GetMarket().GetId(),
			msg.GetSelect().GetByOhlcvtKey().GetPriceType(),
			msg.GetSelect().GetByOhlcvtKey().GetTimestamp(),
		)
	}

	return qb, nil
}

func (s *OhlcvtRepository) QbGetOne(msg *pbOhlcvt.GetRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	switch msg.GetSelect().Select.(type) {
	case *pbOhlcvt.Select_ById:
		qb.Where("ohlcvt.id = ?", msg.GetSelect().GetById())
	case *pbOhlcvt.Select_ByOhlcvtKey:
		qb.Where("ohlcvt.source_id = ? AND ohlcvt.market_id = ? AND ohlcvt.type = ? AND ohlcvt.timestamp = ? ",
			msg.GetSelect().GetByOhlcvtKey().GetSource().GetId(),
			msg.GetSelect().GetByOhlcvtKey().GetMarket().GetId(),
			msg.GetSelect().GetByOhlcvtKey().GetPriceType(),
			msg.GetSelect().GetByOhlcvtKey().GetTimestamp(),
		)
	}

	return qb
}

func (s *OhlcvtRepository) QbGetList(msg *pbOhlcvt.GetListRequest) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(_fields)

	//TODO: relationship source and market and timestamp
	handleTypicalOhlcvtFields(
		msg,
		func(field string, props any, value []any) {
			if len(value) > 0 {
				qb.Where(
					fmt.Sprintf(
						"%s = ANY(%s)",
						field,
						strings.Join(strings.Split(strings.Repeat("?", len(value)), ""), ", "),
					),
					value...,
				)
			}
		},
	)

	return qb
}

func decimalToAny(req *pbCommon.DecimalValueList) []any {
	var arr []any
	for _, v := range req.GetList() {
		arr = append(arr, v)
	}
	return arr
}

func handleTypicalOhlcvtFields(msg *pbOhlcvt.GetListRequest, handleFn func(field string, props any, value []any)) {
	if msg.Type != nil {
		var arr []any
		for _, v := range msg.GetType().GetList() {
			arr = append(arr, v)
		}
		handleFn("price_type", msg.Type, arr)
	}
	// TODO timestamps
	if msg.Open != nil {
		arr := decimalToAny(msg.GetOpen())
		handleFn("open", msg.Open, arr)
	}
	if msg.High != nil {
		arr := decimalToAny(msg.GetHigh())
		handleFn("high", msg.High, arr)
	}
	if msg.Low != nil {
		arr := decimalToAny(msg.GetLow())
		handleFn("low", msg.Low, arr)
	}
	if msg.Close != nil {
		arr := decimalToAny(msg.GetClose())
		handleFn("close", msg.Close, arr)
	}
	if msg.VolumeInQuantityUom != nil {
		arr := decimalToAny(msg.GetVolumeInQuantityUom())
		handleFn("volume_in_quantity_uom", msg.VolumeInQuantityUom, arr)
	}
	if msg.VolumeInPriceUom != nil {
		arr := decimalToAny(msg.GetVolumeInPriceUom())
		handleFn("volume_in_price_uom", msg.VolumeInPriceUom, arr)
	}
	if msg.Trades != nil {
		var arr []any
		for _, v := range msg.GetTrades().GetList() {
			arr = append(arr, v)
		}
		handleFn("trades", msg.Trades, arr)
	}
	if msg.Status != nil {
		var arr []any
		for _, v := range msg.GetStatus().GetList() {
			arr = append(arr, v)
		}
		handleFn("ohlcvt.status", msg.Status, arr)
	}
}

func handleTypicalOhlcvtUpdateFields(msg *pbOhlcvt.UpdateRequest, handleFn func(field string, value any)) {
	if msg.PriceType != nil {
		handleFn("price_type", msg.GetPriceType())
	}
	if msg.Timestamp != nil {
		handleFn("timestamp", msg.GetTimestamp())
	}
	if msg.Open != nil {
		handleFn("open", msg.GetOpen())
	}
	if msg.High != nil {
		handleFn("high", msg.GetHigh())
	}
	if msg.Low != nil {
		handleFn("low", msg.GetLow())
	}
	if msg.Close != nil {
		handleFn("close", msg.GetClose())
	}
	if msg.VolumeInQuantityUom != nil {
		handleFn("volume_in_quantity_uom", msg.GetVolumeInQuantityUom())
	}
	if msg.VolumeInPriceUom != nil {
		handleFn("volume_in_price_uom", msg.GetVolumeInPriceUom())
	}
	if msg.Trades != nil {
		handleFn("trades", msg.GetTrades())
	}
	if msg.Status != nil {
		handleFn("status", msg.GetStatus())
	}
}

func (s *OhlcvtRepository) ScanMainEntity(row pgx.Row) (*pbOhlcvt.OHLCVT, error) {
	var (
		id                  string
		sourceID            string
		marketID            string
		priceType           pbMarkets.PriceType
		timestamp           timestamppb.Timestamp
		open                pbCommon.Decimal
		high                pbCommon.Decimal
		low                 pbCommon.Decimal
		ohlcvtClose         pbCommon.Decimal
		volumeInQuantityUom pbCommon.Decimal
		volumeInPriceUom    pbCommon.Decimal
		trades              uint32
		status              pbCommon.Status
	)

	err := row.Scan(
		&id,
		&sourceID,
		&marketID,
		&priceType,
		&timestamp,
		&open,
		&high,
		&low,
		&ohlcvtClose,
		&volumeInQuantityUom,
		&volumeInPriceUom,
		&trades,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbOhlcvt.OHLCVT{
		Id:                  id,
		Source:              &pbDataSources.DataSource{Id: sourceID},
		Market:              &pbMarkets.Market{Id: marketID},
		PriceType:           priceType,
		Timestamp:           &timestamp,
		Open:                &open,
		High:                &high,
		Low:                 &low,
		Close:               &ohlcvtClose,
		VolumeInQuantityUom: &volumeInQuantityUom,
		VolumeInPriceUom:    &volumeInPriceUom,
		Trades:              trades,
		Status:              status,
	}, nil
}

func (s *OhlcvtRepository) ScanWithRelationship(row pgx.Row) (*pbOhlcvt.OHLCVT, error) {
	var (
		dsID           string
		dsType         pbDataSources.Type
		dsName         string
		dsFsProviderID string
		dsIcon         sql.NullString
		dsStatus       pbCommon.Status
	)
	var (
		marketID            string
		marketSymbol        string
		marketType          pbMarkets.Type
		marketTradingPairID string
		marketAlgorithm     pbMarkets.MatchingAlgorithm
		marketPriceType     pbMarkets.PriceType
		marketTickSize      sql.NullFloat64
		marketState         pbMarkets.State
		marketStatus        pbCommon.Status
	)
	var (
		id                  string
		priceType           pbMarkets.PriceType
		timestamp           sql.NullTime
		open                sql.NullFloat64
		high                sql.NullFloat64
		low                 sql.NullFloat64
		ohlcvtClose         sql.NullFloat64
		volumeInQuantityUom sql.NullFloat64
		volumeInPriceUom    sql.NullFloat64
		trades              uint32
		status              pbCommon.Status
	)

	err := row.Scan(
		&id,
		&dsID,
		&marketID,
		&priceType,
		&timestamp,
		&open,
		&high,
		&low,
		&ohlcvtClose,
		&volumeInQuantityUom,
		&volumeInPriceUom,
		&trades,
		&status,
		&dsID,
		&dsType,
		&dsName,
		&dsFsProviderID,
		&dsIcon,
		&dsStatus,
		&marketID,
		&marketSymbol,
		&marketTickSize,
		&marketStatus,
		&marketType,
		&marketTradingPairID,
		&marketAlgorithm,
		&marketPriceType,
		&marketState,
	)
	if err != nil {
		return nil, err
	}

	return &pbOhlcvt.OHLCVT{
		Id: dsID,
		Source: &pbDataSources.DataSource{
			Id:   dsID,
			Type: dsType,
			Name: dsName,
			Provider: &pbFSProvider.FSProvider{
				Id: dsFsProviderID,
			},
			Icon:   util.GetSQLNullString(dsIcon),
			Status: dsStatus,
		},
		Market: &pbMarkets.Market{
			Id:     marketID,
			Symbol: marketSymbol,
			Type:   marketType,
			Tradingpair: &pbTradingpairs.TradingPair{
				Id: marketTradingPairID,
			},
			Algorithm: marketAlgorithm,
			PriceType: marketPriceType,
			TickSize:  util.FloatToDecimal(util.GetSQLNullFloat(marketTickSize)),
			State:     marketState,
			Status:    marketStatus,
		},
		PriceType:           priceType,
		Timestamp:           util.GetSQLNullTime(timestamp),
		Open:                util.FloatToDecimal(util.GetSQLNullFloat(open)),
		High:                util.FloatToDecimal(util.GetSQLNullFloat(high)),
		Low:                 util.FloatToDecimal(util.GetSQLNullFloat(low)),
		Close:               util.FloatToDecimal(util.GetSQLNullFloat(ohlcvtClose)),
		VolumeInQuantityUom: util.FloatToDecimal(util.GetSQLNullFloat(volumeInQuantityUom)),
		VolumeInPriceUom:    util.FloatToDecimal(util.GetSQLNullFloat(volumeInPriceUom)),
		Trades:              trades,
		Status:              status,
	}, nil
}
