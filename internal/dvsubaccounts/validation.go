package dvsubaccounts

import (
	"context"

	pbCommon "davensi.com/core/gen/common"
	pbDvSubAccounts "davensi.com/core/gen/dvsubaccounts"
	pbRecipients "davensi.com/core/gen/recipients"
	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) validateQueryUpdate(req *pbDvSubAccounts.UpdateRequest) (errUpdate *common.ErrWithCode) {
	errUpdate = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)

	if req.GetRecipient().GetSelect() == nil {
		return errUpdate.UpdateMessage("recipient select must be specified")
	}
	if _, errDvbot := GetSingletonServiceServer(s.db).Get(
		context.Background(),
		connect.NewRequest(&pbRecipients.GetRequest{
			Select: req.GetRecipient().GetSelect(),
		}),
	); errDvbot != nil {
		return errUpdate.UpdateMessage(errDvbot.Error())
	}

	return nil
}

func (s *ServiceServer) validateQueryInsert(req *pbDvSubAccounts.CreateRequest) (errCreate *common.ErrWithCode) {
	errCreate = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if req.GetRecipient().GetType() != pbRecipients.Type_TYPE_DV_SUBACCOUNT {
		return errCreate.UpdateMessage("Recipient's type must be DV_SUBACCOUNT")
	}
	return nil
}
