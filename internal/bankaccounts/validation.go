package bankaccounts

import (
	"davensi.com/core/internal/common"

	pbBankAccounts "davensi.com/core/gen/bankaccounts"
	pbBankBranches "davensi.com/core/gen/bankbranches"
	pbCommon "davensi.com/core/gen/common"
	pbUoms "davensi.com/core/gen/uoms"
)

// Big TODO: deal with relationships

// for Create gRPC
func (s *ServiceServer) validateCreate(msg *pbBankAccounts.CreateRequest) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	// Verify that Recipient is specified
	if msg.Recipient == nil {
		return errCreation.UpdateMessage("recipient must be specified")
	}

	// Optional Bank Branch and Currency field
	bankAccountRl := s.GetRelationship(
		msg.GetBankBranch(),
		msg.GetCurrency(),
	)

	if msg.BankBranch != nil && bankAccountRl.BankBranch == nil {
		return errCreation.UpdateMessage("bank branch does not exist")
	} else if bankAccountRl.BankBranch != nil {
		msg.BankBranch = &pbBankBranches.Select{
			Select: &pbBankBranches.Select_ById{
				ById: bankAccountRl.BankBranch.Id,
			},
		}
	}

	if msg.Currency != nil && bankAccountRl.Currency == nil {
		return errCreation.UpdateMessage("currency does not exist")
	} else {
		msg.Currency = &pbUoms.Select{
			Select: &pbUoms.Select_ById{
				ById: bankAccountRl.Currency.Id,
			},
		}
	}

	return nil
}

// For Update gRPC
// Check whether the relationships exist
func (s *ServiceServer) validateUpdateQuery(msg *pbBankAccounts.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)

	bankAccountRl := s.GetRelationship(
		msg.GetBankBranch(),
		msg.GetCurrency(),
	)

	if msg.BankBranch != nil && bankAccountRl.BankBranch == nil {
		return errUpdate.UpdateMessage("bank branch does not exist")
	}

	if msg.Currency != nil && bankAccountRl.Currency == nil {
		return errUpdate.UpdateMessage("currency does not exist")
	}

	if bankAccountRl.BankBranch != nil {
		msg.BankBranch = &pbBankBranches.Select{
			Select: &pbBankBranches.Select_ById{
				ById: bankAccountRl.BankBranch.Id,
			},
		}
	}

	if bankAccountRl.Currency != nil {
		msg.Currency = &pbUoms.Select{
			Select: &pbUoms.Select_ById{
				ById: bankAccountRl.Currency.Id,
			},
		}
	}

	return nil
}
