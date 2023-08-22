package defiwallets

import (
	"context"

	"github.com/bufbuild/connect-go"

	pbBlockchains "davensi.com/core/gen/blockchains"
	pbCommon "davensi.com/core/gen/common"
	pbDefiwallets "davensi.com/core/gen/defiwallets"
	pbRecipients "davensi.com/core/gen/recipients"
	"davensi.com/core/internal/common"
)

func (s *ServiceServer) validateQueryInsert(req *pbDefiwallets.CreateRequest) (errCreate *common.ErrWithCode) {
	errCreate = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if req.GetRecipient().GetType() != pbRecipients.Type_TYPE_DEFI_WALLET {
		return errCreate.UpdateMessage("Recipient's type must be DEFI_WALLET")
	}
	return nil
}

func (s *ServiceServer) validateCreate(msg *pbDefiwallets.CreateRequest) *common.ErrWithCode {
	// Verify that Type and Name are specified
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if msg.Recipient == nil || msg.Blockchain == nil || msg.Address == "" {
		return errCreation.UpdateMessage("properties must be specified")
	}
	defiwalletRl := s.GetRelationship(msg.GetBlockchain())

	msg.Blockchain = &pbBlockchains.Select{
		Select: &pbBlockchains.Select_ById{
			ById: defiwalletRl.blockchain.Id,
		},
	}

	return nil
}

func (s *ServiceServer) validateQueryUpdate(req *pbDefiwallets.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
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
