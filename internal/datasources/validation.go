package datasources

import (
	"context"
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbDataSources "davensi.com/core/gen/datasources"
	pbFsproviders "davensi.com/core/gen/fsproviders"
	"davensi.com/core/internal/common"
	"davensi.com/core/internal/fsproviders"
	"github.com/bufbuild/connect-go"
)

func (s *ServiceServer) IsDataSourceUniq(selectByTypeName *pbDataSources.Select_ByTypeName) (isUniq bool, errCode pbCommon.ErrorCode) {
	dataSource, err := s.Get(context.Background(), &connect.Request[pbDataSources.GetRequest]{
		Msg: &pbDataSources.GetRequest{
			Select: &pbDataSources.Select{
				Select: selectByTypeName,
			},
		},
	})

	if err == nil || dataSource.Msg.GetError().Code == pbCommon.ErrorCode_ERROR_CODE_MULTIPLE_VALUES_FOUND {
		return false, pbCommon.ErrorCode_ERROR_CODE_DUPLICATE_KEY
	}

	if dataSource.Msg.GetError() != nil && dataSource.Msg.GetError().Code != pbCommon.ErrorCode_ERROR_CODE_NOT_FOUND {
		return false, dataSource.Msg.GetError().Code
	}

	return true, dataSource.Msg.GetError().Code
}

func (s *ServiceServer) validateCreate(msg *pbDataSources.CreateRequest) *common.ErrWithCode {
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

	if msg.Provider == nil {
		return errCreation.UpdateMessage("provider must be specified")
	}

	if errSelectProvider := fsproviders.ValidateSelect(msg.Provider, "creating"); errSelectProvider != nil {
		return errSelectProvider
	}

	datasourceRl := s.GetRelationship(msg.GetProvider())

	if datasourceRl.fsprovider == nil {
		return errCreation.UpdateMessage("provider does not exist")
	}

	msg.Provider = &pbFsproviders.Select{
		Select: &pbFsproviders.Select_ById{
			ById: datasourceRl.fsprovider.Id,
		},
	}

	if isUniq, errCode := s.IsDataSourceUniq(&pbDataSources.Select_ByTypeName{
		ByTypeName: &pbDataSources.TypeName{
			Type: msg.GetType(),
			Name: msg.GetName(),
		},
	}); !isUniq {
		return errCreation.UpdateCode(errCode).UpdateMessage("type and name have been used")
	}

	return nil
}

func ValidateSelect(msg *pbDataSources.Select, method string) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		method,
		_package,
		"",
	)
	if msg == nil {
		return errUpdate.UpdateMessage("by_id or by_type_name must be specified")
	}
	if msg.Select == nil {
		return errUpdate.UpdateMessage("by_id or by_type_name must be specified")
	}
	switch msg.GetSelect().(type) {
	case *pbDataSources.Select_ByTypeName:
		if msg.GetByTypeName() == nil {
			return errUpdate.UpdateMessage("by_type_name must be specified")
		}
		if msg.GetByTypeName().Type == 0 || msg.GetByTypeName().Name == "" {
			return errUpdate.UpdateMessage("type and name must be specified")
		}
	case *pbDataSources.Select_ById:
		// Verify that ID is specified
		if msg.GetById() == "" {
			return errUpdate.UpdateMessage("by_id must be specified")
		}
	}

	return nil
}

func (s *ServiceServer) validateQueryUpdate(msg *pbDataSources.UpdateRequest) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_entityName,
		"",
	)
	if errValidateSelect := ValidateSelect(msg.Select, "updating"); errValidateSelect != nil {
		return errValidateSelect
	}

	datasourceRl := s.GetRelationship(msg.GetProvider())

	if msg.Provider != nil && datasourceRl.fsprovider == nil {
		return errUpdate.UpdateMessage("provider does not exist")
	}

	if datasourceRl.fsprovider != nil {
		msg.Provider = &pbFsproviders.Select{
			Select: &pbFsproviders.Select_ById{
				ById: datasourceRl.fsprovider.Id,
			},
		}
	}

	return nil
}

func (s *ServiceServer) validateUpdateValue(
	dataSource *pbDataSources.DataSource,
	msg *pbDataSources.UpdateRequest,
) *common.ErrWithCode {
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	checkType := dataSource.Type
	checkName := dataSource.Name
	isUpdateTypeName := false

	if msg.Type != nil && msg.GetType() != dataSource.GetType() {
		checkType = msg.GetType()
		isUpdateTypeName = true
	}

	if msg.Name != nil && msg.GetName() != dataSource.GetName() {
		checkName = msg.GetName()
		isUpdateTypeName = true
	}

	if isUpdateTypeName {
		if isUniq, errCode := s.IsDataSourceUniq(&pbDataSources.Select_ByTypeName{
			ByTypeName: &pbDataSources.TypeName{
				Type: checkType,
				Name: checkName,
			},
		}); !isUniq {
			return errUpdate.UpdateCode(errCode).UpdateMessage(
				fmt.Sprintf("type/name '%s/%s' have been used", checkType, checkName),
			)
		}
	}

	msg.Select = &pbDataSources.Select{
		Select: &pbDataSources.Select_ById{
			ById: dataSource.Id,
		},
	}

	return nil
}
