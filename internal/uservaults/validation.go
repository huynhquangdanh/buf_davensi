package uservaults

import (
	"encoding/hex"
	"fmt"
	"strconv"

	pbCommon "davensi.com/core/gen/common"
	pbUserVaults "davensi.com/core/gen/uservaults"
	"davensi.com/core/internal/common"
)

func (s *ServiceServer) validateCreate(req *pbUserVaults.SetRequest,
) (hexDataString string, valueType pbUserVaults.ValueType, errno pbCommon.ErrorCode, err error) {
	key := []byte(encryptKey)
	errnoInvalidArg := pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT

	// Verify that User, Key, Value are specified
	if req.User == nil {
		return "", 0, errno, fmt.Errorf(common.Errors[uint32(errnoInvalidArg)], "setting "+_entityName, "user.User must be specified")
	}
	if req.Key == "" {
		return "", 0, errno, fmt.Errorf(common.Errors[uint32(errnoInvalidArg)], "setting "+_entityName, "Key must be specified")
	}
	if req.Value == nil {
		return "", 0, errno, fmt.Errorf(common.Errors[uint32(errnoInvalidArg)], "setting "+_entityName, "Value must be specified")
	}

	switch req.GetValue().(type) {
	case *pbUserVaults.SetRequest_Bool:
		boolString := strconv.FormatBool(req.GetBool())
		encryptedBoolBytes, err := encrypt([]byte(boolString), key)
		if err != nil {
			return "", 0, pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err
		}
		hexDataString = "\\x" + hex.EncodeToString(encryptedBoolBytes)
		valueType = pbUserVaults.ValueType_VALUE_TYPE_BOOL
	case *pbUserVaults.SetRequest_Integer:
		intString := strconv.Itoa(int(req.GetInteger()))
		encryptedIntBytes, err := encrypt([]byte(intString), key)
		if err != nil {
			return "", 0, pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err
		}
		hexDataString = "\\x" + hex.EncodeToString(encryptedIntBytes)
		valueType = pbUserVaults.ValueType_VALUE_TYPE_INT64
	case *pbUserVaults.SetRequest_Decimal:
		decimalString := req.GetDecimal().Value
		_, err := strconv.ParseFloat(decimalString, 64)
		if err != nil {
			return "", 0, pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err
		}
		encryptedDecimalBytes, err := encrypt([]byte(decimalString), key)
		if err != nil {
			return "", 0, pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err
		}
		hexDataString = "\\x" + hex.EncodeToString(encryptedDecimalBytes)
		valueType = pbUserVaults.ValueType_VALUE_TYPE_DECIMAL
	case *pbUserVaults.SetRequest_String_:
		stringString := req.GetString_()
		encryptedStringBytes, err := encrypt([]byte(stringString), key)
		if err != nil {
			return "", 0, pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err
		}
		hexDataString = "\\x" + hex.EncodeToString(encryptedStringBytes)
		valueType = pbUserVaults.ValueType_VALUE_TYPE_STRING
	case *pbUserVaults.SetRequest_Bytes:
		bytes := req.GetBytes()
		encryptedBytesBytes, err := encrypt(bytes, key)
		if err != nil {
			return "", 0, pbCommon.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err
		}
		hexDataString = "\\x" + hex.EncodeToString(encryptedBytesBytes)
		valueType = pbUserVaults.ValueType_VALUE_TYPE_BYTE
	}

	return hexDataString, valueType, pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}
