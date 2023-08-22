package fsproviders

import (
	"context"

	pbCommon "davensi.com/core/gen/common"
	pbFSProviders "davensi.com/core/gen/fsproviders"
	"davensi.com/core/internal/common"

	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) IsFSProviderUniq(fsproviderTypeSymbol *pbFSProviders.TypeName) (isUniq bool, errno pbCommon.ErrorCode) {
	fsprovider, err := s.Get(context.Background(), &connect.Request[pbFSProviders.GetRequest]{
		Msg: &pbFSProviders.GetRequest{
			Select: &pbFSProviders.Select{
				Select: &pbFSProviders.Select_ByTypeName{
					ByTypeName: fsproviderTypeSymbol,
				},
			},
		},
	})

	if err == nil || fsprovider.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, fsprovider.Msg.GetError().Code
	}

	if fsprovider.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, fsprovider.Msg.GetError().Code
	}

	return true, fsprovider.Msg.GetError().Code
}

func ValidateSelect(selectProvider *pbFSProviders.Select, method string) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_package,
		"",
	)
	if selectProvider == nil {
		return errValidate.UpdateMessage("by_id or by_type_name must be specified")
	}
	if selectProvider.Select == nil {
		return errValidate.UpdateMessage("by_id or by_type_name must be specified")
	}
	switch selectProvider.GetSelect().(type) {
	case *pbFSProviders.Select_ById:
		// Verify that ID is specified
		if selectProvider.GetById() == "" {
			return errValidate.UpdateMessage("id must be specified")
		}
	case *pbFSProviders.Select_ByTypeName:
		if selectProvider.GetByTypeName() == nil {
			return errValidate.UpdateMessage("type_name must be specified")
		}
		if selectProvider.GetByTypeName().Type == 0 || selectProvider.GetByTypeName().Name == "" {
			return errValidate.UpdateMessage("type and name must be specified")
		}
	}
	return nil
}

// for Create gRPC
func (s *ServiceServer) validateCreate(msg *pbFSProviders.CreateRequest) *common.ErrWithCode {
	// Verify that Type and Name are specified
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	if msg.Type == 0 || msg.Name == "" {
		return errCreation.UpdateMessage("type and name must be specified")
	}

	if isUniq, errCode := s.IsFSProviderUniq(&pbFSProviders.TypeName{
		Type: msg.GetType(),
		Name: msg.GetName(),
	}); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("name and type have been used")
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	oldFsprovider *pbFSProviders.FSProvider,
	msg *pbFSProviders.UpdateRequest,
) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)

	checkType := oldFsprovider.Type
	checkName := oldFsprovider.Name
	isUpddateTypeName := false

	if msg.Type != nil {
		checkType = msg.GetType()
		isUpddateTypeName = true
	}

	if msg.Name != nil {
		checkName = msg.GetName()
		isUpddateTypeName = true
	}

	if isUpddateTypeName {
		if isUniq, errCode := s.IsFSProviderUniq(&pbFSProviders.TypeName{
			Type: checkType,
			Name: checkName,
		}); !isUniq {
			return errValidate.UpdateCode(errCode).UpdateMessage("name and type have been used")
		}
	}

	return nil
}
