package dvbots

import (
	"context"

	pbCommon "davensi.com/core/gen/common"
	pbDVBots "davensi.com/core/gen/dvbots"
	pbRecipients "davensi.com/core/gen/recipients"
	"davensi.com/core/internal/common"
	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) validateQueryUpdate(req *pbDVBots.UpdateRequest) (errUpdate *common.ErrWithCode) {
	errUpdate = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)
	if errValidateSelect := s.validateSelect(req.Recipient.Select, "updating"); errValidateSelect != nil {
		return errValidateSelect
	}

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

func (s *ServiceServer) validateCreate(req *pbDVBots.CreateRequest) (errCreate *common.ErrWithCode) {
	errCreate = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if req.GetRecipient().GetType() != pbRecipients.Type_TYPE_DV_BOT ||
		req.Recipient == nil {
		return errCreate.UpdateMessage("Recipient's type must be DV_BOT")
	}
	return nil
}

func (s *ServiceServer) validateSelect(msg *pbRecipients.Select, method string) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_package,
		"",
	)
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_legal_entity_user_label must be specified")
	}
	switch msg.GetSelect().(type) {
	case *pbRecipients.Select_ByLegalEntityUserLabel:
		if msg.GetByLegalEntityUserLabel() == nil {
			return errUpdate.UpdateMessage("by_legal_entity_user_label must be specified")
		}
		if msg.GetByLegalEntityUserLabel().LegalEntity == nil ||
			msg.GetByLegalEntityUserLabel().User == nil ||
			msg.GetByLegalEntityUserLabel().Label == "" {
			return errUpdate.UpdateMessage("legal entity, label and user must be specified")
		}
	case *pbRecipients.Select_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errUpdate.UpdateMessage("by_id must be specified")
		}
	}

	return nil
}
