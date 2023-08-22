package physiques

import (
	"errors"

	pbCommon "davensi.com/core/gen/common"
	pbPhysiques "davensi.com/core/gen/physiques"
	"davensi.com/core/internal/common"
)

// for Update gRPC
func validateQueryUpdate(msg *pbPhysiques.UpdateRequest) error {
	// Verify that ID is specified
	if msg.GetId() == "" {
		return errors.New("id must be specified")
	}
	return nil
}

func validateQueryGet(msg *pbPhysiques.GetRequest) *common.ErrWithCode {
	errGet := common.CreateErrWithCode(
		pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT,
		"get",
		_package,
		"",
	)
	// Verify that ID is specified
	if msg.GetId() == "" {
		return errGet.UpdateMessage("by_id must be specified")
	}

	return nil
}
