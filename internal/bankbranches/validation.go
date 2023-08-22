package bankbranches

import (
	"context"

	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"

	pbBankBranches "davensi.com/core/gen/bankbranches"
	pbBanks "davensi.com/core/gen/banks"
	pbCommon "davensi.com/core/gen/common"
	"davensi.com/core/internal/banks"
)

// for Create gRPC
func (s *ServiceServer) validateCreate(msg *pbBankBranches.CreateRequest) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	// Verify that Bank, Branch Code, Type and Name are specified
	if msg.BranchCode == "" {
		return errCreation.UpdateMessage("branch_code must be specified")
	}

	if msg.Type == pbBanks.Type_TYPE_UNSPECIFIED {
		return errCreation.UpdateMessage("bank type must be specified")
	}

	if msg.Name == "" {
		return errCreation.UpdateMessage("name must be specified")
	}

	if errSelectBank := banks.ValidateSelect(msg.GetBank(), "creating"); errSelectBank != nil {
		return errSelectBank
	}

	bankBranchRl := s.GetRelationship(
		msg.GetBank(),
	)

	if bankBranchRl.Bank == nil {
		return errCreation.UpdateMessage("bank does not exist")
	}

	msg.Bank = &pbBanks.Select{
		Select: &pbBanks.Select_ById{
			ById: bankBranchRl.Bank.Id,
		},
	}

	if isUniq, errCode := s.IsBankBranchUnique(&pbBankBranches.Select_ByBankBranchCode{
		ByBankBranchCode: &pbBankBranches.BankBranchCode{
			Bank:       msg.GetBank(),
			BranchCode: msg.GetBranchCode(),
		},
	}); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("bank, branch code have been used")
	}

	return nil
}

func (s *ServiceServer) IsBankBranchUnique(selectByBankBranchCode *pbBankBranches.Select_ByBankBranchCode) (
	isUniq bool, errCode pbCommon.ErrorCode,
) {
	bankBranch, err := s.Get(context.Background(), &connect.Request[pbBankBranches.GetRequest]{
		Msg: &pbBankBranches.GetRequest{
			Select: &pbBankBranches.Select{
				Select: selectByBankBranchCode,
			},
		},
	})

	if err == nil || bankBranch.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if bankBranch.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, bankBranch.Msg.GetError().Code
	}

	return true, bankBranch.Msg.GetError().Code
}

func (s *ServiceServer) validateUpdateQuery(msg *pbBankBranches.UpdateRequest) *common.ErrWithCode {
	if errValidateSelect := ValidateSelect(msg.GetSelect(), "updating"); errValidateSelect != nil {
		return errValidateSelect
	}

	return nil
}

func ValidateSelect(selectBankBranch *pbBankBranches.Select, method string) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_entityName,
		"",
	)

	if selectBankBranch == nil {
		return errValidate.UpdateMessage("Select by ID or Bank Branch Code must be specified")
	}

	if selectBankBranch.Select == nil {
		return errValidate.UpdateMessage("Select by ID or Bank Branch Code must be specified")
	}

	switch selectBankBranch.GetSelect().(type) {
	case *pbBankBranches.Select_ById:
		if selectBankBranch.GetById() == "" {
			return errValidate.UpdateMessage("by_id must be specified")
		}
	case *pbBankBranches.Select_ByBankBranchCode:
		if selectBankBranch.GetByBankBranchCode() == nil {
			return errValidate.UpdateMessage("by_bank_branch_code must be specified")
		} else {
			if selectBankBranch.GetByBankBranchCode().GetBranchCode() == "" {
				return errValidate.UpdateMessage("branch_code must be specified")
			}
			if selectBankBranch.GetByBankBranchCode().GetBank() == nil {
				return errValidate.UpdateMessage("bank must be specified")
			} else {
				switch selectBankBranch.GetByBankBranchCode().GetBank().GetSelect().(type) {
				case *pbBanks.Select_ById:
					if selectBankBranch.GetByBankBranchCode().GetBank().GetById() == "" {
						return errValidate.UpdateMessage("bank_id must be specified")
					}
				case *pbBanks.Select_ByName:
					if selectBankBranch.GetByBankBranchCode().GetBank().GetByName() == "" {
						return errValidate.UpdateMessage("bank_name must be specified")
					}
				}
			}
		}
	}

	return nil
}
