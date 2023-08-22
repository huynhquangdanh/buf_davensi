package userprefs

import (
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	pbUserPrefs "davensi.com/core/gen/userprefs"
	"davensi.com/core/internal/common"
)

func (s *ServiceServer) ValidateCreate(req *pbUserPrefs.SetRequest) (errno pbCommon.ErrorCode, err error) {
	// Verify that User, Key are specified
	if req.User == nil {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf(common.Errors[uint32(errno)], "setting "+_entityName, "user.User must be specified")
	}
	if req.Key == "" {
		errno = pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT
		return errno, fmt.Errorf(common.Errors[uint32(errno)], "setting "+_entityName, "Key must be specified")
	}

	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}
