package banks

import (
	"context"

	"davensi.com/core/internal/common"

	pbBanks "davensi.com/core/gen/banks"
	pbCommon "davensi.com/core/gen/common"

	"github.com/bufbuild/connect-go"
	"github.com/google/uuid"
)

func (s *ServiceServer) IsBankUniq(bankName *pbBanks.Select_ByName) (isUniq bool, errno pbCommon.ErrorCode) {
	bank, err := s.Get(context.Background(), &connect.Request[pbBanks.GetRequest]{
		Msg: &pbBanks.GetRequest{
			Select: &pbBanks.Select{
				Select: bankName,
			},
		},
	})

	if err == nil && bank.Msg.GetBank() != nil {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if err == nil || bank.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
	}

	if bank.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, bank.Msg.GetError().Code
	}

	return true, bank.Msg.GetError().Code
}

func (s *ServiceServer) ValidateGet(
	req *pbBanks.GetRequest,
) (errGet *common.ErrWithCode) {
	errGet = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"getting",
		_entityName,
		"",
	)

	if req.GetSelect() == nil {
		return errGet.UpdateMessage("by_id or by_name must be specified")
	}

	return nil
}

func (s *ServiceServer) validateCreate(
	req *pbBanks.CreateRequest,
) (parentID *uuid.UUID, errCreate *common.ErrWithCode) {
	// Verify that Name, Bic, BankCode is specified
	if req.GetName() == "" {
		return parentID, errCreate.UpdateMessage("name must be specified")
	}
	if req.GetBic() == "" {
		return parentID, errCreate.UpdateMessage("bic must be specified")
	}
	if req.GetBankCode() == "" {
		return parentID, errCreate.UpdateMessage("bank_code must be specified")
	}
	if req.GetParent() != nil {
		parent, _ := s.getBankSelect(req.GetParent())
		if parent != nil {
			parse, err := uuid.Parse(parent.GetId())
			if err != nil {
				parentID = &parse
			}
		} else {
			return parentID, errCreate.UpdateMessage("parentID must be specified")
		}
	}

	if isUniq, errno := s.IsBankUniq(&pbBanks.Select_ByName{
		ByName: req.GetName(),
	}); !isUniq {
		return parentID, errCreate.UpdateCode(errno).UpdateMessage("Name must be specified")
	}

	return parentID, nil
}

func (s *ServiceServer) validateUpdate(
	req *pbBanks.UpdateRequest,
) (updateBankID *uuid.UUID, pkResNew BankRelationshipIds, errUpdate *common.ErrWithCode) {
	errUpdate = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)

	if req.GetSelect() == nil {
		return updateBankID, pkResNew, errUpdate.UpdateMessage("by_id or by_name must be specified")
	}

	pkResNew = s.GetBankRelationshipIds(
		req.GetParent(),
	)

	BankResponse, errGetBank := s.getBankSelect(req.GetSelect())
	if BankResponse != nil {
		parse, err := uuid.Parse(BankResponse.GetId())
		if err != nil {
			updateBankID = &parse
		} else {
			return updateBankID, pkResNew, errUpdate.UpdateMessage("BankID cant not parse")
		}
	}
	if errGetBank != nil {
		return updateBankID, pkResNew, errUpdate.UpdateCode(errGetBank.Code).UpdateMessage("Bank not found")
	}

	checkName := BankResponse.GetName()

	// Verify that Name, Bic, BankCode is specified
	if req.Name != nil {
		if req.GetName() == "" {
			return updateBankID, pkResNew, errUpdate.UpdateMessage("name must be specified")
		} else {
			checkName = req.GetName()
		}
	}
	if req.Bic != nil && req.GetBic() == "" {
		return updateBankID, pkResNew, errUpdate.UpdateMessage("bic must be specified")
	}
	if req.BankCode != nil && req.GetBankCode() == "" {
		return updateBankID, pkResNew, errUpdate.UpdateMessage("bank_code must be specified")
	}

	if checkName != BankResponse.GetName() {
		if isUniq, errCheckUniq := s.IsBankUniq(&pbBanks.Select_ByName{
			ByName: req.GetName(),
		}); !isUniq {
			return updateBankID, pkResNew, errUpdate.UpdateCode(errCheckUniq).UpdateMessage("Name must be specified")
		}
	}
	return updateBankID, pkResNew, nil
}

func ValidateSelect(selectBank *pbBanks.Select, method string) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_entityName,
		"",
	)

	if selectBank == nil {
		return errValidate.UpdateMessage("bank must be specified")
	}

	if selectBank.Select == nil { // panic if selectBank is nil from the start
		return errValidate.UpdateMessage("by_id or by_name must be specified")
	}

	switch selectBank.GetSelect().(type) {
	case *pbBanks.Select_ByName:
		if selectBank.GetByName() == "" {
			return errValidate.UpdateMessage("by_name must be specified")
		}
	case *pbBanks.Select_ById:
		if selectBank.GetById() == "" {
			return errValidate.UpdateMessage("by_id must be specified")
		}
	}

	return nil
}
