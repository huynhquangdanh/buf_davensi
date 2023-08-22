package prices

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	pbCommon "davensi.com/core/gen/common"
	pbDataSources "davensi.com/core/gen/datasources"
	pbFsproviders "davensi.com/core/gen/fsproviders"
	pbMarkets "davensi.com/core/gen/markets"
	pbPrices "davensi.com/core/gen/prices"
	pbPricesConnect "davensi.com/core/gen/prices/pricesconnect"
	pbTradingpairs "davensi.com/core/gen/tradingpairs"

	"davensi.com/core/internal/datasources"
	"davensi.com/core/internal/markets"
	"davensi.com/core/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PriceRepository struct {
	pbPricesConnect.UnimplementedServiceHandler
	db *pgxpool.Pool
}

func SetQBBySelect(selectPrice *pbPrices.Select, qb *util.QueryBuilder) {
	switch selectPrice.GetSelect().(type) {
	case *pbPrices.Select_ById:
		qb.Where("prices.id = ?", selectPrice.GetById())
	case *pbPrices.Select_ByPriceKey:
		filterArgs := []any{}

		qbGetSource := util.CreateQueryBuilder(util.Select, "core.datasources")
		datasources.SetQBBySelect(
			selectPrice.GetByPriceKey().GetSource(),
			qbGetSource.Select("id"),
		)
		sourceSQL, sourceArgs, _ := qbGetSource.GenerateSQL()

		qbGetMarkets := util.CreateQueryBuilder(util.Select, "core.markets")
		markets.SetQBBySelect(
			selectPrice.GetByPriceKey().GetMarket(),
			qbGetMarkets.Select("id"),
		)
		marketSQL, marketArgs, _ := qbGetMarkets.GenerateSQL()

		filterArgs = append(filterArgs, sourceArgs...)
		filterArgs = append(filterArgs, marketArgs...)
		filterArgs = append(
			filterArgs,
			selectPrice.GetByPriceKey().GetType(),
			util.GetDBTimestampValue(selectPrice.GetByPriceKey().GetTimestamp()),
		)

		qb.Where(
			fmt.Sprintf(
				"prices.source_id IN (%s) AND prices.market_id IN (%s) AND prices.type = ? AND prices.timestamp = ? ",
				sourceSQL,
				marketSQL,
			),
			filterArgs...,
		)
	}
}

func NewPriceRepository(db *pgxpool.Pool) *PriceRepository {
	return &PriceRepository{
		db: db,
	}
}

func (s *PriceRepository) QbInsert(msg *pbPrices.CreateRequest) (*util.QueryBuilder, error) {
	qb := util.CreateQueryBuilder(util.Insert, _tableName)
	singlePriceValue := []any{}

	qb.SetInsertField("source_id").
		SetInsertField("market_id").
		SetInsertField("type").
		SetInsertField("timestamp").
		SetInsertField("price").
		SetInsertField("status")

	singlePriceValue = append(singlePriceValue,
		msg.GetSource().GetById(),
		msg.GetMarket().GetById(),
		msg.GetType(),
		util.GetDBTimestampValue(msg.GetTimestamp()),
		msg.GetPrice().GetValue(),
		msg.GetStatus(),
	)
	_, err := qb.SetInsertValues(singlePriceValue)

	return qb, err
}

func (s *PriceRepository) QbUpdate(msg *pbPrices.UpdateRequest) (qb *util.QueryBuilder, err error) {
	qb = util.CreateQueryBuilder(util.Update, _tableName)

	if msg.Source != nil {
		qb.SetUpdate("source_id", msg.GetSource().GetById())
	}
	if msg.Market != nil {
		qb.SetUpdate("market_id", msg.GetMarket().GetById())
	}
	if msg.Type != nil {
		qb.SetUpdate("type", msg.GetType())
	}
	if msg.Timestamp != nil {
		qb.SetUpdate("timestamp", util.GetDBTimestampValue(msg.GetTimestamp()))
	}
	if msg.Price != nil {
		qb.SetUpdate("price", msg.GetPrice().GetValue())
	}
	if msg.Status != nil {
		qb.SetUpdate("status", msg.GetStatus())
	}

	if !qb.IsUpdatable() {
		return qb, errors.New("cannot update without new value")
	}

	SetQBBySelect(msg.GetSelect(), qb)

	return qb, nil
}

