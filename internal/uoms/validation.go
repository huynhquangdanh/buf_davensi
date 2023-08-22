package uoms

import (
	"context"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbUoMs "davensi.com/core/gen/uoms"
	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
)

func ValidateSelectList(selectUoMs *pbUoMs.SelectList, method string) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_package,
		"",
	)

	if len(selectUoMs.GetList()) == 0 {
		return errValidate.UpdateMessage("list must be specified")
	}

	for index, selectUom := range selectUoMs.GetList() {
		if errValidateSelect := ValidateSelect(selectUom, method); errValidateSelect != nil {
			return errValidate.UpdateMessage(
				fmt.Sprintf("select index: %d have error: %s", index, errValidate.Err.Error()),
			)
		}
	}

	return nil
}

func ValidateSelect(uomSelect *pbUoMs.Select, method string) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_entityName,
		"",
	)

	if uomSelect == nil {
		return errValidate.UpdateMessage("currency1 must be specified")
	}

	if uomSelect.Select == nil { // panic if msg is nill from the start
		return errValidate.UpdateMessage("by_id or by_type_symbol must be specified")
	}

	switch uomSelect.GetSelect().(type) {
	case *pbUoMs.Select_ByTypeSymbol:
		if uomSelect.GetByTypeSymbol() == nil {
			return errValidate.UpdateMessage("by_type_symbol must be specified")
		}
		if uomSelect.GetByTypeSymbol().Type == 0 || uomSelect.GetByTypeSymbol().Symbol == "" {
			return errValidate.UpdateMessage("type and symbol must be specified")
		}
	case *pbUoMs.Select_ById:
		// Verify that ID is specified
		if uomSelect.GetById() == "" {
			return errValidate.UpdateMessage("by_id must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) IsUomUniq(uomTypeSymbol *pbUoMs.TypeSymbol) (isUniq bool, errCode pbCommon.ErrorCode) {
	uom, err := s.Get(context.Background(), &connect.Request[pbUoMs.GetRequest]{
		Msg: &pbUoMs.GetRequest{
			Select: &pbUoMs.Select{
				Select: &pbUoMs.Select_ByTypeSymbol{
					ByTypeSymbol: uomTypeSymbol,
				},
			},
		},
	})

	if err == nil || uom.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if uom.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, uom.Msg.GetError().Code
	}

	return true, uom.Msg.GetError().Code
}

// for Create gRPC
func (s *ServiceServer) validateMsgCreate(msg *pbUoMs.CreateRequest) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if msg.Type == 0 || msg.Symbol == "" {
		return errCreation.UpdateMessage("type and symbol must be specified")
	}

	if isUniq, errCode := s.IsUomUniq(&pbUoMs.TypeSymbol{
		Type:   msg.GetType(),
		Symbol: msg.GetSymbol(),
	}); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("type and symbol have been used")
	}

	if msg.ManagedDecimals != nil && msg.GetManagedDecimals() > uint32(^uint16(0)) {
		return errCreation.UpdateMessage("managed_decimals must be uint16")
	}

	if msg.DisplayedDecimals != nil && msg.GetDisplayedDecimals() > uint32(^uint16(0)) {
		return errCreation.UpdateMessage("displayed_decimals must be uint16")
	}

	return nil
}

// for Update gRPC
func validateQueryUpdate(msg *pbUoMs.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_type_symbol or by_id must be specified")
	}

	switch msg.GetSelect().Select.(type) {
	case *pbUoMs.Select_ById:
		if msg.Select.GetById() == "" {
			return errUpdate.UpdateMessage("by_id must be specified")
		}
	case *pbUoMs.Select_ByTypeSymbol:
		if msg.Select.GetByTypeSymbol() == nil {
			return errUpdate.UpdateMessage("by_type_symbol must be specified")
		}
		if msg.Select.GetByTypeSymbol().GetSymbol() == "" ||
			msg.Select.GetByTypeSymbol().GetType() == pbUoMs.Type_TYPE_UNSPECIFIED {
			return errUpdate.UpdateMessage("type and symbol must be specified")
		}
	}

	if msg.ManagedDecimals != nil && msg.GetManagedDecimals() > uint32(^uint16(0)) {
		return errUpdate.UpdateMessage("managed_decimals must be uint16")
	}

	if msg.DisplayedDecimals != nil && msg.GetDisplayedDecimals() > uint32(^uint16(0)) {
		return errUpdate.UpdateMessage("displayed_decimals must be uint16")
	}

	return nil
}

// for Update gRPC
func ValidateQueryGet(msg *pbUoMs.GetRequest) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"fetching",
		_entityName,
		"",
	)
	if msg.Select == nil {
		return errValidate.UpdateMessage("by_type_symbol or by_id must be specified")
	}

	switch msg.GetSelect().Select.(type) {
	case *pbUoMs.Select_ById:
		if msg.Select.GetById() == "" {
			return errValidate.UpdateMessage("by_id must be specified")
		}
	case *pbUoMs.Select_ByTypeSymbol:
		if msg.Select.GetByTypeSymbol() == nil {
			return errValidate.UpdateMessage("by_type_symbol must be specified")
		}
		if msg.Select.GetByTypeSymbol().GetSymbol() == "" ||
			msg.Select.GetByTypeSymbol().GetType() == pbUoMs.Type_TYPE_UNSPECIFIED {
			return errValidate.UpdateMessage("type and symbol must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateMsgUpdate(
	oldUom *pbUoMs.UoM,
	msg *pbUoMs.UpdateRequest,
) *common.ErrWithCode {
	checkType := oldUom.Type
	checkSymbol := oldUom.Symbol
	isUpddateTypeSymbol := false
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)

	if msg.Type != nil && msg.GetType() != oldUom.GetType() {
		checkType = msg.GetType()
		isUpddateTypeSymbol = true
	}

	if msg.Symbol != nil && msg.GetSymbol() != oldUom.GetSymbol() {
		checkSymbol = msg.GetSymbol()
		isUpddateTypeSymbol = true
	}

	if isUpddateTypeSymbol {
		isUomUniq, errCode := s.IsUomUniq(&pbUoMs.TypeSymbol{
			Type:   checkType,
			Symbol: checkSymbol,
		})
		if !isUomUniq {
			return errUpdate.UpdateCode(errCode).UpdateMessage("type/symbol have been used")
		}
	}

	return nil
}
