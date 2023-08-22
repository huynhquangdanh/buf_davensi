package users

import (
	"context"
	"fmt"

	"davensi.com/core/internal/common"

	pbCommon "davensi.com/core/gen/common"
	pbUsers "davensi.com/core/gen/users"
	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) validateCreate(req *pbUsers.CreateRequest) *common.ErrWithCode {
	// Verify that Login is specified
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if req.Login == "" {
		return errCreation.UpdateMessage("login must be specified")
	}

	isUniq, errCode := s.IsUserUniq(&req.Login)
	if !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("type and name have been used")
	}

	return nil
}

func (s *ServiceServer) IsUserUniq(recipientLogin *string) (isUniq bool, errCode pbCommon.ErrorCode) {
	user, err := s.Get(context.Background(), &connect.Request[pbUsers.GetRequest]{
		Msg: &pbUsers.GetRequest{
			Select: &pbUsers.GetRequest_ByLogin{
				ByLogin: *recipientLogin,
			},
		},
	})

	if err == nil || user.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if user.Msg.GetError() == nil && user.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
	}

	return true, user.Msg.GetError().Code
}

func validateQueryUpdate(msg *pbUsers.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_type_name must be specified")
	}
	switch msg.GetSelect().(type) {
	case *pbUsers.UpdateRequest_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errUpdate.UpdateMessage("id must be specified")
		}
	case *pbUsers.UpdateRequest_ByLogin:
		// Verify that Login is specified
		if msg.GetByLogin() == "" {
			return errUpdate.UpdateMessage("login must be specified")
		}
	}

	return nil
}

func validateQueryGet(msg *pbUsers.GetRequest) *common.ErrWithCode {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"get",
		_package,
		"",
	)
	if msg.Select == nil {
		return errGet.UpdateMessage("by_id or by_type_name must be specified")
	}
	switch msg.GetSelect().(type) {
	case *pbUsers.GetRequest_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errGet.UpdateMessage("by_id must be specified")
		}
	case *pbUsers.GetRequest_ByLogin:
		if msg.GetByLogin() == "" {
			return errGet.UpdateMessage("login must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	oldUser *pbUsers.User,
	req *pbUsers.UpdateRequest,
) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)

	if req.Login != nil && req.GetLogin() != oldUser.Login {
		if isUniq, errCode := s.IsUserUniq(req.Login); !isUniq {
			return errUpdate.UpdateCode(errCode).UpdateMessage(
				fmt.Sprintf("login '%s' have been used", req.GetLogin()),
			)
		}
	}

	return nil
}
