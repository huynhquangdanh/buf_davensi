package recipients

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/google/uuid"

	pbCommon "davensi.com/core/gen/common"
	pbLegalEntities "davensi.com/core/gen/legalentities"
	pbOrgs "davensi.com/core/gen/orgs"
	pbRecipients "davensi.com/core/gen/recipients"
	pbUsers "davensi.com/core/gen/users"

	"davensi.com/core/internal/common"
	"davensi.com/core/internal/legalentities"
	"davensi.com/core/internal/orgs"
	"davensi.com/core/internal/users"
)

type RecipientRelationshipIds struct {
	UserID        *uuid.UUID
	LegalEntityID *uuid.UUID
	OrgID         *uuid.UUID
}

func (s *ServiceServer) GetRecipientRelationshipIds(
	selectUser *pbUsers.Select,
	selectLegalEntity *pbLegalEntities.Select,
	selectOrg *pbOrgs.Select,
) RecipientRelationshipIds {
	userIDChan := make(chan *uuid.UUID)
	legalEntityIDChan := make(chan *uuid.UUID)
	orgIDChan := make(chan *uuid.UUID)

	go func() {
		var existUserID *uuid.UUID
		if selectUser != nil {
			existUserID, _ = s.GetUserID(selectUser)
		}
		userIDChan <- existUserID
	}()

	go func() {
		var existLegalEntityID *uuid.UUID
		if selectLegalEntity != nil {
			existLegalEntityID, _ = s.GetLegalEntityID(selectLegalEntity)
		}
		legalEntityIDChan <- existLegalEntityID
	}()

	go func() {
		var existOrgID *uuid.UUID
		if selectOrg != nil {
			existOrgID, _ = s.GetOrgID(selectOrg)
		}
		orgIDChan <- existOrgID
	}()

	return RecipientRelationshipIds{
		UserID:        <-userIDChan,
		LegalEntityID: <-legalEntityIDChan,
		OrgID:         <-orgIDChan,
	}
}

func (s *ServiceServer) GetUserID(req *pbUsers.Select) (userID *uuid.UUID, errGet *common.ErrWithCode) {
	errGet = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"getting",
		_entityName+"_User",
		"",
	)

	userInput := &pbUsers.GetRequest{}
	login := req.GetByLogin()
	id := req.GetById()

	if req.GetSelect() == nil {
		return nil, errGet.UpdateMessage("by_id or by_login must be specified")
	}
	switch req.GetSelect().(type) {
	case *pbUsers.Select_ByLogin:
		userInput.Select = &pbUsers.GetRequest_ByLogin{
			ByLogin: login,
		}
	case *pbUsers.Select_ById:
		userInput.Select = &pbUsers.GetRequest_ById{
			ById: id,
		}
	}

	userResponse, errUser := users.GetSingletonServiceServer(s.db).Get(
		context.Background(),
		connect.NewRequest(userInput),
	)

	if errUser != nil {
		return nil, errGet.UpdateMessage(errUser.Error())
	}
	parse, err := uuid.Parse(userResponse.Msg.GetUser().Id)
	if err != nil {
		return nil, errGet.UpdateMessage(err.Error())
	}
	return &parse, nil
}

func (s *ServiceServer) GetLegalEntityID(req *pbLegalEntities.Select) (legalEntityID *uuid.UUID, errGet *common.ErrWithCode) {
	errGet = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"getting",
		_entityName+"_LegalEntity",
		"",
	)

	legalEntityInput := &pbLegalEntities.GetRequest{}

	if req.GetSelect() == nil {
		return nil, errGet.UpdateMessage("by_id or by_name must be specified")
	}
	switch req.GetSelect().(type) {
	case *pbLegalEntities.Select_ById:
		legalEntityInput.Select = &pbLegalEntities.Select{
			Select: &pbLegalEntities.Select_ById{
				ById: req.GetById(),
			},
		}
	case *pbLegalEntities.Select_ByName:
		legalEntityInput.Select = &pbLegalEntities.Select{
			Select: &pbLegalEntities.Select_ByName{
				ByName: req.GetByName(),
			},
		}
	}

	legalEntityResponse, errLegalEntity := legalentities.GetSingletonServiceServer(s.db).Get(
		context.Background(),
		connect.NewRequest(legalEntityInput),
	)

	if errLegalEntity != nil {
		return nil, errGet.UpdateMessage(errLegalEntity.Error())
	}
	parse, err := uuid.Parse(legalEntityResponse.Msg.GetLegalEntity().Id)
	if err != nil {
		return nil, errGet.UpdateMessage(err.Error())
	}
	return &parse, nil
}

func (s *ServiceServer) GetOrgID(req *pbOrgs.Select) (orgID *uuid.UUID, errGet *common.ErrWithCode) {
	errGet = common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"getting",
		_entityName+"_Org",
		"",
	)

	orgInput := &pbOrgs.GetRequest{}
	var id, name string
	id = req.GetById()
	name = req.GetByName()

	if req.GetSelect() == nil {
		return nil, errGet.UpdateMessage("by_id or by_name must be specified")
	}
	switch req.GetSelect().(type) {
	case *pbOrgs.Select_ByName:
		orgInput.Select = &pbOrgs.GetRequest_ByName{
			ByName: name,
		}
	case *pbOrgs.Select_ById:
		orgInput.Select = &pbOrgs.GetRequest_ById{
			ById: id,
		}
	}

	orgResponse, errOrg := orgs.GetSingletonServiceServer(s.db).Get(
		context.Background(),
		connect.NewRequest(orgInput),
	)

	if errOrg != nil {
		return nil, errGet.UpdateMessage(errOrg.Error())
	}
	parse, err := uuid.Parse(orgResponse.Msg.GetOrg().Id)
	if err != nil {
		return nil, errGet.UpdateMessage(err.Error())
	}
	return &parse, nil
}

func (s *ServiceServer) getRecipientSelect(req *pbRecipients.Select) (*pbRecipients.Recipient, *common.ErrWithCode) {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"getting",
		_entityName,
		"",
	)

	recipientInput := &pbRecipients.GetRequest{
		Select: req,
	}

	type magicType string
	var magicKey magicType = "magicValue"
	recipientResponse, errRecipient := GetSingletonServiceServer(s.db).Get(
		context.WithValue(context.Background(), magicKey, magicKey),
		connect.NewRequest(recipientInput),
	)

	if errRecipient != nil {
		return nil, errGet.UpdateMessage(errRecipient.Error())
	}

	return recipientResponse.Msg.GetRecipient(), nil
}
