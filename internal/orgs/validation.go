package orgs

import (
	"context"

	"davensi.com/core/internal/common"

	pbCommon "davensi.com/core/gen/common"
	pbOrgs "davensi.com/core/gen/orgs"
	"github.com/bufbuild/connect-go"
)

// for Create gRPC
func (s *ServiceServer) validateCreate(req *pbOrgs.CreateRequest) *common.ErrWithCode {
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if req.Name == "" {
		return errCreation.UpdateMessage("name must be specified")
	}

	isUniq, errCode := s.IsOrgUniq(&req.Name)
	if !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("name must be specified")
	}

	return nil
}

func (s *ServiceServer) IsOrgUniq(orgName *string) (isUniq bool, errCode pbCommon.ErrorCode) {
	org, err := s.Get(context.Background(), &connect.Request[pbOrgs.GetRequest]{
		Msg: &pbOrgs.GetRequest{
			Select: &pbOrgs.GetRequest_ByName{
				ByName: *orgName,
			},
		},
	})

	if err == nil && org.Msg.GetOrg() != nil {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if err == nil || org.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if org.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, org.Msg.GetError().Code
	}

	return true, org.Msg.GetError().Code
}

// for Update gRPC
func validateQueryUpdate(msg *pbOrgs.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updateing",
		_entityName,
		"",
	)
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_name must be specified")
	}

	switch msg.GetSelect().(type) {
	case *pbOrgs.UpdateRequest_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errUpdate.UpdateMessage("id must be specified")
		}
	case *pbOrgs.UpdateRequest_ByName:
		// Verify that name is specified
		if msg.GetByName() == "" {
			return errUpdate.UpdateMessage("name must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	org *pbOrgs.Org,
	msg *pbOrgs.UpdateRequest,
) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updateing",
		_entityName,
		"",
	)
	if msg.Name != nil && msg.GetName() != org.GetName() {
		isUniq, errCode := s.IsOrgUniq(msg.Name)
		if !isUniq {
			return errUpdate.UpdateCode(errCode).UpdateMessage("name have been used")
		}
	}

	return nil
}
