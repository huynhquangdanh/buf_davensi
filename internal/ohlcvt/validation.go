package ohlcvt

import (
	"context"

	"github.com/bufbuild/connect-go"

	pbCommon "davensi.com/core/gen/common"
	pbDataSources "davensi.com/core/gen/datasources"
	pbMarkets "davensi.com/core/gen/markets"
	pbOhlcvt "davensi.com/core/gen/ohlcvt"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/datasources"
	"davensi.com/core/internal/markets"
)

func (s *ServiceServer) IsOhlcvtUniq(ohlcvtKey *pbOhlcvt.Select_ByOhlcvtKey) (isUniq bool, errno pbCommon.ErrorCode) {
	ohlvct, err := s.Get(context.Background(), &connect.Request[pbOhlcvt.GetRequest]{
		Msg: &pbOhlcvt.GetRequest{
			Select: &pbOhlcvt.Select{
				Select: &pbOhlcvt.Select_ByOhlcvtKey{
					ByOhlcvtKey: ohlcvtKey.ByOhlcvtKey,
				},
			},
		},
	})

	if err == nil || ohlvct.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, ohlvct.Msg.GetError().Code
	}

	if ohlvct.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, ohlvct.Msg.GetError().Code
	}

	return true, ohlvct.Msg.GetError().Code
}

// for Create gRPC
func (s *ServiceServer) validateCreate(msg *pbOhlcvt.CreateRequest) *common.ErrWithCode {
	// Verify that Human Keys are specified
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if msg.Source.GetById() == "" {
		return errCreation.UpdateMessage("human keys must be specified")
	}
	if msg.Market.GetById() == "" {
		return errCreation.UpdateMessage("human keys must be specified")
	}
	if msg.PriceType == 0 {
		return errCreation.UpdateMessage("human keys must be specified")
	}
	if msg.Timestamp == nil {
		return errCreation.UpdateMessage("human keys must be specified")
	}

	if err := datasources.ValidateSelect(msg.Source, "creating"); err != nil {
		return err
	}

	ohlcvtRl := s.GetRelationship(msg.GetSource(), msg.GetMarket())

	if ohlcvtRl.dataSource == nil {
		return errCreation.UpdateMessage("data source does not exist")
	}

	msg.Source = &pbDataSources.Select{
		Select: &pbDataSources.Select_ById{
			ById: ohlcvtRl.dataSource.Id,
		},
	}

	if err := markets.ValidateSelect(msg.Market, "creating"); err != nil {
		return err
	}

	if ohlcvtRl.market == nil {
		return errCreation.UpdateMessage("market does not exist")
	}

	msg.Market = &pbMarkets.Select{
		Select: &pbMarkets.Select_ById{
			ById: ohlcvtRl.market.Id,
		},
	}

	if isUniq, errCode := s.IsOhlcvtUniq(&pbOhlcvt.Select_ByOhlcvtKey{
		ByOhlcvtKey: &pbOhlcvt.OHLCVTKey{
			Source:    &pbDataSources.DataSource{Id: msg.GetSource().GetById()},
			Market:    &pbMarkets.Market{Id: msg.GetMarket().GetById()},
			PriceType: msg.GetPriceType(),
			Timestamp: msg.GetTimestamp(),
		},
	}); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("type and name have been used")
	}

	return nil
}

// for Update gRPC
func (s *ServiceServer) validateQueryUpdate(msg *pbOhlcvt.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)

	ohlcvtRl := s.GetRelationship(msg.GetSource(), msg.GetMarket())
	if msg.Source != nil && ohlcvtRl.dataSource == nil {
		return errUpdate.UpdateMessage("data source does not exist")
	}
	if msg.Market != nil && ohlcvtRl.market == nil {
		return errUpdate.UpdateMessage("market does not exist")
	}

	if ohlcvtRl.dataSource != nil {
		msg.Source = &pbDataSources.Select{
			Select: &pbDataSources.Select_ById{
				ById: ohlcvtRl.dataSource.Id,
			},
		}
	}
	if ohlcvtRl.market != nil {
		msg.Market = &pbMarkets.Select{
			Select: &pbMarkets.Select_ById{
				ById: ohlcvtRl.market.Id,
			},
		}
	}

	switch msg.GetSelect().Select.(type) {
	case *pbOhlcvt.Select_ById:
		if msg.GetSelect().GetById() == "" {
			return errUpdate.UpdateMessage("id must be specified")
		}
	case *pbOhlcvt.Select_ByOhlcvtKey:
		if msg.GetSelect().GetByOhlcvtKey() == nil {
			return errUpdate.UpdateMessage("ohlvct key must be specified")
		}
		if msg.GetSelect().GetByOhlcvtKey().GetSource() == nil ||
			msg.GetSelect().GetByOhlcvtKey().GetPriceType() == pbMarkets.PriceType_PRICE_TYPE_UNSPECIFIED ||
			msg.GetSelect().GetByOhlcvtKey().GetMarket() == nil ||
			msg.GetSelect().GetByOhlcvtKey().GetTimestamp() == nil {
			return errUpdate.UpdateMessage("get '%s' type and symbol must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) ValidateSelect(msg *pbOhlcvt.Select, method string) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_package,
		"",
	)
	if msg == nil {
		return errUpdate.UpdateMessage("by_id or by_type_name must be specified")
	}
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_type_name must be specified")
	}
	switch msg.GetSelect().(type) {
	case *pbOhlcvt.Select_ByOhlcvtKey:
		if msg.GetByOhlcvtKey() == nil {
			return errUpdate.UpdateMessage("OHLCVT key must be specified")
		}
		if msg.GetByOhlcvtKey().Source == nil ||
			msg.GetByOhlcvtKey().Market == nil ||
			msg.GetByOhlcvtKey().PriceType == 0 ||
			msg.GetByOhlcvtKey().Timestamp == nil {
			return errUpdate.UpdateMessage("fields of OHLCVT key must be specified")
		}
	case *pbOhlcvt.Select_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errUpdate.UpdateMessage("by_id must be specified")
		}
	}

	return nil
}

func validateSelect(msg *pbOhlcvt.GetRequest) *common.ErrWithCode {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"get",
		_package,
		"",
	)
	if msg == nil {
		return errGet.UpdateMessage("by_id or by_type_name must be specified")
	}
	if msg.Select == nil {
		return errGet.UpdateMessage("by_id or by_type_name must be specified")
	}
	switch msg.GetSelect().Select.(type) {
	case *pbOhlcvt.Select_ById:
		// Verify that ID is specified
		if msg.GetSelect().GetById() == "" {
			return errGet.UpdateMessage("by_id must be specified")
		}
	case *pbOhlcvt.Select_ByOhlcvtKey:
		if msg.GetSelect().GetByOhlcvtKey() == nil {
			return errGet.UpdateMessage("by_type_name must be specified")
		}
		if msg.GetSelect().GetByOhlcvtKey().Timestamp == nil ||
			msg.GetSelect().GetByOhlcvtKey().Market == nil ||
			msg.GetSelect().GetByOhlcvtKey().Source == nil ||
			msg.GetSelect().GetByOhlcvtKey().PriceType == 0 {
			return errGet.UpdateMessage("type and name must be specified")
		}
	}

	return nil
}
