package authgroups

import (
	"context"
	"fmt"

	pbAuthGroups "davensi.com/core/gen/authgroups"
	pbCommon "davensi.com/core/gen/common"
	"davensi.com/core/internal/common"

	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) IsAuthGroupUniq(authGroupName *pbAuthGroups.Select_ByName) (isUniq bool, errno pbCommon.ErrorCode) {
	authGroup, err := s.Get(context.Background(), &connect.Request[pbAuthGroups.GetRequest]{
		Msg: &pbAuthGroups.GetRequest{
			Select: &pbAuthGroups.Select{
				Select: &pbAuthGroups.Select_ByName{
					ByName: authGroupName.ByName,
				},
			},
		},
	})

	if err == nil || authGroup.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND
	}

	if authGroup.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, authGroup.Msg.GetError().Code
	}

	return true, authGroup.Msg.GetError().Code
}

// for Create gRPC
func (s *ServiceServer) validateCreate(msg *pbAuthGroups.CreateRequest) (errno pbCommon.ErrorCode, err error) {
	// Verify that Type and Symbol are specified
	if msg.Name == "" {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf("creating '%s' type and symbol must be specified", _entityName)
	}

	if isUniq, errno := s.IsAuthGroupUniq(&pbAuthGroups.Select_ByName{
		ByName: msg.GetName(),
	}); !isUniq {
		return errno, fmt.Errorf("create %s name = '%s' Name", _entityName, msg.GetName())
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}

// for Update gRPC
func (s *ServiceServer) validateQueryUpdate(msg *pbAuthGroups.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_type_name must be specified")
	}

	switch msg.GetSelect().Select.(type) {
	case *pbAuthGroups.Select_ById:
		// Verify that ID is specified
		if msg.GetSelect().GetById() == "" {
			return errUpdate.UpdateMessage("id must be specified")
		}
	case *pbAuthGroups.Select_ByName:
		// Verify that name is specified
		if msg.GetSelect().GetByName() == "" {
			return errUpdate.UpdateMessage("name must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	oldAuthGroup *pbAuthGroups.AuthGroup,
	msg *pbAuthGroups.UpdateRequest,
) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	checkName := oldAuthGroup.Name
	isUpdateName := false

	if msg.Name != nil {
		checkName = msg.GetName()
		isUpdateName = true
	}

	if isUpdateName {
		if isUniq, errCode := s.IsAuthGroupUniq(&pbAuthGroups.Select_ByName{
			ByName: checkName,
		}); !isUniq && checkName != oldAuthGroup.Name {
			return errUpdate.UpdateCode(errCode).UpdateMessage(
				fmt.Sprintf("name '%s' have been used", checkName),
			)
		}
	}

	return nil
}

func validateQueryGet(msg *pbAuthGroups.GetRequest) *common.ErrWithCode {
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
	case *pbAuthGroups.Select_ById:
		// Verify that ID is specified
		if msg.GetSelect().GetById() == "" {
			return errGet.UpdateMessage("by_id must be specified")
		}
	case *pbAuthGroups.Select_ByName:
		if msg.GetSelect().GetByName() == "" {
			return errGet.UpdateMessage("name must be specified")
		}
	}

	return nil
}
