package documents

import (
	pbCommon "davensi.com/core/gen/common"
	pbDocuments "davensi.com/core/gen/documents"
	"davensi.com/core/internal/common"
)

// for Get gRPC
func validateQueryGet(msg *pbDocuments.GetRequest) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.GetId() == "" {
		return errValidate.UpdateMessage("id must be specified")
	}

	return nil
}

// for Update gRPC
func validateQueryUpdate(msg *pbDocuments.UpdateRequest) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.GetId() == "" {
		return errValidate.UpdateMessage("id must be specified")
	}

	if msg.File == nil || msg.GetFile() == "" {
		return errValidate.UpdateMessage("file must be specified")
	}

	return nil
}

// for Create gRPC
func validateCreation(msg *pbDocuments.CreateRequest) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.File == "" {
		return errValidate.UpdateMessage("file must be specified")
	}

	if len(msg.Data) == 0 {
		return errValidate.UpdateMessage("document data must be specified")
	}

	return nil
}

func validateSetData(msg *pbDocuments.SetDataRequest) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.Id == "" {
		return errValidate.UpdateMessage("file must be specified")
	}

	if len(msg.Data) == 0 {
		return errValidate.UpdateMessage("document data must be specified")
	}

	return nil
}

func validateUpdateData(msg *pbDocuments.UpdateDataRequest) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.Id == "" {
		return errValidate.UpdateMessage("file must be specified")
	}

	if len(msg.Data) == 0 {
		return errValidate.UpdateMessage("document data must be specified")
	}

	return nil
}

func validateRemoveData(msg *pbDocuments.RemoveDataRequest) *common.ErrWithCode {
	errValidate := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"updating",
		_package,
		"",
	)
	if msg.Id == "" {
		return errValidate.UpdateMessage("file must be specified")
	}

	if msg.Keys == nil || len(msg.Keys.List) == 0 {
		return errValidate.UpdateMessage("document data must be specified")
	}

	return nil
}
