package livelinesses

import (
	"errors"

	pbCommon "davensi.com/core/gen/common"
	pbLiveliness "davensi.com/core/gen/liveliness"
)

func (s *ServiceServer) validateQueryUpdate(msg *pbLiveliness.UpdateRequest) error {
	// Verify that ID is specified
	if msg.GetId() == "" {
		return errors.New("id must be specified")
	}
	return nil
}

func (s *ServiceServer) validateMsgGetOne(msg *pbLiveliness.GetRequest) (err error) {
	err = nil
	if msg.GetId() == "" {
		err = errors.New("id must be specified")
	}
	return
}

func (s *ServiceServer) validateQueryInsert(msg *pbLiveliness.CreateRequest) (err error) {
	err = nil
	if msg.GetIdOwnershipPhotoFileType() == "" || msg.GetLivelinessVideoFileType() == "" || msg.GetTimestampVideoFileType() == "" {
		err = errors.New("type must be specified")
	}
	return
}

func (s *ServiceServer) validateQueryDelete(msg *pbLiveliness.DeleteRequest) error {
	// Verify that ID is specified
	if msg.GetId() == "" {
		return errors.New("id must be specified")
	}
	return nil
}

func (s *ServiceServer) validateUpdateValue(_ *pbLiveliness.Liveliness, _ *pbLiveliness.UpdateRequest) (pbCommon.ErrorCode, error) {
	return pbCommon.ErrorCode_ERROR_CODE_UNSPECIFIED, nil
}
