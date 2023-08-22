package common

import (
	"fmt"

	pbCommon "davensi.com/core/gen/common"
	"github.com/rs/zerolog/log"
)

var Errors = map[uint32]string{
	0: "unspecified error",
	1: "database error while %s %s with '%s'",
	2: "database field mapping error while %s %s '%s'",
	3: "multiple %s found for '%s'",
	4: "cannot %s %s with '%s' as a record already exists with the same %s",
	5: "no %s found for '%s'",
	6: "streaming error while %s %s%s",
	7: "invalid argument while %s: %s",
}

func StreamError(entityName string, errCode pbCommon.ErrorCode, err error, handleErr func(errStream *pbCommon.Error) error) error {
	_err := fmt.Errorf(Errors[uint32(errCode.Number())], "listing", entityName, "<Selection>")
	log.Error().Err(err).Msg(_err.Error())

	if errSend := handleErr(&pbCommon.Error{
		Code:    errCode,
		Package: entityName,
		Text:    _err.Error() + "(" + err.Error() + ")",
	}); errSend != nil {
		_errnoSend := pbCommon.ErrorCode_ERROR_CODE_STREAMING_ERROR
		_errSend := fmt.Errorf(Errors[uint32(_errnoSend.Number())], "listing", entityName, "<Selection>")
		log.Error().Err(errSend).Msg(_errSend.Error())
		_err = _errSend
	}
	return _err
}

type ErrWithCode struct {
	method      string
	packageName string
	Code        pbCommon.ErrorCode
	Err         error
}

// Must be used before update message
func (errWithCode *ErrWithCode) UpdateCode(errCode pbCommon.ErrorCode) *ErrWithCode {
	errWithCode.Code = errCode
	return errWithCode
}

func (errWithCode *ErrWithCode) UpdateMessage(message string) *ErrWithCode {
	errWithCode.Err = fmt.Errorf(
		Errors[uint32(errWithCode.Code.Number())],
		"%s '%s' error: '%s'",
		errWithCode.method, errWithCode.packageName, message,
	)
	return errWithCode
}

func CreateErrWithCode(errCode pbCommon.ErrorCode, method, packageName, message string) *ErrWithCode {
	return &ErrWithCode{
		method:      method,
		packageName: packageName,
		Code:        errCode,
		Err:         fmt.Errorf("%s '%s' error: '%s'", method, packageName, message),
	}
}
