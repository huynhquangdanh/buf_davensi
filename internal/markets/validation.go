package markets

import (
	"context"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbMarkets "davensi.com/core/gen/markets"
	pbTradingPairs "davensi.com/core/gen/tradingpairs"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/tradingpairs"
	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) IsMarketUniq(marketSymbol *pbMarkets.Select_BySymbol) (isUniq bool, errCode pbCommon.ErrorCode) {
	market, err := s.Get(context.Background(), &connect.Request[pbMarkets.GetRequest]{
		Msg: &pbMarkets.GetRequest{
			Select: &pbMarkets.Select{
				Select: marketSymbol,
			},
		},
	})

	if err == nil || market.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if market.Msg.GetError() != nil && market.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, market.Msg.GetError().Code
	}

	return true, market.Msg.GetError().Code
}

func ValidateSelect(selectMarket *pbMarkets.Select, method string) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_entityName,
		"",
	)
	if selectMarket == nil {
		return errUpdate.UpdateMessage("id must be specified")
	}
	if selectMarket.Select == nil {
		return errUpdate.UpdateMessage("id must be specified")
	}
	switch selectMarket.GetSelect().(type) {
	case *pbMarkets.Select_ById:
		if selectMarket.GetById() == "" {
			return errUpdate.UpdateMessage("id must be specified")
		}
	case *pbMarkets.Select_BySymbol:
		if selectMarket.GetBySymbol() == "" {
			return errUpdate.UpdateMessage("symbol must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateCreate(msg *pbMarkets.CreateRequest) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	if msg.Symbol == "" {
		return errCreation.UpdateMessage("type and symbol must be specified")
	}

	if errSelectTradingPair := tradingpairs.ValidateSelect(msg.Tradingpair, "creating"); errSelectTradingPair != nil {
		return errSelectTradingPair
	}

	marketRelationship := s.GetRelationship(msg.GetTradingpair())

	if marketRelationship.tradingpair == nil {
		return errCreation.UpdateMessage("trading pair does not exist")
	}

	msg.Tradingpair = &pbTradingPairs.Select{
		Select: &pbTradingPairs.Select_ById{
			ById: marketRelationship.tradingpair.Id,
		},
	}

	if isUniq, errCode := s.IsMarketUniq(&pbMarkets.Select_BySymbol{
		BySymbol: msg.GetSymbol(),
	}); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage(
			fmt.Sprintf("symbol = '%s' have been used", msg.GetSymbol()),
		)
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	market *pbMarkets.Market,
	msg *pbMarkets.UpdateRequest,
) *common.ErrWithCode {
	if msg.Symbol != nil && msg.GetSymbol() != market.GetSymbol() {
		if isUniq, errCode := s.IsMarketUniq(&pbMarkets.Select_BySymbol{
			BySymbol: msg.GetSymbol(),
		}); !isUniq {
			return common.CreateErrWithCode(
				errCode,
				"updating",
				_entityName,
				fmt.Sprintf("symbol = '%s' have been used", msg.GetSymbol()),
			)
		}
	}

	return nil
}
