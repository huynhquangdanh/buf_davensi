package incomes

import (
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbIncomes "davensi.com/core/gen/incomes"
)

// for Create gRPC
func (s *ServiceServer) ValidateCreation(msg *pbIncomes.CreateRequest) (errno pbCommon.ErrorCode, err error) {
	if msg.Select == nil {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf("creating '%s' one of select must be specified", _entityName)
	}

	isSelect := false

	if msg.GetSalary() == nil {
		isSelect = true
	}
	if msg.GetFreelancing() == nil {
		isSelect = true
	}
	if msg.GetDividends() == nil {
		isSelect = true
	}
	if msg.GetInvestment() == nil {
		isSelect = true
	}
	if msg.GetRent() == nil {
		isSelect = true
	}
	if msg.GetPension() == nil {
		isSelect = true
	}
	if msg.GetOther() == nil {
		isSelect = true
	}

	if !isSelect {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf("creating '%s' one of select must be specified", _entityName)
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}

// for Update gRPC
func validateQueryUpdate(msg *pbIncomes.UpdateRequest) error {
	if msg.Id == "" {
		return fmt.Errorf("creating '%s' one of id must be specified", _entityName)
	}
	if msg.Select == nil {
		return fmt.Errorf("creating '%s' one of select must be specified", _entityName)
	}

	isSelect := false

	if msg.GetSalary() == nil {
		isSelect = true
	}
	if msg.GetFreelancing() == nil {
		isSelect = true
	}
	if msg.GetDividends() == nil {
		isSelect = true
	}
	if msg.GetInvestment() == nil {
		isSelect = true
	}
	if msg.GetRent() == nil {
		isSelect = true
	}
	if msg.GetPension() == nil {
		isSelect = true
	}
	if msg.GetOther() == nil {
		isSelect = true
	}

	if !isSelect {
		return fmt.Errorf("creating '%s' one of select must be specified", _entityName)
	}

	return nil
}
