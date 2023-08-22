package contacts

import (
	pbCommon "davensi.com/core/gen/common"
	pbContacts "davensi.com/core/gen/contacts"
	"davensi.com/core/internal/common"
)

// for Update gRPC
func validateQueryUpdate(msg *pbContacts.UpdateRequest) *common.ErrWithCode {
	// Verify that ID is specified
	errUpdate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.GetId() == "" {
		return errUpdate.UpdateMessage("id must be specified")
	}

	return nil
}

// for Create gRPC
func (s *ServiceServer) ValidateCreate(msg *pbContacts.CreateRequest) *common.ErrWithCode {
	// Verify that Name are specified
	errCreation := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"creating",
		_entityName,
		"",
	)
	if msg.Value == "" {
		return errCreation.UpdateMessage("value must be specified")
	}

	return nil
}

func validateQueryGet(msg *pbContacts.GetRequest) *common.ErrWithCode {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"get",
		_package,
		"",
	)
	// Verify that ID is specified
	if msg.GetId() == "" {
		return errGet.UpdateMessage("id must be specified")
	}

	return nil
}
