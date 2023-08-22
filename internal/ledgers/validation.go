package ledgers

import (
	"context"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbLedgers "davensi.com/core/gen/ledgers"
	"davensi.com/core/internal/common"

	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) IsLedgerUniq(ledgerName *pbLedgers.Select_ByName) (isUniq bool, errCode pbCommon.ErrorCode) {
	ledger, err := s.Get(context.Background(), &connect.Request[pbLedgers.GetRequest]{
		Msg: &pbLedgers.GetRequest{
			Select: &pbLedgers.GetRequest_ByName{
				ByName: ledgerName.ByName,
			},
		},
	})

	if err == nil || ledger.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if ledger.Msg.GetError() != nil && ledger.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, ledger.Msg.GetError().Code
	}

	return true, ledger.Msg.GetError().Code
}

// for Create gRPC
func (s *ServiceServer) validateCreate(msg *pbLedgers.CreateRequest) *common.ErrWithCode {
	// Verify that Name are specified
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if msg.Name == "" {
		return errCreation.UpdateMessage("name must be specified")
	}

	if isUniq, errCode := s.IsLedgerUniq(&pbLedgers.Select_ByName{
		ByName: msg.GetName(),
	}); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("name have been used")
	}

	return nil
}

// for Update gRPC
func validateQueryUpdate(msg *pbLedgers.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_name must be specified")
	}
	switch msg.GetSelect().(type) {
	case *pbLedgers.UpdateRequest_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errUpdate.UpdateMessage("by_id must be specified")
		}
	case *pbLedgers.UpdateRequest_ByName:
		// Verify that Type and Symbol are specified
		if msg.GetByName() == "" {
			return errUpdate.UpdateMessage("by_name must be specified")
		}
	}

	return nil
}

func validateQueryGet(msg *pbLedgers.GetRequest) *common.ErrWithCode {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"get",
		_package,
		"",
	)
	if msg.Select == nil {
		return errGet.UpdateMessage("by_id or by_name must be specified")
	}
	switch msg.GetSelect().(type) {
	case *pbLedgers.GetRequest_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errGet.UpdateMessage("by_id must be specified")
		}
	case *pbLedgers.GetRequest_ByName:
		if msg.GetByName() == "" {
			return errGet.UpdateMessage("by_name must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	oldLedger *pbLedgers.Ledger,
	msg *pbLedgers.UpdateRequest,
) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.Name != nil && msg.GetName() != oldLedger.Name {
		if isUniq, errCode := s.IsLedgerUniq(&pbLedgers.Select_ByName{
			ByName: msg.GetName(),
		}); !isUniq {
			return errUpdate.UpdateCode(errCode).UpdateMessage(
				fmt.Sprintf("name '%s' have been used", msg.GetName()),
			)
		}
	}

	return nil
}
