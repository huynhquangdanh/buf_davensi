package blockchains

import (
	"context"
	"fmt"

	pbBlockchains "davensi.com/core/gen/blockchains"
	pbCommon "davensi.com/core/gen/common"
	pbUoMs "davensi.com/core/gen/uoms"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/uoms"
	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) IsBlockchainUniq(dsName string) (isUniq bool, errCode pbCommon.ErrorCode) {
	blockchain, err := s.Get(context.Background(), &connect.Request[pbBlockchains.GetRequest]{
		Msg: &pbBlockchains.GetRequest{
			Select: &pbBlockchains.Select{
				Select: &pbBlockchains.Select_ByName{
					ByName: dsName,
				},
			},
		},
	})

	if err == nil || blockchain.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if blockchain.Msg.GetError() != nil && blockchain.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, blockchain.Msg.GetError().Code
	}

	return true, blockchain.Msg.GetError().Code
}

/**
 * @Todo validate provider id exist before insert to db
 */
func (s *ServiceServer) validateCreate(msg *pbBlockchains.CreateRequest) *common.ErrWithCode {
	// Verify that Type and Name are specified
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if msg.Name == "" {
		return errCreation.UpdateMessage("name must be specified")
	}

	if isUniq, errCode := s.IsBlockchainUniq(msg.GetName()); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("name have been used")
	}

	return nil
}

// for Update gRPC
func validateQueryUpdate(msg *pbBlockchains.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_name must be specified")
	}
	switch msg.GetSelect().Select.(type) {
	case *pbBlockchains.Select_ByName:
		if msg.GetSelect().GetByName() == "" {
			return errUpdate.UpdateMessage("name must be specified")
		}
	case *pbBlockchains.Select_ById:
		if msg.GetSelect().GetById() == "" {
			return errUpdate.UpdateMessage("by_id must be specified")
		}
	}

	return nil
}

func validateQueryGet(msg *pbBlockchains.GetRequest) *common.ErrWithCode {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"get",
		_package,
		"",
	)
	if msg.Select == nil {
		return errGet.UpdateMessage("by_id or by_name must be specified")
	}
	switch msg.GetSelect().Select.(type) {
	case *pbBlockchains.Select_ById:
		if msg.GetSelect().GetById() == "" {
			return errGet.UpdateMessage("by_id must be specified")
		}
	case *pbBlockchains.Select_ByName:
		if msg.GetSelect().GetByName() == "" {
			return errGet.UpdateMessage("by_name must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	blockchain *pbBlockchains.Blockchain,
	msg *pbBlockchains.UpdateRequest,
) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)

	if msg.Name != nil && msg.GetName() != blockchain.GetName() {
		if isUniq, errCode := s.IsBlockchainUniq(msg.GetName()); !isUniq {
			return errUpdate.UpdateCode(errCode).UpdateMessage(
				fmt.Sprintf("name '%s' have been used", msg.GetName()),
			)
		}
	}

	return nil
}

func ValidateSelect(selectBlockchain *pbBlockchains.Select, method string) *common.ErrWithCode {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_package,
		"",
	)
	if selectBlockchain == nil || selectBlockchain.Select == nil {
		return errGet.UpdateMessage("by_id or by_name must be specified")
	}
	switch selectBlockchain.GetSelect().(type) {
	case *pbBlockchains.Select_ById:
		if selectBlockchain.GetById() == "" {
			return errGet.UpdateMessage("by_id must be specified")
		}
	case *pbBlockchains.Select_ByName:
		if selectBlockchain.GetByName() == "" {
			return errGet.UpdateMessage("by_name must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateHandleCryptos(selectBlockchain *pbBlockchains.Select, selectUoMs *pbUoMs.SelectList) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"set",
		_package,
		"",
	)
	if validateSelectErr := ValidateSelect(selectBlockchain, "set"); validateSelectErr != nil {
		return validateSelectErr
	}
	if selectUoMs == nil || len(selectUoMs.List) == 0 {
		return errValidate.UpdateMessage("cryptos must be specified")
	}

	for _, selectUoM := range selectUoMs.List {
		if err := uoms.ValidateSelect(selectUoM, "set"); err != nil {
			return err
		}
	}

	return nil
}