func (s *PriceRepository) QbGetOne(msg *pbPrices.GetRequest) *util.QueryBuilder {
	qb := util.
		CreateQueryBuilder(util.Select, _tableName).
		Select(util.GetFieldsWithTableName(PriceFields, _tableName)).
		Where("prices.status = ?", pbCommon.Status_STATUS_ACTIVE)

	SetQBBySelect(msg.GetSelect(), qb)

	return qb
}

func (s *PriceRepository) QbGetList(
	msg *pbPrices.GetListRequest,
) *util.QueryBuilder {
	qb := util.CreateQueryBuilder(util.Select, _tableName)
	qb.Select(util.GetFieldsWithTableName(PriceFields, "prices"))

	// TODO: relationship source and market and timestamp
	if msg.Type != nil {
		priceTypes := msg.GetType().GetList()

		if len(priceTypes) > 0 {
			args := []any{}
			for _, v := range priceTypes {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"prices.type IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(priceTypes)), ""), ", "),
				),
				args...,
			)
		}
	}
	if msg.Status != nil {
		socialStatuses := msg.GetStatus().GetList()

		if len(socialStatuses) > 0 {
			args := []any{}
			for _, v := range socialStatuses {
				args = append(args, v)
			}

			qb.Where(
				fmt.Sprintf(
					"prices.status IN(%s)",
					strings.Join(strings.Split(strings.Repeat("?", len(socialStatuses)), ""), ", "),
				),
				args...,
			)
		}
	}

	return qb
}

func (s *PriceRepository) ScanWithRelationship(row pgx.Row) (*pbPrices.Price, error) {
	var (
		priceID        string
		priceSourceID  string
		priceMarketID  string
		priceType      pbMarkets.PriceType
		priceTimestamp sql.NullTime
		price          sql.NullFloat64
		priceStatus    pbCommon.Status
	)
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
		dsID           string
		dsType         pbDataSources.Type
		dsName         string
		dsFsProviderID sql.NullString
		dsIcon         sql.NullString
		dsStatus       pbCommon.Status
	)

	err := row.Scan(
		&priceID,
		&priceSourceID,
		&priceMarketID,
		&priceType,
		&priceTimestamp,
		&price,
		&priceStatus,
		&marketID,
		&marketSymbol,
		&marketTickSize,
		&marketStatus,
		&marketType,
		&marketTradingpairID,
		&marketAlgorithm,
		&marketPriceType,
		&marketState,
		&dsID,
		&dsType,
		&dsName,
		&dsFsProviderID,
		&dsIcon,
		&dsStatus,
	)
	if err != nil {
		return nil, err
	}

	return &pbPrices.Price{
		Id: priceID,
		Source: &pbDataSources.DataSource{
			Id:       dsID,
			Type:     dsType,
			Name:     dsName,
			Provider: &pbFsproviders.FSProvider{Id: util.GetPointString(util.GetSQLNullString(dsFsProviderID))},
			Icon:     util.GetSQLNullString(dsIcon),
			Status:   dsStatus,
		},
		Market: &pbMarkets.Market{
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
		},
		Type:      priceType,
		Timestamp: util.GetSQLNullTime(priceTimestamp),
		Price:     util.FloatToDecimal(util.GetSQLNullFloat(price)),
		Status:    priceStatus,
	}, nil
}

func (s *PriceRepository) ScanMainEntity(row pgx.Row) (*pbPrices.Price, error) {
	var (
		id        string
		sourceID  string
		marketID  string
		priceType pbMarkets.PriceType
		timestamp sql.NullTime
		price     sql.NullFloat64
		status    pbCommon.Status
	)

	err := row.Scan(
		&id,
		&sourceID,
		&marketID,
		&priceType,
		&timestamp,
		&price,
		&status,
	)
	if err != nil {
		return nil, err
	}

	return &pbPrices.Price{
		Id:        id,
		Source:    &pbDataSources.DataSource{Id: sourceID},
		Market:    &pbMarkets.Market{Id: marketID},
		Type:      priceType,
		Timestamp: util.GetSQLNullTime(timestamp),
		Price:     util.FloatToDecimal(util.GetSQLNullFloat(price)),
		Status:    status,
	}, nil
}
