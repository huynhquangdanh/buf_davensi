package recipients

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/google/uuid"

	pbCommon "davensi.com/core/gen/common"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbRecipients "davensi.com/core/gen/recipients"
	pbUsers "davensi.com/core/gen/users"

	"davensi.com/core/internal/common"
)

func (s *ServiceServer) IsRecipientUniq(legalEntityUserLabel *pbRecipients.Select_ByLegalEntityUserLabel,
) (isUniq bool, errno pbCommon.ErrorCode) {
	recipient, err := s.Get(context.Background(), &connect.Request[pbRecipients.GetRequest]{
		Msg: &pbRecipients.GetRequest{
			Select: &pbRecipients.Select{
				Select: legalEntityUserLabel,
			},
		},
	})

	if err != nil {
		if recipient.Msg.GetError().GetCode() == pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
			return true, pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED
		}
		return false, recipient.Msg.GetError().GetCode()
	}

	if recipient.Msg.GetRecipient() != nil {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	return true, pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED
}

func (s *ServiceServer) ValidateGet(
	req *pbRecipients.GetRequest,
) (pkRes RecipientRelationshipIds, errGet *common.ErrWithCode) {
	errGet = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"getting",
		_entityName,
		"",
	)

	if req.GetSelect() == nil {
		return pkRes, errGet.UpdateMessage("by_id or by_legal_entity_user_label must be specified")
	}
	if req.GetSelect().GetByLegalEntityUserLabel() != nil {
		selectUser := req.GetSelect().GetByLegalEntityUserLabel().GetUser()
		selectLegalEntity := req.GetSelect().GetByLegalEntityUserLabel().GetLegalEntity()
		if req.GetSelect().GetByLegalEntityUserLabel().GetLabel() == "" {
			return pkRes, errGet.UpdateMessage("label legalEntity must be specified")
		}
		if selectUser == nil && selectLegalEntity == nil {
			return pkRes, errGet.UpdateMessage("user and legalEntity must be specified")
		}
		if selectUser.GetById() != "" {
			selectUser = nil
		}
		if selectLegalEntity.GetById() != "" {
			selectLegalEntity = nil
		}
		if selectUser != nil || selectLegalEntity != nil {
			pkRes = s.GetRecipientRelationshipIds(
				selectUser,
				selectLegalEntity,
				nil,
			)
		}
	}

	return pkRes, nil
}

func (s *ServiceServer) ValidateCreate(
	req *pbRecipients.CreateRequest,
) (pkRes RecipientRelationshipIds, errCreate *common.ErrWithCode) {
	errCreate = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	pkRes = s.GetRecipientRelationshipIds(
		req.GetUser(),
		req.GetLegalEntity(),
		req.GetOrg(),
	)

	if req.GetLabel() == "" {
		return pkRes, errCreate.UpdateMessage("label must be specified")
	}
	UserIDString := ""
	if pkRes.UserID == nil {
		UserIDString = ""
	} else {
		UserIDString = pkRes.UserID.String()
	}
	LegalEntityIDString := ""
	if pkRes.LegalEntityID == nil {
		LegalEntityIDString = ""
	} else {
		LegalEntityIDString = pkRes.LegalEntityID.String()
	}
	if pkRes.UserID == nil && pkRes.LegalEntityID == nil {
		return pkRes, errCreate.UpdateMessage("user and legalEntity must be specified")
	}
	if isUniq, errno := s.IsRecipientUniq(&pbRecipients.Select_ByLegalEntityUserLabel{
		ByLegalEntityUserLabel: &pbRecipients.LegalEntityUserLabel{
			LegalEntity: &pbLegalEntities.Select{
				Select: &pbLegalEntities.Select_ById{
					ById: LegalEntityIDString,
				},
			},
			User: &pbUsers.Select{
				Select: &pbUsers.Select_ById{
					ById: UserIDString,
				},
			},
			Label: req.GetLabel(),
		},
	}); !isUniq {
		return pkRes, errCreate.UpdateCode(errno).UpdateMessage("LegalEntity/User/Label must be specified")
	}

	return pkRes, nil
}

func (s *ServiceServer) ValidateUpdate(
	req *pbRecipients.UpdateRequest,
) (pkResOld, pkResNew RecipientRelationshipIds, errUpdate *common.ErrWithCode) {
	errUpdate = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)

	if req.GetSelect() == nil {
		return pkResOld, pkResNew, errUpdate.UpdateMessage("by_id or by_legal_entity_user_label must be specified")
	}

	recipientResponse, errGetRecipient := s.getRecipientSelect(req.GetSelect())
	if errGetRecipient != nil {
		return pkResOld, pkResNew, errUpdate.UpdateCode(errGetRecipient.Code).UpdateMessage("Recipient not found")
	}

	UserID, _ := uuid.Parse(recipientResponse.GetUser().GetId())
	LegalEntityID, _ := uuid.Parse(recipientResponse.GetLegalEntity().GetId())

	pkResOld.UserID = &UserID
	pkResOld.LegalEntityID = &LegalEntityID

	pkResNew = s.GetRecipientRelationshipIds(
		req.GetUser(),
		req.GetLegalEntity(),
		req.GetOrg(),
	)

	if req.User != nil && pkResNew.UserID == nil && req.LegalEntity != nil && pkResNew.LegalEntityID == nil {
		return pkResOld, pkResNew, errUpdate.UpdateMessage("LegalEntity/User must be specified")
	}

	checkLabel := recipientResponse.GetLabel()
	if req.Label != nil {
		if req.GetLabel() == "" {
			return pkResOld, pkResNew, errUpdate.UpdateMessage("Label must be specified")
		} else {
			checkLabel = req.GetLabel()
		}
	}
	checkUserID := recipientResponse.GetUser().GetId()
	if pkResNew.UserID != nil {
		checkUserID = pkResNew.UserID.String()
	}
	checkLegalEntityID := recipientResponse.GetLegalEntity().GetId()
	if pkResNew.LegalEntityID != nil {
		checkLegalEntityID = pkResNew.LegalEntityID.String()
	}

	if (checkLabel != recipientResponse.GetLabel()) ||
		(checkUserID != recipientResponse.GetUser().GetId()) ||
		(checkLegalEntityID != recipientResponse.GetLegalEntity().GetId()) {
		if isUniq, errCheckUniq := s.IsRecipientUniq(&pbRecipients.Select_ByLegalEntityUserLabel{
			ByLegalEntityUserLabel: &pbRecipients.LegalEntityUserLabel{
				LegalEntity: &pbLegalEntities.Select{
					Select: &pbLegalEntities.Select_ById{
						ById: checkLegalEntityID,
					},
				},
				User: &pbUsers.Select{
					Select: &pbUsers.Select_ById{
						ById: checkUserID,
					},
				},
				Label: req.GetLabel(),
			},
		}); !isUniq {
			return pkResOld, pkResNew, errUpdate.UpdateCode(errCheckUniq).UpdateMessage("LegalEntity/User/Label must be specified")
		}
	}
	return pkResOld, pkResNew, nil
}
