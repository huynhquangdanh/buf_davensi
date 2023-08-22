package tradingpairs

import (
	"context"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbTradingPairs "davensi.com/core/gen/tradingpairs"
	pbUoms "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/common"

	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) validateCreate(msg *pbTradingPairs.CreateRequest) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	// Verify that Symbol is specified
	if msg.Symbol == "" {
		return errCreation.UpdateMessage("Symbol must be specified")
	}

	tradingpairRelationship := s.GetRelationship(msg.GetQuantityUom(), msg.GetPriceUom())

	if tradingpairRelationship.qUom == nil {
		return errCreation.UpdateMessage("quantity uom does not exist")
	}
	msg.QuantityUom = &pbUoms.Select{
		Select: &pbUoms.Select_ById{
			ById: tradingpairRelationship.qUom.Id,
		},
	}

	if tradingpairRelationship.pUom == nil {
		return errCreation.UpdateMessage("price uom does not exist")
	}
	msg.PriceUom = &pbUoms.Select{
		Select: &pbUoms.Select_ById{
			ById: tradingpairRelationship.pUom.Id,
		},
	}

	isUniq, errCode := s.IsTradingPairUniq(&msg.Symbol)
	if !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage(fmt.Sprintf("symbol = %s has been used", msg.GetSymbol()))
	}

	return nil
}

func (s *ServiceServer) IsTradingPairUniq(tradingPairSymbol *string) (isUniq bool, errno pbCommon.ErrorCode) {
	tradingPair, err := s.Get(context.Background(), &connect.Request[pbTradingPairs.GetRequest]{
		Msg: &pbTradingPairs.GetRequest{
			Select: &pbTradingPairs.Select{
				Select: &pbTradingPairs.Select_BySymbol{
					BySymbol: *tradingPairSymbol,
				},
			},
		},
	})

	if err == nil && tradingPair.Msg.GetTradingpair() != nil {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if err == nil || tradingPair.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if tradingPair.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, tradingPair.Msg.GetError().Code
	}

	return true, tradingPair.Msg.GetError().Code
}

func ValidateSelect(msg *pbTradingPairs.Select, method string) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_entityName,
		"",
	)
	if msg == nil {
		return errValidate.UpdateMessage("id must be specified")
	}
	if msg.Select == nil {
		return errValidate.UpdateMessage("id must be specified")
	}

	switch msg.GetSelect().(type) {
	case *pbTradingPairs.Select_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errValidate.UpdateMessage("id must be specified")
		}
	case *pbTradingPairs.Select_BySymbol:
		// Verify that symbol is specified
		if msg.GetBySymbol() == "" {
			return errValidate.UpdateMessage("symbol must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	oldTradingPair *pbTradingPairs.TradingPair,
	req *pbTradingPairs.UpdateRequest,
) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	checkSymbol := oldTradingPair.Symbol
	isUpdateSymbol := false

	if req.Symbol != nil {
		checkSymbol = req.GetSymbol()
		isUpdateSymbol = true
	}

	if isUpdateSymbol {
		isUniq, errCode := s.IsTradingPairUniq(&checkSymbol)
		if !isUniq {
			return errUpdate.UpdateCode(errCode).UpdateMessage(
				fmt.Sprintf("symbol '%s' have been used", checkSymbol),
			)
		}
	}

	req.Select = &pbTradingPairs.Select{
		Select: &pbTradingPairs.Select_ById{
			ById: oldTradingPair.Id,
		},
	}

	return nil
}
