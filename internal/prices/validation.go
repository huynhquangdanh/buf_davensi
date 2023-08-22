package prices

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bufbuild/connect-go"

	pbCommon "davensi.com/core/gen/common"
	pbDataSources "davensi.com/core/gen/datasources"
	pbMarkets "davensi.com/core/gen/markets"
	pbPrices "davensi.com/core/gen/prices"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/datasources"
	"davensi.com/core/internal/markets"
)

func (s *ServiceServer) IsPriceUniq(selectByPriceKey *pbPrices.Select_ByPriceKey) (isUniq bool, errCode pbCommon.ErrorCode) {
	price, err := s.Get(context.Background(), &connect.Request[pbPrices.GetRequest]{
		Msg: &pbPrices.GetRequest{
			Select: &pbPrices.Select{
				Select: selectByPriceKey,
			},
		},
	})

	fmt.Println(price, err)

	if err == nil || price.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if price.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, price.Msg.GetError().Code
	}

	return true, price.Msg.GetError().Code
}

func ValidateSelect(selectPrice *pbPrices.Select, method string) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_entityName,
		"",
	)
	if selectPrice == nil {
		return errValidate.UpdateMessage("by_id or by_price_key must be specified")
	}
	if selectPrice.Select == nil {
		return errValidate.UpdateMessage("by_id or by_price_key must be specified")
	}

	switch selectPrice.GetSelect().(type) {
	case *pbPrices.Select_ById:
		if selectPrice.GetById() == "" {
			return errValidate.UpdateMessage("by_id must be specified")
		}
	case *pbPrices.Select_ByPriceKey:
		if selectPrice.GetByPriceKey() == nil {
			return errValidate.UpdateMessage("by_price_key must be specified")
		}
		if errSelectSource := datasources.ValidateSelect(selectPrice.GetByPriceKey().GetSource(), "updating"); errSelectSource != nil {
			return errSelectSource
		}
		if errSelectMarket := markets.ValidateSelect(selectPrice.GetByPriceKey().GetMarket(), "updating"); errSelectMarket != nil {
			return errSelectMarket
		}
		if selectPrice.GetByPriceKey().GetType() == pbMarkets.PriceType_PRICE_TYPE_UNSPECIFIED ||
			selectPrice.GetByPriceKey().GetTimestamp() == nil {
			return errValidate.UpdateMessage("source id, market id, type, timestamp must be specified")
		}
	}

	return nil
}

// for Create gRPC
func (s *ServiceServer) validateCreate(msg *pbPrices.CreateRequest) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if errSelectSource := datasources.ValidateSelect(msg.GetSource(), "updating"); errSelectSource != nil {
		return errSelectSource
	}
	if errSelectMarket := markets.ValidateSelect(msg.GetMarket(), "updating"); errSelectMarket != nil {
		return errSelectMarket
	}
	if msg.Type == 0 {
		return errCreation.UpdateMessage("type must be specified")
	}
	if msg.Timestamp == nil {
		return errCreation.UpdateMessage("timestamp must be specified")
	}
	if msg.Price == nil {
		return errCreation.UpdateMessage("price must be specified")
	} else if _, parseErr := strconv.ParseFloat(msg.Price.GetValue(), 64); parseErr != nil {
		return errCreation.UpdateMessage("price must be decimal value")
	}
	if msg.Status == pbCommon.Status_STATUS_UNSPECIFIED {
		return errCreation.UpdateMessage("status must be specified")
	}
	priceRl := s.GetRelationship(
		msg.GetMarket(),
		msg.GetSource(),
	)

	if priceRl.DataSource == nil {
		return errCreation.UpdateMessage("datasource is not exist")
	}
	if priceRl.Market == nil {
		return errCreation.UpdateMessage("market is not exist")
	}

	msg.Market = &pbMarkets.Select{
		Select: &pbMarkets.Select_ById{
			ById: priceRl.Market.Id,
		},
	}
	msg.Source = &pbDataSources.Select{
		Select: &pbDataSources.Select_ById{
			ById: priceRl.DataSource.Id,
		},
	}

	if isUniq, errCode := s.IsPriceUniq(&pbPrices.Select_ByPriceKey{
		ByPriceKey: &pbPrices.PriceKey{
			Source:    msg.GetSource(),
			Market:    msg.GetMarket(),
			Type:      msg.GetType(),
			Timestamp: msg.GetTimestamp(),
		},
	}); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("source id, market id, type and timestamp have been used")
	}

	return nil
}

// for Update gRPC
func (s *ServiceServer) validateQueryUpdate(msg *pbPrices.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updateing",
		_entityName,
		"",
	)
	if msg.Price != nil {
		if _, parseErr := strconv.ParseFloat(msg.Price.GetValue(), 64); parseErr != nil {
			return errUpdate.UpdateMessage("price must be decimal value")
		}
	}
	if errValidateSelect := ValidateSelect(msg.GetSelect(), "updating"); errValidateSelect != nil {
		return errValidateSelect
	}

	priceRl := s.GetRelationship(
		msg.GetMarket(),
		msg.GetSource(),
	)

	if msg.Source != nil && priceRl.DataSource == nil {
		return errUpdate.UpdateMessage("datasource is not exist")
	}
	if msg.Market != nil && priceRl.Market == nil {
		return errUpdate.UpdateMessage("market is not exist")
	}

	if priceRl.DataSource != nil {
		msg.Source = &pbDataSources.Select{
			Select: &pbDataSources.Select_ById{
				ById: priceRl.DataSource.Id,
			},
		}
	}

	if priceRl.Market != nil {
		msg.Market = &pbMarkets.Select{
			Select: &pbMarkets.Select_ById{
				ById: priceRl.Market.Id,
			},
		}
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	msg *pbPrices.UpdateRequest,
	oldPrice *pbPrices.Price,
) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)
	isUpdate := false
	checkSourceID := oldPrice.GetSource().GetId()
	checkMarketID := oldPrice.GetMarket().GetId()
	checkType := oldPrice.GetType()
	checkTimestamp := oldPrice.GetTimestamp()

	if msg.Source != nil && msg.GetSource().GetById() != checkSourceID {
		checkSourceID = msg.GetSource().GetById()
		isUpdate = true
	}
	if msg.Market != nil && msg.GetMarket().GetById() != checkMarketID {
		checkMarketID = msg.GetMarket().GetById()
		isUpdate = true
	}
	if msg.GetType() != checkType {
		checkType = msg.GetType()
		isUpdate = true
	}
	if msg.Timestamp.AsTime() != checkTimestamp.AsTime() {
		checkTimestamp = msg.GetTimestamp()
		isUpdate = true
	}

	if isUpdate {
		if isUniq, errCode := s.IsPriceUniq(&pbPrices.Select_ByPriceKey{
			ByPriceKey: &pbPrices.PriceKey{
				Source: &pbDataSources.Select{
					Select: &pbDataSources.Select_ById{
						ById: checkSourceID,
					},
				},
				Market: &pbMarkets.Select{
					Select: &pbMarkets.Select_ById{
						ById: checkMarketID,
					},
				},
				Type:      checkType,
				Timestamp: checkTimestamp,
			},
		}); !isUniq {
			return errCreation.UpdateCode(errCode).UpdateMessage("source id, market id, type and timestamp have been used")
		}
	}

	return nil
}
